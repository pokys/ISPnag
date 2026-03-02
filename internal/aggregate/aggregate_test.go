package aggregate

import (
	"testing"

	"ispnag/internal/parse"
)

func TestBuildExcludesCriticalDevices(t *testing.T) {
	events := []parse.Event{
		{Device: "LocA___dev1", State: "WARNING", PacketLoss: 10, RTA: 40},
		{Device: "LocA___dev1", State: "CRITICAL"},
		{Device: "LocB___dev2", State: "WARNING", PacketLoss: 5, RTA: 20},
		{Device: "dev3", State: "WARNING", PacketLoss: 1, RTA: 10},
	}

	d := Build(events)
	if d.IgnoredDevices != 1 {
		t.Fatalf("expected 1 ignored device, got %d", d.IgnoredDevices)
	}
	if len(d.Devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(d.Devices))
	}
	if d.Devices[0].Device != "LocB___dev2" {
		t.Fatalf("expected LocB___dev2 first, got %s", d.Devices[0].Device)
	}
}

func TestNetworkMoodFromTopDevices(t *testing.T) {
	events := []parse.Event{}
	for i := 0; i < 12; i++ {
		events = append(events, parse.Event{Device: "X___d" + string(rune('a'+i)), State: "WARNING", PacketLoss: 30, RTA: 400})
	}

	d := Build(events)
	if d.NetworkMood == "🙂 klidná" {
		t.Fatalf("expected non-klidná network mood, got %s", d.NetworkMood)
	}
}
