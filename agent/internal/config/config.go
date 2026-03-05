package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	BackendWSURL      string
	NodeID            string
	NodeToken         string
	PublicIP          string
	AgentVersion      string
	HeartbeatInterval time.Duration
	ReconnectInterval time.Duration
	IperfBinary       string
}

func Load() Config {
	return Config{
		BackendWSURL:      getEnv("BACKEND_WS_URL", "ws://127.0.0.1:8080/agent/v1/ws"),
		NodeID:            strings.TrimSpace(os.Getenv("NODE_ID")),
		NodeToken:         strings.TrimSpace(os.Getenv("NODE_TOKEN")),
		PublicIP:          strings.TrimSpace(os.Getenv("PUBLIC_IP")),
		AgentVersion:      getEnv("AGENT_VERSION", "0.1.0"),
		HeartbeatInterval: getSeconds("HEARTBEAT_INTERVAL_SEC", 15),
		ReconnectInterval: getSeconds("RECONNECT_INTERVAL_SEC", 5),
		IperfBinary:       getEnv("IPERF_BINARY", "iperf3"),
	}
}

func (c Config) Valid() bool {
	return c.BackendWSURL != "" && c.NodeID != "" && c.NodeToken != ""
}

func getEnv(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

func getSeconds(key string, def int) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return time.Duration(def) * time.Second
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return time.Duration(def) * time.Second
	}
	return time.Duration(n) * time.Second
}
