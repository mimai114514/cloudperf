# CloudPerf 草案文档（v0.2，已对齐当前实现）

## 1. 文档目标

本文档用于描述 CloudPerf 当前代码实现（P1）与后续扩展边界，作为开发与部署基线。

## 2. 当前实现范围（P1）

- 形态：`Web 前端 + 后端 + VPS Agent`
- 协议：`iperf3`，支持 `TCP/UDP`
- 部署：非容器化（systemd + Nginx + PostgreSQL）
- 用户模型：单用户
- 测速模式：一对一 / 多对多（Grid 手动勾选，单向 `A->B`）
- 历史数据：写入 PostgreSQL（无清理策略）
- 定时任务：`未实现（P2）`
- 告警：`未实现（P2）`

## 3. 总体架构（当前）

```text
[Vue Web]
   |
HTTPS (REST)
   |
[Go Backend: API + Session + Orchestrator + Agent WS Hub] ---- [PostgreSQL]
   |
WSS/WS
   |
[Go Agent on VPS] -> exec iperf3
```

### 3.1 Web 前端（`web/`）

- 技术栈：`Vue 3 + TypeScript + Vite + Pinia + Vue Router + Tailwind + Radix Vue + ECharts`
- 已实现页面：
- 登录页
- 节点页（列表/创建节点，显示一次性 token）
- 即时测速页（协议参数 + 节点 Grid 勾选）
- 任务详情页（run 状态 + pair 状态）
- 历史结果页（筛选 + 图表 + CSV 导出）

### 3.2 后端（`backend/`）

- 技术栈：`Go net/http + database/sql + lib/pq + gorilla/websocket`
- 模块：
- Auth：本地账号 + Session Cookie
- Nodes：节点创建、状态查询、心跳更新
- Runs：创建 run、拆分 pair、状态聚合
- Results：查询与 CSV 导出
- Agent WS Hub：维护 agent 长连接与命令应答

### 3.3 Agent（`agent/`）

- 技术栈：Go 单二进制
- 通信：与后端保持 WebSocket 长连接（非轮询）
- 功能：
- 首连 `agent.hello` 鉴权
- 周期心跳 `agent.heartbeat`
- 执行 `job.start_server` / `job.run_client`
- 解析 iperf3 JSON 并回传 `job.result`
- 断线自动重连

## 4. 当前接口定义（已实现）

### 4.1 REST API（前端使用）

- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`
- `POST /api/v1/nodes`
- `GET /api/v1/nodes`
- `POST /api/v1/runs`
- `GET /api/v1/runs/{id}`
- `GET /api/v1/runs/{id}/pairs`
- `GET /api/v1/results`
- `GET /api/v1/results/export.csv`

### 4.2 Agent WebSocket

- 连接：`GET /agent/v1/ws`
- 统一消息结构：`{type, request_id, timestamp, payload}`

Agent -> Backend：
- `agent.hello`
- `agent.heartbeat`
- `job.ack`
- `job.result`
- `job.log`（预留）

Backend -> Agent：
- `agent.welcome`
- `job.start_server`
- `job.run_client`
- `job.cancel`

## 5. 数据模型（已落地）

> 实际 DDL 见 `backend/migrations/001_init.sql`。

### 5.1 `users`
- `id` (text, pk)
- `username` (text, unique)
- `password_hash` (text)
- `created_at` (timestamptz)

### 5.2 `nodes`
- `id` (text, pk)
- `name` (text)
- `token_hash` (text)
- `public_ip` (text)
- `status` (text: online/offline)
- `last_heartbeat_at` (timestamptz, nullable)
- `agent_version` (text)
- `created_at`, `updated_at`

### 5.3 `test_runs`
- `id` (text, pk)
- `mode` (text: one_to_one/many_to_many)
- `protocol` (text: tcp/udp)
- `params_json` (jsonb)
- `status` (text: pending/running/success/partial_failed/failed)
- `started_at`, `finished_at` (timestamptz, nullable)
- `created_at`

### 5.4 `test_run_pairs`
- `id` (text, pk)
- `run_id` (text, fk -> test_runs.id)
- `source_node_id` (text, fk -> nodes.id)
- `target_node_id` (text, fk -> nodes.id)
- `status` (text)
- `error_message` (text)
- `started_at`, `finished_at`, `created_at`

### 5.5 `test_results`
- `id` (text, pk)
- `pair_id` (text, fk -> test_run_pairs.id)
- `protocol` (text)
- `metrics_json` (jsonb)
- `raw_output` (text)
- `created_at`

## 6. 运行策略（已实现）

- 并发：
- 全局并发上限 `20`
- 单节点并发上限 `3`
- 执行流程（每个 pair）：
1. 目标节点启动 iperf3 server
2. 源节点执行 iperf3 client
3. 回传并落库结果
- 失败策略：
- server 未就绪：重试 `3` 次，间隔 `2s`
- client 失败：该 pair 标记失败，不阻断其他 pair
- run 状态聚合：`success / partial_failed / failed`
- 参数默认值：
- `duration=10`
- `parallel_streams=1`
- `udp_bandwidth=100M`

## 7. 安全与配置（当前）

- 登录认证：Session Cookie（`HttpOnly` + `SameSite=Lax`，是否 `Secure` 由 `COOKIE_SECURE` 控制）
- 密码存储：`bcrypt`
- Agent 鉴权：`node_id + token`（后端存储 token hash）
- 关键环境变量：
- Backend：`POSTGRES_DSN`, `BACKEND_ADDR`, `ADMIN_USERNAME`, `ADMIN_PASSWORD`, `COOKIE_SECURE`
- Agent：`BACKEND_WS_URL`, `NODE_ID`, `NODE_TOKEN`, `PUBLIC_IP`, `IPERF_BINARY`

## 8. 与 v0.1 的主要差异

- 移除“Agent 拉模式”描述，改为 WebSocket 长连接模式。
- 移除 `test_tasks/schedules` 与定时任务接口（当前未实现）。
- 技术栈由“建议”更新为“已实现栈”（Go net/http + SQL + WS）。
- 数据模型字段改为与当前 migration 一致（`text` 主键 + `*_json` 字段）。

## 9. 下一阶段（P2）建议

- 定时任务（schedule + cron）
- 告警与通知
- 审计日志持久化
- Session 持久化（替换当前内存会话）
- Agent TLS 证书与更细粒度鉴权
