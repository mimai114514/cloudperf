package ws

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"cloudperf/backend/internal/db"
	"cloudperf/backend/internal/model"
)

type AgentConn struct {
	NodeID string
	Conn   *websocket.Conn
	mu     sync.Mutex
}

func (a *AgentConn) write(msg model.Message) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.Conn.WriteJSON(msg)
}

type Hub struct {
	store *db.Store

	upgrader websocket.Upgrader

	mu      sync.RWMutex
	agents  map[string]*AgentConn
	pending map[string]chan model.Message
}

func NewHub(store *db.Store) *Hub {
	return &Hub{
		store: store,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		agents:  make(map[string]*AgentConn),
		pending: make(map[string]chan model.Message),
	}
}

func (h *Hub) HasAgent(nodeID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.agents[nodeID]
	return ok
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, data, err := conn.ReadMessage()
	if err != nil {
		_ = conn.Close()
		return
	}
	var hello model.Message
	if err := json.Unmarshal(data, &hello); err != nil || hello.Type != "agent.hello" {
		_ = conn.Close()
		return
	}

	nodeID, _ := hello.Payload["node_id"].(string)
	token, _ := hello.Payload["token"].(string)
	publicIP, _ := hello.Payload["public_ip"].(string)
	version, _ := hello.Payload["agent_version"].(string)
	if nodeID == "" || token == "" {
		_ = conn.Close()
		return
	}
	node, err := h.store.ValidateNodeToken(nodeID, db.HashToken(token))
	if err != nil || node == nil {
		_ = conn.Close()
		return
	}
	if err := h.store.UpdateNodeHeartbeat(nodeID, publicIP, version); err != nil {
		_ = conn.Close()
		return
	}

	agent := &AgentConn{NodeID: nodeID, Conn: conn}
	h.mu.Lock()
	h.agents[nodeID] = agent
	h.mu.Unlock()

	_ = agent.write(model.Message{
		Type:      "agent.welcome",
		RequestID: hello.RequestID,
		Timestamp: time.Now().Unix(),
		Payload: map[string]any{
			"node_id": nodeID,
		},
	})

	h.readLoop(agent)
}

func (h *Hub) readLoop(agent *AgentConn) {
	defer func() {
		h.mu.Lock()
		delete(h.agents, agent.NodeID)
		h.mu.Unlock()
		_ = h.store.MarkNodeOffline(agent.NodeID)
		_ = agent.Conn.Close()
	}()

	for {
		var msg model.Message
		if err := agent.Conn.ReadJSON(&msg); err != nil {
			return
		}
		switch msg.Type {
		case "agent.heartbeat":
			ip, _ := msg.Payload["public_ip"].(string)
			version, _ := msg.Payload["agent_version"].(string)
			_ = h.store.UpdateNodeHeartbeat(agent.NodeID, ip, version)
		case "job.ack", "job.result":
			h.mu.RLock()
			ch, ok := h.pending[msg.RequestID]
			h.mu.RUnlock()
			if ok {
				select {
				case ch <- msg:
				default:
				}
			}
		case "job.log":
			// log stream reserved for future persistence.
		default:
		}
	}
}

func (h *Hub) SendCommand(nodeID, typ string, payload map[string]any, timeout time.Duration) (model.Message, error) {
	h.mu.RLock()
	agent, ok := h.agents[nodeID]
	h.mu.RUnlock()
	if !ok {
		return model.Message{}, errors.New("agent not connected")
	}

	reqID := db.NewID("req")
	msg := model.Message{Type: typ, RequestID: reqID, Timestamp: time.Now().Unix(), Payload: payload}
	ch := make(chan model.Message, 1)

	h.mu.Lock()
	h.pending[reqID] = ch
	h.mu.Unlock()

	if err := agent.write(msg); err != nil {
		h.mu.Lock()
		delete(h.pending, reqID)
		h.mu.Unlock()
		return model.Message{}, err
	}

	select {
	case resp := <-ch:
		h.mu.Lock()
		delete(h.pending, reqID)
		h.mu.Unlock()
		return resp, nil
	case <-time.After(timeout):
		h.mu.Lock()
		delete(h.pending, reqID)
		h.mu.Unlock()
		return model.Message{}, errors.New("agent command timeout")
	}
}
