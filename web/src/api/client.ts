export interface NodeItem {
  id: string
  name: string
  public_ip: string
  status: string
  last_heartbeat_at?: string
  agent_version: string
}

export interface RunPair {
  id: string
  run_id: string
  source_node_id: string
  target_node_id: string
  status: string
  error_message?: string
}

export interface RunInfo {
  id: string
  mode: 'one_to_one' | 'many_to_many'
  protocol: 'tcp' | 'udp'
  params: Record<string, unknown>
  status: string
  started_at?: string
  finished_at?: string
  summary?: Record<string, unknown>
}

export interface ResultItem {
  id: string
  pair_id: string
  protocol: 'tcp' | 'udp'
  metrics: Record<string, unknown>
  raw_output: string
  created_at: string
}

const base = import.meta.env.VITE_API_BASE ?? ''

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${base}${path}`, {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', ...(init?.headers ?? {}) },
    ...init,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error ?? `request failed: ${res.status}`)
  }
  if (res.status === 204) return {} as T
  return (await res.json()) as T
}

export const api = {
  login: (username: string, password: string) => req<{ id: string; username: string }>('/api/v1/auth/login', { method: 'POST', body: JSON.stringify({ username, password }) }),
  logout: () => req<void>('/api/v1/auth/logout', { method: 'POST' }),
  me: () => req<{ user_id: string }>('/api/v1/auth/me'),
  listNodes: () => req<{ items: NodeItem[] }>('/api/v1/nodes'),
  createNode: (name: string) => req<{ node: NodeItem; token: string }>('/api/v1/nodes', { method: 'POST', body: JSON.stringify({ name }) }),
  deleteNode: (id: string) => req<void>(`/api/v1/nodes/${id}`, { method: 'DELETE' }),
  createRun: (payload: Record<string, unknown>) => req<{ run_id: string }>('/api/v1/runs', { method: 'POST', body: JSON.stringify(payload) }),
  getRun: (id: string) => req<RunInfo>(`/api/v1/runs/${id}`),
  getRunPairs: (id: string) => req<{ items: RunPair[] }>(`/api/v1/runs/${id}/pairs`),
  listResults: (query = '') => req<{ items: ResultItem[] }>(`/api/v1/results${query}`),
  exportResults: (query = '') => `${base}/api/v1/results/export.csv${query}`,
}
