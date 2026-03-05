package iperf

import (
	"encoding/json"
	"errors"
)

func Parse(protocol string, raw []byte) (map[string]any, error) {
	if len(raw) == 0 {
		return nil, errors.New("empty iperf output")
	}
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, err
	}
	end, _ := root["end"].(map[string]any)
	if end == nil {
		return nil, errors.New("missing end section")
	}

	metrics := map[string]any{}
	if protocol == "udp" {
		sum := firstMap(end, "sum", "sum_received", "sum_sent")
		if sum == nil {
			return nil, errors.New("missing udp summary")
		}
		metrics["mbps"] = bitsToMbps(asFloat(sum["bits_per_second"]))
		metrics["jitter_ms"] = asFloat(sum["jitter_ms"])
		metrics["lost_percent"] = asFloat(sum["lost_percent"])
		return metrics, nil
	}

	sent := firstMap(end, "sum_sent", "sum")
	recv := firstMap(end, "sum_received", "sum")
	if sent == nil || recv == nil {
		return nil, errors.New("missing tcp summary")
	}
	metrics["sender_mbps"] = bitsToMbps(asFloat(sent["bits_per_second"]))
	metrics["receiver_mbps"] = bitsToMbps(asFloat(recv["bits_per_second"]))
	metrics["retransmits"] = asInt(sent["retransmits"])
	return metrics, nil
}

func firstMap(root map[string]any, keys ...string) map[string]any {
	for _, k := range keys {
		v, _ := root[k].(map[string]any)
		if v != nil {
			return v
		}
	}
	return nil
}

func bitsToMbps(v float64) float64 {
	return v / 1000.0 / 1000.0
}

func asFloat(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	default:
		return 0
	}
}

func asInt(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int64:
		return int(x)
	default:
		return 0
	}
}
