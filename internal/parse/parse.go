package parse

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	serviceAlertPattern = regexp.MustCompile(`SERVICE ALERT:\s*([^;]+);([^;]+);([^;]+);([^;]+);([^;]+);(.*)$`)
	packetLossPattern  = regexp.MustCompile(`Packet loss\s*=\s*([0-9]+(?:\.[0-9]+)?)%`)
	rtaPattern         = regexp.MustCompile(`RTA\s*=\s*([0-9]+(?:\.[0-9]+)?)\s*ms`)
	epochPrefixPattern = regexp.MustCompile(`^\[(\d{10})\]\s*`)
	datetimePattern    = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}[ T]\d{2}:\d{2}:\d{2})\s+`)
	bracketedDateTime  = regexp.MustCompile(`\[(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\]`)
)

// Event is a parsed Nagios service alert line.
type Event struct {
	Device     string
	Service    string
	State      string
	StateType  string
	Attempt    int
	Message    string
	PacketLoss float64
	RTA        float64
	At         time.Time
	HasTime    bool
}

// Parse extracts service alert events from raw text. It intentionally ignores HTML structure.
func Parse(body string, now time.Time, lookback time.Duration) ([]Event, error) {
	cutoff := now.Add(-lookback)
	lines := strings.Split(body, "\n")
	events := make([]Event, 0, len(lines)/8)

	for _, rawLine := range lines {
		if !strings.Contains(rawLine, "SERVICE ALERT") {
			continue
		}

		line := strings.TrimSpace(html.UnescapeString(htmlToText(rawLine)))
		if line == "" {
			continue
		}

		eventTime, hasTime, content := extractTimestamp(line)
		if hasTime && eventTime.Before(cutoff) {
			continue
		}

		loc := serviceAlertPattern.FindStringSubmatch(content)
		if len(loc) != 7 {
			continue
		}

		state := strings.TrimSpace(loc[3])
		if state != "WARNING" && state != "CRITICAL" {
			continue
		}

		attempt, err := strconv.Atoi(strings.TrimSpace(loc[5]))
		if err != nil {
			attempt = 0
		}

		event := Event{
			Device:    strings.TrimSpace(loc[1]),
			Service:   strings.TrimSpace(loc[2]),
			State:     state,
			StateType: strings.TrimSpace(loc[4]),
			Attempt:   attempt,
			Message:   strings.TrimSpace(loc[6]),
			At:        eventTime,
			HasTime:   hasTime,
		}

		if state == "WARNING" {
			event.PacketLoss = extractFloat(packetLossPattern, event.Message)
			event.RTA = extractFloat(rtaPattern, event.Message)
		}

		events = append(events, event)
	}

	if len(events) == 0 {
		return nil, nil
	}
	return events, nil
}

// htmlToText strips lightweight HTML tags from a line.
func htmlToText(s string) string {
	inTag := false
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

func extractTimestamp(line string) (time.Time, bool, string) {
	if m := epochPrefixPattern.FindStringSubmatch(line); len(m) == 2 {
		epoch, err := strconv.ParseInt(m[1], 10, 64)
		if err == nil {
			return time.Unix(epoch, 0), true, strings.TrimSpace(strings.TrimPrefix(line, m[0]))
		}
	}

	if m := datetimePattern.FindStringSubmatch(line); len(m) == 2 {
		for _, layout := range []string{"2006-01-02 15:04:05", "2006-01-02T15:04:05"} {
			if t, err := time.Parse(layout, m[1]); err == nil {
				return t, true, strings.TrimSpace(strings.TrimPrefix(line, m[0]))
			}
		}
	}

	if m := bracketedDateTime.FindStringSubmatch(line); len(m) == 2 {
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", m[1], time.Local); err == nil {
			return t, true, line
		}
	}

	return time.Time{}, false, line
}

func extractFloat(re *regexp.Regexp, src string) float64 {
	m := re.FindStringSubmatch(src)
	if len(m) != 2 {
		return 0
	}
	f, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0
	}
	return f
}

// DebugString renders an event as one line for troubleshooting.
func (e Event) DebugString() string {
	return fmt.Sprintf("%s state=%s loss=%.2f rta=%.2f", e.Device, e.State, e.PacketLoss, e.RTA)
}
