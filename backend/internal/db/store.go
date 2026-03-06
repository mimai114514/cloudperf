package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"cloudperf/backend/internal/model"
)

type PairSpec struct {
	SourceNodeID string `json:"source_node_id"`
	TargetNodeID string `json:"target_node_id"`
}

type Store struct {
	db *sql.DB
}

func Open(dsn string) (*Store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Migrate(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(string(b))
	return err
}

func (s *Store) EnsureUser(id, username, passwordHash string) error {
	_, err := s.db.Exec(`
		INSERT INTO users (id, username, password_hash)
		VALUES ($1, $2, $3)
		ON CONFLICT (username) DO NOTHING
	`, id, username, passwordHash)
	return err
}

func (s *Store) GetUserByUsername(username string) (*model.User, error) {
	row := s.db.QueryRow(`SELECT id, username, password_hash, created_at FROM users WHERE username = $1`, username)
	var u model.User
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (s *Store) CreateNode(id, name, tokenHash string) (*model.Node, error) {
	_, err := s.db.Exec(`
		INSERT INTO nodes (id, name, token_hash)
		VALUES ($1, $2, $3)
	`, id, name, tokenHash)
	if err != nil {
		return nil, err
	}
	return s.GetNodeByID(id)
}

func (s *Store) GetNodeByID(id string) (*model.Node, error) {
	row := s.db.QueryRow(`
		SELECT id, name, public_ip, status, last_heartbeat_at, agent_version, created_at, updated_at
		FROM nodes WHERE id = $1
	`, id)
	return scanNode(row)
}

func (s *Store) GetNodeByTokenHash(tokenHash string) (*model.Node, error) {
	row := s.db.QueryRow(`
		SELECT id, name, public_ip, status, last_heartbeat_at, agent_version, created_at, updated_at
		FROM nodes WHERE token_hash = $1
	`, tokenHash)
	return scanNode(row)
}

func scanNode(row scanner) (*model.Node, error) {
	var n model.Node
	if err := row.Scan(&n.ID, &n.Name, &n.PublicIP, &n.Status, &n.LastHeartbeatAt, &n.AgentVersion, &n.CreatedAt, &n.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &n, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func (s *Store) ListNodes() ([]model.Node, error) {
	rows, err := s.db.Query(`
		SELECT id, name, public_ip, status, last_heartbeat_at, agent_version, created_at, updated_at
		FROM nodes ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.Node, 0)
	for rows.Next() {
		var n model.Node
		if err := rows.Scan(&n.ID, &n.Name, &n.PublicIP, &n.Status, &n.LastHeartbeatAt, &n.AgentVersion, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, n)
	}
	return items, rows.Err()
}

func (s *Store) UpdateNodeHeartbeat(nodeID, ip, version string) error {
	_, err := s.db.Exec(`
		UPDATE nodes
		SET public_ip = COALESCE(NULLIF($2, ''), public_ip),
		    status = 'online',
		    agent_version = COALESCE(NULLIF($3, ''), agent_version),
		    last_heartbeat_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
	`, nodeID, ip, version)
	return err
}

func (s *Store) MarkNodeOffline(nodeID string) error {
	_, err := s.db.Exec(`UPDATE nodes SET status = 'offline', updated_at = NOW() WHERE id = $1`, nodeID)
	return err
}

func (s *Store) DeleteNode(nodeID string) error {
	_, err := s.db.Exec(`DELETE FROM nodes WHERE id = $1`, nodeID)
	return err
}

func (s *Store) CreateRun(id, mode, protocol string, params map[string]any, pairs []PairSpec) error {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		INSERT INTO test_runs (id, mode, protocol, params_json, status)
		VALUES ($1, $2, $3, $4::jsonb, 'pending')
	`, id, mode, protocol, string(paramsJSON)); err != nil {
		return err
	}

	for _, p := range pairs {
		pairID := NewID("pair")
		if _, err := tx.Exec(`
			INSERT INTO test_run_pairs (id, run_id, source_node_id, target_node_id, status)
			VALUES ($1, $2, $3, $4, 'pending')
		`, pairID, id, p.SourceNodeID, p.TargetNodeID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) GetRun(runID string) (*model.Run, error) {
	row := s.db.QueryRow(`
		SELECT id, mode, protocol, params_json, status, started_at, finished_at, created_at
		FROM test_runs WHERE id = $1
	`, runID)

	var run model.Run
	var raw []byte
	if err := row.Scan(&run.ID, &run.Mode, &run.Protocol, &raw, &run.Status, &run.StartedAt, &run.FinishedAt, &run.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(raw, &run.Params); err != nil {
		return nil, err
	}

	summary, err := s.RunSummary(runID)
	if err != nil {
		return nil, err
	}
	run.Summary = summary
	return &run, nil
}

func (s *Store) ListRunPairs(runID string) ([]model.RunPair, error) {
	rows, err := s.db.Query(`
		SELECT id, run_id, source_node_id, target_node_id, status, error_message, started_at, finished_at
		FROM test_run_pairs WHERE run_id = $1 ORDER BY created_at ASC
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	pairs := make([]model.RunPair, 0)
	for rows.Next() {
		var p model.RunPair
		if err := rows.Scan(&p.ID, &p.RunID, &p.SourceNodeID, &p.TargetNodeID, &p.Status, &p.ErrorMessage, &p.StartedAt, &p.FinishedAt); err != nil {
			return nil, err
		}
		pairs = append(pairs, p)
	}
	return pairs, rows.Err()
}

func (s *Store) UpdateRunStatus(runID, status string, setStarted, setFinished bool) error {
	clauses := []string{"status = $2"}
	if setStarted {
		clauses = append(clauses, "started_at = COALESCE(started_at, NOW())")
	}
	if setFinished {
		clauses = append(clauses, "finished_at = NOW()")
	}
	q := fmt.Sprintf("UPDATE test_runs SET %s WHERE id = $1", strings.Join(clauses, ", "))
	_, err := s.db.Exec(q, runID, status)
	return err
}

func (s *Store) UpdatePairStatus(pairID, status, errMsg string, setStarted, setFinished bool) error {
	clauses := []string{"status = $2", "error_message = $3"}
	if setStarted {
		clauses = append(clauses, "started_at = COALESCE(started_at, NOW())")
	}
	if setFinished {
		clauses = append(clauses, "finished_at = NOW()")
	}
	q := fmt.Sprintf("UPDATE test_run_pairs SET %s WHERE id = $1", strings.Join(clauses, ", "))
	_, err := s.db.Exec(q, pairID, status, errMsg)
	return err
}

func (s *Store) InsertResult(id, pairID, protocol string, metrics map[string]any, rawOutput string) error {
	m, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
		INSERT INTO test_results (id, pair_id, protocol, metrics_json, raw_output)
		VALUES ($1, $2, $3, $4::jsonb, $5)
	`, id, pairID, protocol, string(m), rawOutput)
	return err
}

func (s *Store) RunSummary(runID string) (map[string]any, error) {
	rows, err := s.db.Query(`SELECT status, COUNT(*) FROM test_run_pairs WHERE run_id = $1 GROUP BY status`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	summary := map[string]any{"total": 0, "pending": 0, "running": 0, "success": 0, "failed": 0}
	for rows.Next() {
		var status string
		var c int
		if err := rows.Scan(&status, &c); err != nil {
			return nil, err
		}
		summary[status] = c
		summary["total"] = summary["total"].(int) + c
	}
	return summary, rows.Err()
}

func (s *Store) ListResults(runID, from, to string, limit int) ([]model.Result, error) {
	query := `
		SELECT r.id, r.pair_id, r.protocol, r.metrics_json, r.raw_output, r.created_at
		FROM test_results r
		JOIN test_run_pairs p ON p.id = r.pair_id
		JOIN test_runs t ON t.id = p.run_id
		WHERE 1 = 1
	`
	args := make([]any, 0)
	idx := 1
	if runID != "" {
		query += fmt.Sprintf(" AND t.id = $%d", idx)
		args = append(args, runID)
		idx++
	}
	if from != "" {
		query += fmt.Sprintf(" AND r.created_at >= $%d::timestamptz", idx)
		args = append(args, from)
		idx++
	}
	if to != "" {
		query += fmt.Sprintf(" AND r.created_at <= $%d::timestamptz", idx)
		args = append(args, to)
		idx++
	}
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	query += fmt.Sprintf(" ORDER BY r.created_at DESC LIMIT $%d", idx)
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([]model.Result, 0)
	for rows.Next() {
		var res model.Result
		var raw []byte
		if err := rows.Scan(&res.ID, &res.PairID, &res.Protocol, &raw, &res.RawOutput, &res.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &res.Metrics); err != nil {
			return nil, err
		}
		results = append(results, res)
	}
	return results, rows.Err()
}

func (s *Store) ListRunPairsForExecution(runID string) ([]model.RunPair, error) {
	return s.ListRunPairs(runID)
}

func (s *Store) GetNodeIP(nodeID string) (string, error) {
	var ip string
	if err := s.db.QueryRow(`SELECT public_ip FROM nodes WHERE id = $1`, nodeID).Scan(&ip); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return ip, nil
}

func (s *Store) ValidateNodeToken(nodeID, tokenHash string) (*model.Node, error) {
	row := s.db.QueryRow(`
		SELECT id, name, public_ip, status, last_heartbeat_at, agent_version, created_at, updated_at
		FROM nodes WHERE id = $1 AND token_hash = $2
	`, nodeID, tokenHash)
	return scanNode(row)
}
