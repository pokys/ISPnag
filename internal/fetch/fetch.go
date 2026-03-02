package fetch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func NagiosLog(ctx context.Context, nagiosURL string, lookbackHours int, timeout time.Duration) (string, error) {
	return nagiosLog(ctx, nagiosURL, lookbackHours, timeout, nil)
}

func NagiosLogArchive(ctx context.Context, nagiosURL string, lookbackHours int, timeout time.Duration, archive int) (string, error) {
	return nagiosLog(ctx, nagiosURL, lookbackHours, timeout, &archive)
}

func nagiosLog(ctx context.Context, nagiosURL string, lookbackHours int, timeout time.Duration, archive *int) (string, error) {
	u, err := url.Parse(nagiosURL)
	if err != nil {
		return "", fmt.Errorf("invalid NAGIOS_URL: %w", err)
	}

	// Preserve caller URL and add lookback hint for common showlog.cgi deployments.
	q := u.Query()
	if q.Get("hours") == "" {
		q.Set("hours", fmt.Sprintf("%d", lookbackHours))
	}
	if archive != nil {
		q.Set("archive", strconv.Itoa(*archive))
	}
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("building request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch nagios log: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("nagios endpoint returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read nagios response: %w", err)
	}

	return string(body), nil
}
