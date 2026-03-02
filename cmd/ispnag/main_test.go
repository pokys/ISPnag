package main

import (
	"testing"
	"time"
)

func TestLookbackCrossesMidnight(t *testing.T) {
	at0615 := time.Date(2026, 3, 2, 6, 15, 0, 0, time.Local)
	if !lookbackCrossesMidnight(at0615, 24) {
		t.Fatalf("expected 24h lookback at 06:15 to cross midnight")
	}
	if lookbackCrossesMidnight(at0615, 4) {
		t.Fatalf("expected 4h lookback at 06:15 not to cross midnight")
	}

	at2355 := time.Date(2026, 3, 2, 23, 55, 0, 0, time.Local)
	if lookbackCrossesMidnight(at2355, 1) {
		t.Fatalf("expected 1h lookback at 23:55 not to cross midnight")
	}
}
