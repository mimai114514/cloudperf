CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS nodes (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  token_hash TEXT NOT NULL,
  public_ip TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'offline',
  last_heartbeat_at TIMESTAMPTZ,
  agent_version TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS test_runs (
  id TEXT PRIMARY KEY,
  mode TEXT NOT NULL,
  protocol TEXT NOT NULL,
  params_json JSONB NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending',
  started_at TIMESTAMPTZ,
  finished_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS test_run_pairs (
  id TEXT PRIMARY KEY,
  run_id TEXT NOT NULL REFERENCES test_runs(id) ON DELETE CASCADE,
  source_node_id TEXT NOT NULL REFERENCES nodes(id),
  target_node_id TEXT NOT NULL REFERENCES nodes(id),
  status TEXT NOT NULL DEFAULT 'pending',
  error_message TEXT NOT NULL DEFAULT '',
  started_at TIMESTAMPTZ,
  finished_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS test_results (
  id TEXT PRIMARY KEY,
  pair_id TEXT NOT NULL REFERENCES test_run_pairs(id) ON DELETE CASCADE,
  protocol TEXT NOT NULL,
  metrics_json JSONB NOT NULL,
  raw_output TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_nodes_status ON nodes(status);
CREATE INDEX IF NOT EXISTS idx_runs_created_at ON test_runs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_pairs_run_id ON test_run_pairs(run_id);
CREATE INDEX IF NOT EXISTS idx_results_pair_id ON test_results(pair_id);
