package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"cloudperf/agent/internal/config"
	"cloudperf/agent/internal/iperf"
)

type Message struct {
	Type      string         `json:"type"`
	RequestID string         `json:"request_id"`
	Timestamp int64          `json:"timestamp"`
	Payload   map[string]any `json:"payload"`
}

type Agent struct {
	cfg config.Config

	writeMu sync.Mutex

	serverMu     sync.Mutex
	serverCancel map[string]context.CancelFunc
}

func New(cfg config.Config) *Agent {
	return &Agent{cfg: cfg, serverCancel: make(map[string]context.CancelFunc)}
}

func (a *Agent) Run(ctx context.Context) {
	for {
		if err := a.runSession(ctx); err != nil {
			log.Printf("session ended: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(a.cfg.ReconnectInterval):
		}
	}
}

func (a *Agent) runSession(ctx context.Context) error {
	u, err := url.Parse(a.cfg.BackendWSURL)
	if err != nil {
		return err
	}
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := a.write(conn, Message{
		Type:      "agent.hello",
		RequestID: newRequestID(),
		Timestamp: time.Now().Unix(),
		Payload: map[string]any{
			"node_id":       a.cfg.NodeID,
			"token":         a.cfg.NodeToken,
			"public_ip":     a.cfg.PublicIP,
			"agent_version": a.cfg.AgentVersion,
		},
	}); err != nil {
		return err
	}

	ticker := time.NewTicker(a.cfg.HeartbeatInterval)
	done := make(chan struct{})
	defer func() {
		close(done)
		ticker.Stop()
	}()
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				_ = a.write(conn, Message{
					Type:      "agent.heartbeat",
					RequestID: newRequestID(),
					Timestamp: time.Now().Unix(),
					Payload: map[string]any{
						"public_ip":     a.cfg.PublicIP,
						"agent_version": a.cfg.AgentVersion,
					},
				})
			}
		}
	}()

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			return err
		}
		switch msg.Type {
		case "agent.welcome":
			log.Printf("connected as node %s", a.cfg.NodeID)
		case "job.start_server":
			go a.handleStartServer(conn, msg)
		case "job.run_client":
			go a.handleRunClient(conn, msg)
		case "job.cancel":
			go a.handleCancel(conn, msg)
		default:
		}
	}
}

func (a *Agent) write(conn *websocket.Conn, msg Message) error {
	a.writeMu.Lock()
	defer a.writeMu.Unlock()
	return conn.WriteJSON(msg)
}

func (a *Agent) handleStartServer(conn *websocket.Conn, msg Message) {
	pairID, _ := msg.Payload["pair_id"].(string)
	port := toInt(msg.Payload["port"], 5201)
	duration := toInt(msg.Payload["duration"], 10)
	if pairID == "" {
		pairID = "pair-unknown"
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(duration+15)*time.Second)
	cmd := exec.CommandContext(ctx, a.cfg.IperfBinary, "-s", "-1", "-J", "-p", strconv.Itoa(port))
	if err := cmd.Start(); err != nil {
		cancel()
		_ = a.write(conn, Message{
			Type:      "job.result",
			RequestID: msg.RequestID,
			Timestamp: time.Now().Unix(),
			Payload:   map[string]any{"ok": false, "error": err.Error()},
		})
		return
	}

	a.serverMu.Lock()
	a.serverCancel[pairID] = cancel
	a.serverMu.Unlock()

	go func() {
		_ = cmd.Wait()
		a.serverMu.Lock()
		delete(a.serverCancel, pairID)
		a.serverMu.Unlock()
		cancel()
	}()

	time.Sleep(300 * time.Millisecond)
	_ = a.write(conn, Message{
		Type:      "job.result",
		RequestID: msg.RequestID,
		Timestamp: time.Now().Unix(),
		Payload:   map[string]any{"ok": true},
	})
}

func (a *Agent) handleRunClient(conn *websocket.Conn, msg Message) {
	targetIP, _ := msg.Payload["target_ip"].(string)
	protocol, _ := msg.Payload["protocol"].(string)
	udpBw, _ := msg.Payload["udp_bandwidth"].(string)
	port := toInt(msg.Payload["port"], 5201)
	duration := toInt(msg.Payload["duration"], 10)
	parallel := toInt(msg.Payload["parallel_streams"], 1)

	args := []string{"-c", targetIP, "-p", strconv.Itoa(port), "-J", "-t", strconv.Itoa(duration), "-P", strconv.Itoa(parallel)}
	if protocol == "udp" {
		if udpBw == "" {
			udpBw = "100M"
		}
		args = append(args, "-u", "-b", udpBw)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(duration+10)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, a.cfg.IperfBinary, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	raw := out.Bytes()
	payload := map[string]any{"ok": err == nil, "raw_output": string(raw)}
	if err != nil {
		payload["error"] = err.Error()
	} else {
		metrics, parseErr := iperf.Parse(protocol, raw)
		if parseErr != nil {
			payload["ok"] = false
			payload["error"] = parseErr.Error()
		} else {
			payload["metrics"] = metrics
		}
	}

	_ = a.write(conn, Message{
		Type:      "job.result",
		RequestID: msg.RequestID,
		Timestamp: time.Now().Unix(),
		Payload:   payload,
	})
}

func (a *Agent) handleCancel(conn *websocket.Conn, msg Message) {
	pairID, _ := msg.Payload["pair_id"].(string)
	if pairID == "" {
		return
	}
	a.serverMu.Lock()
	cancel, ok := a.serverCancel[pairID]
	if ok {
		cancel()
		delete(a.serverCancel, pairID)
	}
	a.serverMu.Unlock()
	_ = a.write(conn, Message{
		Type:      "job.ack",
		RequestID: msg.RequestID,
		Timestamp: time.Now().Unix(),
		Payload:   map[string]any{"ok": ok},
	})
}

func toInt(v any, def int) int {
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	case string:
		i, err := strconv.Atoi(n)
		if err != nil {
			return def
		}
		return i
	default:
		return def
	}
}

func newRequestID() string {
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}

func (m Message) String() string {
	b, _ := json.Marshal(m)
	return string(b)
}
