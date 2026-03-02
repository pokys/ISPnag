package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Type string

const (
	TypeSlack Type = "slack"
	TypeTeams Type = "teams"
)

type Client struct {
	URL     string
	Type    Type
	Timeout time.Duration
}

func (c Client) Send(ctx context.Context, report string) error {
	if c.URL == "" {
		return fmt.Errorf("WEBHOOK_URL is required")
	}
	t := Type(strings.ToLower(string(c.Type)))
	if t != TypeSlack && t != TypeTeams {
		return fmt.Errorf("WEBHOOK_TYPE must be 'slack' or 'teams'")
	}
	if c.Timeout <= 0 {
		c.Timeout = 30 * time.Second
	}

	var payload []byte
	var err error
	if t == TypeTeams {
		payload, err = teamsAdaptiveCardPayload(report)
		if err != nil {
			return fmt.Errorf("marshal teams adaptive card payload: %w", err)
		}
	} else {
		payload, err = json.Marshal(map[string]string{"text": report})
		if err != nil {
			return fmt.Errorf("marshal webhook payload: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: c.Timeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("webhook returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func teamsAdaptiveCardPayload(report string) ([]byte, error) {
	lines := strings.Split(report, "\n")
	body := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			body = append(body, map[string]any{
				"type": "TextBlock",
				"text": " ",
				"wrap": true,
			})
			continue
		}
		body = append(body, map[string]any{
			"type": "TextBlock",
			"text": line,
			"wrap": true,
		})
	}

	message := map[string]any{
		"type": "message",
		"attachments": []map[string]any{
			{
				"contentType": "application/vnd.microsoft.card.adaptive",
				"content": map[string]any{
					"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
					"type":    "AdaptiveCard",
					"version": "1.4",
					"body":    body,
				},
			},
		},
	}
	return json.Marshal(message)
}
