# ISPnag

It nags before users call.

ISPnag is a small batch CLI that reads Nagios `showlog.cgi` output, extracts `WARNING` service alerts, excludes any device that escalated to `CRITICAL`, and sends a daily plain-text digest to Slack or Microsoft Teams webhook.

If the configured lookback crosses midnight, ISPnag automatically also fetches `?archive=1` and merges it with the current log so a full 24h window still works on installations with daily log rotation.

## Environment variables

- `NAGIOS_URL` (required)
- `WEBHOOK_URL` (optional; if empty, ISPnag runs in dry-run mode and only prints report)
- `WEBHOOK_TYPE` (`slack` or `teams`, default `slack`)
- `LOOKBACK_HOURS` (default `24`)
- `HTTP_TIMEOUT_SECONDS` (default `30`)
- `REPORT_TOP_DEVICES` (default `10`)
- `TZ` (recommended `Europe/Prague` for correct local-time lookback behavior)
- `DEBUG` (`1|true|yes|on` enables debug logs)

## Run locally

```bash
go run ./cmd/ispnag
```

## Run with Docker

```bash
docker build -t ispnag .
docker run --rm \
  -e NAGIOS_URL="https://example.com/showlog.cgi" \
  -e WEBHOOK_URL="https://hooks.slack.com/services/..." \
  -e WEBHOOK_TYPE="slack" \
  ispnag
```

## Run with Docker Compose

```bash
cp .env.example .env
# fill in NAGIOS_URL and optional webhook settings
docker compose run --rm ispnag
```

## GitHub container image (GHCR)

Workflow is in `.github/workflows/container.yml` and does:
- `go test ./...`
- multi-arch build (`linux/amd64`, `linux/arm64`)
- push to `ghcr.io/<owner>/<repo>`

Triggers:
- push to `main`
- tag push `v*`
- manual `workflow_dispatch`
