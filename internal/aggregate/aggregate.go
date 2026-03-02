package aggregate

import (
	"sort"
	"strings"

	"ispnag/internal/mood"
	"ispnag/internal/parse"
)

type DeviceStats struct {
	Device         string
	Location       string
	WarningCount   int
	AvgPacketLoss  float64
	MaxPacketLoss  float64
	AvgRTA         float64
	MaxRTA         float64
	AnnoyanceScore int
}

type LocationStats struct {
	Location     string
	WarningCount int
	Mood         string
}

type Digest struct {
	Devices        []DeviceStats
	Locations      []LocationStats
	IgnoredDevices int
	NetworkMood    string
}

func Build(events []parse.Event) Digest {
	criticalDevices := make(map[string]struct{})
	for _, e := range events {
		if e.State == "CRITICAL" {
			criticalDevices[e.Device] = struct{}{}
		}
	}

	type acc struct {
		warnings int
		sumLoss  float64
		sumRTA   float64
		maxLoss  float64
		maxRTA   float64
	}
	byDevice := make(map[string]*acc)

	for _, e := range events {
		if e.State != "WARNING" {
			continue
		}
		if _, blocked := criticalDevices[e.Device]; blocked {
			continue
		}
		a, ok := byDevice[e.Device]
		if !ok {
			a = &acc{}
			byDevice[e.Device] = a
		}
		a.warnings++
		a.sumLoss += e.PacketLoss
		a.sumRTA += e.RTA
		if e.PacketLoss > a.maxLoss {
			a.maxLoss = e.PacketLoss
		}
		if e.RTA > a.maxRTA {
			a.maxRTA = e.RTA
		}
	}

	devices := make([]DeviceStats, 0, len(byDevice))
	locationWarnings := make(map[string]int)

	for device, a := range byDevice {
		if a.warnings == 0 {
			continue
		}
		location := parseLocation(device)
		avgLoss := a.sumLoss / float64(a.warnings)
		avgRTA := a.sumRTA / float64(a.warnings)
		ann := annoyance(a.warnings, a.maxLoss, a.maxRTA)

		d := DeviceStats{
			Device:         device,
			Location:       location,
			WarningCount:   a.warnings,
			AvgPacketLoss:  avgLoss,
			MaxPacketLoss:  a.maxLoss,
			AvgRTA:         avgRTA,
			MaxRTA:         a.maxRTA,
			AnnoyanceScore: ann,
		}
		devices = append(devices, d)
		locationWarnings[location] += a.warnings
	}

	sort.Slice(devices, func(i, j int) bool {
		if devices[i].AnnoyanceScore == devices[j].AnnoyanceScore {
			return devices[i].WarningCount > devices[j].WarningCount
		}
		return devices[i].AnnoyanceScore > devices[j].AnnoyanceScore
	})

	locations := make([]LocationStats, 0, len(locationWarnings))
	for location, warnings := range locationWarnings {
		locations = append(locations, LocationStats{
			Location:     location,
			WarningCount: warnings,
			Mood:         mood.FromAnnoyance(warnings),
		})
	}

	sort.Slice(locations, func(i, j int) bool {
		if locations[i].WarningCount == locations[j].WarningCount {
			return locations[i].Location < locations[j].Location
		}
		return locations[i].WarningCount > locations[j].WarningCount
	})

	networkMood := networkMood(devices)

	return Digest{
		Devices:        devices,
		Locations:      locations,
		IgnoredDevices: len(criticalDevices),
		NetworkMood:    networkMood,
	}
}

func parseLocation(device string) string {
	parts := strings.SplitN(device, "___", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
		return "unknown"
	}
	return parts[0]
}

func annoyance(warnings int, maxPacketLoss, maxRTA float64) int {
	return warnings + int(maxPacketLoss*2) + int(maxRTA/20)
}

func networkMood(devices []DeviceStats) string {
	if len(devices) == 0 {
		return mood.FromAnnoyance(0)
	}
	topN := 10
	if len(devices) < topN {
		topN = len(devices)
	}
	sum := 0
	for i := 0; i < topN; i++ {
		sum += devices[i].AnnoyanceScore
	}
	avg := sum / topN
	return mood.FromAnnoyance(avg)
}
