package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"ispnag/internal/aggregate"
	"ispnag/internal/fetch"
	"ispnag/internal/mood"
	"ispnag/internal/parse"
	"ispnag/internal/report"
	"ispnag/internal/webhook"
)

type config struct {
	NagiosURL          string
	WebhookURL         string
	WebhookType        string
	LookbackHours      int
	HTTPTimeoutSeconds int
	ReportTopDevices   int
	Debug              bool
}

func main() {
	log.SetFlags(0)

	cfg, err := loadConfig()
	if err != nil {
		fatal(err)
	}

	ctx := context.Background()
	timeout := time.Duration(cfg.HTTPTimeoutSeconds) * time.Second
	now := time.Now()

	body, err := fetch.NagiosLog(ctx, cfg.NagiosURL, cfg.LookbackHours, timeout)
	if err != nil {
		fatal(err)
	}
	needPreviousArchive := lookbackCrossesMidnight(now, cfg.LookbackHours)
	if needPreviousArchive {
		prevBody, prevErr := fetch.NagiosLogArchive(ctx, cfg.NagiosURL, cfg.LookbackHours, timeout, 1)
		if prevErr != nil {
			fatal(fmt.Errorf("lookback requires previous archive (?archive=1) but fetch failed: %w", prevErr))
		}
		body = body + "\n" + prevBody
		if cfg.Debug {
			log.Printf("DEBUG: appended previous archive (?archive=1), extra bytes=%d", len(prevBody))
		}
	}
	rawServiceAlerts := strings.Count(body, "SERVICE ALERT")
	rawWarningStates := strings.Count(body, ";WARNING;")
	rawCriticalStates := strings.Count(body, ";CRITICAL;")
	if cfg.Debug {
		log.Printf(
			"DEBUG: fetched response bytes=%d, raw_service_alert_lines=%d, raw_warning_states=%d, raw_critical_states=%d, used_archive1=%t",
			len(body), rawServiceAlerts, rawWarningStates, rawCriticalStates, needPreviousArchive,
		)
		logPreviewServiceAlerts(body, 3)
	}
	if rawServiceAlerts == 0 {
		log.Printf("WARN: no 'SERVICE ALERT' lines in response; endpoint may require auth/session or different showlog filter")
	}

	events, err := parse.Parse(body, now, time.Duration(cfg.LookbackHours)*time.Hour)
	if err != nil {
		fatal(err)
	}
	if rawServiceAlerts > 0 && len(events) == 0 {
		log.Printf("WARN: 'SERVICE ALERT' exists but no WARNING/CRITICAL events matched lookback=%dh", cfg.LookbackHours)
	}

	digest := aggregate.Build(events)
	personality := mood.RandomPersonalityLine(digest.NetworkMood, time.Now().UnixNano())
	reportText := report.Render(report.Input{
		Digest:          digest,
		PersonalityLine: personality,
		MaxDevices:      cfg.ReportTopDevices,
	})

	if cfg.Debug {
		log.Printf(
			"DEBUG: parsed events=%d, devices=%d, locations=%d, ignored_incidents=%d",
			len(events), len(digest.Devices), len(digest.Locations), digest.IgnoredDevices,
		)
		if oldest, newest, count, ok := timedEventRange(events); ok {
			log.Printf("DEBUG: parsed_timestamps=%d oldest=%s newest=%s", count, oldest.Format(time.RFC3339), newest.Format(time.RFC3339))
		} else {
			log.Printf("DEBUG: parsed_timestamps=0 (no timestamped events)")
		}
	}

	fmt.Println(reportText)

	if cfg.WebhookURL == "" {
		log.Printf("INFO: WEBHOOK_URL is not set, skipping webhook send (dry-run mode)")
		return
	}

	wh := webhook.Client{
		URL:     cfg.WebhookURL,
		Type:    webhook.Type(cfg.WebhookType),
		Timeout: timeout,
	}
	if err := wh.Send(ctx, reportText); err != nil {
		fatal(err)
	}

	log.Printf("ISPnag digest sent successfully (%d devices, %d ignored incidents)", len(digest.Devices), digest.IgnoredDevices)
}

func loadConfig() (config, error) {
	cfg := config{
		NagiosURL:          strings.TrimSpace(os.Getenv("NAGIOS_URL")),
		WebhookURL:         strings.TrimSpace(os.Getenv("WEBHOOK_URL")),
		WebhookType:        defaultString(strings.TrimSpace(os.Getenv("WEBHOOK_TYPE")), "slack"),
		LookbackHours:      defaultInt(strings.TrimSpace(os.Getenv("LOOKBACK_HOURS")), 24),
		HTTPTimeoutSeconds: defaultInt(strings.TrimSpace(os.Getenv("HTTP_TIMEOUT_SECONDS")), 30),
		ReportTopDevices:   defaultInt(strings.TrimSpace(os.Getenv("REPORT_TOP_DEVICES")), 10),
		Debug:              defaultBool(strings.TrimSpace(os.Getenv("DEBUG"))),
	}

	var errs []error
	if cfg.NagiosURL == "" {
		errs = append(errs, errors.New("NAGIOS_URL is required"))
	}
	if cfg.WebhookURL != "" && cfg.WebhookType != "slack" && cfg.WebhookType != "teams" {
		errs = append(errs, errors.New("WEBHOOK_TYPE must be 'slack' or 'teams'"))
	}
	if cfg.LookbackHours <= 0 {
		errs = append(errs, errors.New("LOOKBACK_HOURS must be > 0"))
	}
	if cfg.HTTPTimeoutSeconds <= 0 {
		errs = append(errs, errors.New("HTTP_TIMEOUT_SECONDS must be > 0"))
	}
	if cfg.ReportTopDevices <= 0 {
		errs = append(errs, errors.New("REPORT_TOP_DEVICES must be > 0"))
	}
	if len(errs) > 0 {
		return config{}, errors.Join(errs...)
	}

	return cfg, nil
}

func defaultString(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return strings.ToLower(v)
}

func defaultInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

func defaultBool(raw string) bool {
	switch strings.ToLower(raw) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func lookbackCrossesMidnight(now time.Time, lookbackHours int) bool {
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	elapsedToday := now.Sub(midnight)
	return elapsedToday < (time.Duration(lookbackHours) * time.Hour)
}

func logPreviewServiceAlerts(body string, maxLines int) {
	count := 0
	for _, line := range strings.Split(body, "\n") {
		if !strings.Contains(line, "SERVICE ALERT") {
			continue
		}
		count++
		log.Printf("DEBUG: sample_service_alert_%d=%s", count, strings.TrimSpace(line))
		if count >= maxLines {
			return
		}
	}
}

func timedEventRange(events []parse.Event) (time.Time, time.Time, int, bool) {
	var oldest time.Time
	var newest time.Time
	count := 0
	for _, e := range events {
		if !e.HasTime {
			continue
		}
		if count == 0 || e.At.Before(oldest) {
			oldest = e.At
		}
		if count == 0 || e.At.After(newest) {
			newest = e.At
		}
		count++
	}
	return oldest, newest, count, count > 0
}

func fatal(err error) {
	log.Printf("ERROR: %v", err)
	os.Exit(1)
}
