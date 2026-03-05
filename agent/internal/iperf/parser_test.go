package iperf

import "testing"

func TestParseTCP(t *testing.T) {
	raw := []byte(`{"end":{"sum_sent":{"bits_per_second":125000000.0,"retransmits":3},"sum_received":{"bits_per_second":120000000.0}}}`)
	m, err := Parse("tcp", raw)
	if err != nil {
		t.Fatalf("parse tcp: %v", err)
	}
	if m["sender_mbps"].(float64) != 125 {
		t.Fatalf("expected sender_mbps 125, got %v", m["sender_mbps"])
	}
	if m["retransmits"].(int) != 3 {
		t.Fatalf("expected retransmits 3")
	}
}

func TestParseUDP(t *testing.T) {
	raw := []byte(`{"end":{"sum":{"bits_per_second":80000000.0,"jitter_ms":0.321,"lost_percent":1.2}}}`)
	m, err := Parse("udp", raw)
	if err != nil {
		t.Fatalf("parse udp: %v", err)
	}
	if m["mbps"].(float64) != 80 {
		t.Fatalf("expected mbps 80, got %v", m["mbps"])
	}
}
