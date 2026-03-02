package report

import (
	"fmt"
	"math"
	"strings"

	"ispnag/internal/aggregate"
)

type Input struct {
	Digest          aggregate.Digest
	PersonalityLine string
	MaxDevices      int
}

const (
	defaultMaxDevices = 10
	maxDeviceLen      = 36
)

func Render(in Input) string {
	var b strings.Builder

	b.WriteString("👀 ISPnag — předpověď průšvihů\n\n")
	b.WriteString(fmt.Sprintf("Dnešní předpověď: %s %s\n", weatherEmojiForMood(in.Digest.NetworkMood), weatherTextForMood(in.Digest.NetworkMood)))
	if in.PersonalityLine != "" {
		b.WriteString("\n")
		b.WriteString(in.PersonalityLine)
		b.WriteString("\n")
	}

	limit := in.MaxDevices
	if limit <= 0 {
		limit = defaultMaxDevices
	}
	b.WriteString(fmt.Sprintf("\n🌦️ Kde dnes může sprchnout (top %d)\n", limit))
	if len(in.Digest.Devices) == 0 {
		b.WriteString("☀️ Všechno zatím vypadá klidně.\n")
	} else {
		if len(in.Digest.Devices) < limit {
			limit = len(in.Digest.Devices)
		}
		for i := 0; i < limit; i++ {
			d := in.Digest.Devices[i]
			b.WriteString(fmt.Sprintf(
				"%s %s\n",
				weatherEmojiForAnnoyance(d.AnnoyanceScore),
				shorten(d.Device, maxDeviceLen),
			))
			b.WriteString(fmt.Sprintf(
				"_loss %d%% | RTA %d ms | %dx_\n",
				int(math.Round(d.MaxPacketLoss)),
				int(math.Round(d.MaxRTA)),
				d.WarningCount,
			))
		}
	}

	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Ignorováno: %d zařízení eskalovalo na incident.\n", in.Digest.IgnoredDevices))

	return strings.TrimRight(b.String(), "\n")
}

func shorten(s string, maxLen int) string {
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return "…"
	}
	return string(r[:maxLen-1]) + "…"
}

func weatherEmojiForAnnoyance(annoyance int) string {
	switch {
	case annoyance <= 20:
		return "☀️"
	case annoyance <= 50:
		return "⛅"
	case annoyance <= 80:
		return "🌥️"
	case annoyance <= 120:
		return "⛈️"
	default:
		return "🌪️"
	}
}

func weatherEmojiForMood(mood string) string {
	switch moodText(mood) {
	case "klidná":
		return "☀️"
	case "lehce nervózní":
		return "⛅"
	case "mrzutá":
		return "🌥️"
	case "nestabilní":
		return "🌧️"
	default:
		return "⛈️"
	}
}

func weatherTextForMood(mood string) string {
	switch moodText(mood) {
	case "klidná":
		return "jasno"
	case "lehce nervózní":
		return "polojasno"
	case "mrzutá":
		return "oblačno"
	case "nestabilní":
		return "přeháňky"
	default:
		return "bouřky"
	}
}

func moodText(mood string) string {
	parts := strings.SplitN(strings.TrimSpace(mood), " ", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return mood
}
