package parse

import (
	"testing"
	"time"
)

func TestParseFiltersStatesAndExtractsMetrics(t *testing.T) {
	now := time.Unix(1700000500, 0)
	body := `
<div>[1700000000] SERVICE ALERT: A___dev1;PING;WARNING;SOFT;1;PING WARNING - Packet loss = 28%, RTA = 7.55 ms</div>
[1700000100] SERVICE ALERT: A___dev1;PING;CRITICAL;HARD;1;PING CRITICAL - Packet loss = 100%, RTA = 1500 ms
[1700000200] SERVICE ALERT: A___dev1;PING;OK;HARD;1;OK
[1700000300] SERVICE ALERT: A___dev2;PING;WARNING;HARD;1;PING WARNING - Packet loss = 6%, RTA = 92 ms
`

	events, err := Parse(body, now, 24*time.Hour)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	if events[0].PacketLoss != 28 {
		t.Fatalf("expected packet loss 28, got %v", events[0].PacketLoss)
	}
	if events[0].RTA != 7.55 {
		t.Fatalf("expected rta 7.55, got %v", events[0].RTA)
	}
}

func TestParseRespectsLookbackWhenTimestampAvailable(t *testing.T) {
	now := time.Unix(1700000500, 0)
	body := `
[1699900000] SERVICE ALERT: A___old;PING;WARNING;SOFT;1;PING WARNING - Packet loss = 1%, RTA = 1 ms
[1700000400] SERVICE ALERT: A___new;PING;WARNING;SOFT;1;PING WARNING - Packet loss = 2%, RTA = 2 ms
`

	events, err := Parse(body, now, 2*time.Hour)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Device != "A___new" {
		t.Fatalf("expected device A___new, got %s", events[0].Device)
	}
}

func TestParseFindsBracketedTimestampWithServicePrefix(t *testing.T) {
	now := time.Date(2026, 3, 2, 20, 0, 1, 0, time.Local)
	body := `
Service Warning[2026-03-02 19:45:22] SERVICE ALERT: Sal__Lhota_Kacena___Hlavni;PING;WARNING;SOFT;1;PING WARNING - Packet loss = 44%, RTA = 3.50 ms
Service Warning[2026-03-01 18:45:22] SERVICE ALERT: Old___Device;PING;WARNING;SOFT;1;PING WARNING - Packet loss = 10%, RTA = 9.99 ms
`

	events, err := Parse(body, now, 24*time.Hour)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event in lookback, got %d", len(events))
	}
	if events[0].Device != "Sal__Lhota_Kacena___Hlavni" {
		t.Fatalf("unexpected device: %s", events[0].Device)
	}
	if events[0].PacketLoss != 44 {
		t.Fatalf("unexpected packet loss: %v", events[0].PacketLoss)
	}
}

func TestParseHandlesHTMLUnescapedServiceAlertLine(t *testing.T) {
	now := time.Date(2026, 3, 2, 21, 0, 0, 0, time.Local)
	body := `
<img align='left' src='/x/images/warning.png' alt='Service Warning' title='Service Warning'>[2026-03-02 20:52:41]  SERVICE ALERT&#58; spoj&#95;Crystalis&#95;&#95;&#95;Arnica&#59;PING&#59;CRITICAL&#59;SOFT&#59;1&#59;PING CRITICAL &#45; Packet loss &#61; 60&#37;&#44; RTA &#61; 43&#46;24 ms<br clear='all'>
<img align='left' src='/x/images/warning.png' alt='Service Warning' title='Service Warning'>[2026-03-02 20:51:41]  SERVICE ALERT&#58; spoj&#95;Crystalis&#95;&#95;&#95;Arnica&#59;PING&#59;WARNING&#59;SOFT&#59;1&#59;PING WARNING &#45; Packet loss &#61; 44&#37;&#44; RTA &#61; 3&#46;18 ms<br clear='all'>
`

	events, err := Parse(body, now, 24*time.Hour)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Device != "spoj_Crystalis___Arnica" {
		t.Fatalf("unexpected device parsed: %s", events[0].Device)
	}
	if events[1].State != "WARNING" {
		t.Fatalf("expected WARNING state, got %s", events[1].State)
	}
	if events[1].PacketLoss != 44 {
		t.Fatalf("unexpected packet loss parsed: %v", events[1].PacketLoss)
	}
}
