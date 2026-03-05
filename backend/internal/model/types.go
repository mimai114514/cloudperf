package model

import "time"

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Node struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	PublicIP        string     `json:"public_ip"`
	Status          string     `json:"status"`
	LastHeartbeatAt *time.Time `json:"last_heartbeat_at,omitempty"`
	AgentVersion    string     `json:"agent_version"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type Run struct {
	ID         string         `json:"id"`
	Mode       string         `json:"mode"`
	Protocol   string         `json:"protocol"`
	Params     map[string]any `json:"params"`
	Status     string         `json:"status"`
	StartedAt  *time.Time     `json:"started_at,omitempty"`
	FinishedAt *time.Time     `json:"finished_at,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	Summary    map[string]any `json:"summary,omitempty"`
}

type RunPair struct {
	ID           string     `json:"id"`
	RunID        string     `json:"run_id"`
	SourceNodeID string     `json:"source_node_id"`
	TargetNodeID string     `json:"target_node_id"`
	Status       string     `json:"status"`
	ErrorMessage string     `json:"error_message,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
}

type Result struct {
	ID        string         `json:"id"`
	PairID    string         `json:"pair_id"`
	Protocol  string         `json:"protocol"`
	Metrics   map[string]any `json:"metrics"`
	RawOutput string         `json:"raw_output"`
	CreatedAt time.Time      `json:"created_at"`
}

type Message struct {
	Type      string         `json:"type"`
	RequestID string         `json:"request_id"`
	Timestamp int64          `json:"timestamp"`
	Payload   map[string]any `json:"payload"`
}
