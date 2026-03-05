# CloudPerf 草案文档（v0.1）

## 1. 文档目标

本文档用于固定 CloudPerf 首版实现边界与技术方向，作为后续详细设计与开发拆解的基础。

## 2. 已确认约束

- 产品形态：`Web 前端 + 后端 + VPS Agent`
- 测速协议：`iperf3`，支持 `TCP/UDP` 切换
- 部署方式：不使用容器化（直接进程/服务部署）
- 用户模型：单用户
- 多对多选择方式：在节点 Grid 中手动勾选需要测速的节点组合（不设默认拓扑策略）
- 历史数据：永久保存（首版不做归档清理）
- 告警：首版不做

## 3. 首版范围（MVP）

### 3.1 功能范围

1. 节点管理
- Agent 注册、上线/离线状态
- 节点基础信息展示（IP、地区、标签、最后心跳）

2. 即时测速
- 一对一测速
- 多对多测速（基于 Grid 勾选组合）
- TCP/UDP 参数可切换

3. 定时测速
- 支持 Cron 表达式
- 支持启停定时任务
- 定时触发后生成一次完整 Run

4. 历史记录
- 按任务/时间/节点组合查询
- 展示关键指标与明细
- 基础导出（CSV）

### 3.2 非目标（首版不做）

- 多用户与权限系统
- 告警与通知
- 自动拓扑推荐
- 数据生命周期管理

## 4. 总体架构

```text
[Web Frontend]
     |
  HTTPS API
     |
[Backend API + Scheduler + Orchestrator] ---- [PostgreSQL]
     |
  HTTPS (Agent Pull + Report)
     |
[VPS Agent on Node A/B/C...]
     |
 local exec
     |
 [iperf3 binary]
```

### 4.1 前端职责

- 节点管理页面（列表 + 状态）
- 任务创建页面（一对一/多对多、TCP/UDP、参数设置）
- Grid 勾选界面（选择需要测速的节点对）
- 任务执行详情页（Run 进度、子任务状态、结果）
- 历史查询与导出

### 4.2 后端职责

- REST API
- 调度器（定时任务触发）
- 编排器（将一次任务拆分为多个 node pair job）
- Agent 通信网关（任务下发、结果接收、心跳）
- 数据持久化

### 4.3 Agent 职责

- 启动后注册并周期心跳
- 主动拉取可执行任务
- 执行 iperf3 server/client 子流程
- 上报执行日志与结果
- 本地失败重试与超时处理

## 5. 关键交互流程

### 5.1 节点注册

1. Agent 启动，读取本地配置（backend 地址、token、节点名）
2. Agent 调用注册接口，上报指纹信息（hostname、IP、version）
3. Backend 返回 node_id 与短期会话凭据
4. Agent 每 15s 上报心跳

### 5.2 即时测速（多对多）

1. 用户在 Grid 勾选若干节点组合（A->B, A->C, B->C...）
2. 前端提交一次 test_run 请求（包含协议与参数）
3. Backend 将 run 拆成多个 pair_job
4. 对每个 pair_job：
- 通知目标端 Agent 先启动 iperf3 server（固定窗口超时）
- 通知源端 Agent 发起 iperf3 client
- 汇总结果并落库
5. 前端轮询/订阅 run 状态，展示进度与结果

### 5.3 定时测速

1. 用户创建 schedule（cron + 任务模板）
2. Scheduler 到点触发并创建 test_run
3. 后续流程与即时测速一致

## 6. 技术栈草案（建议）

### 6.1 后端

- 语言：`Go 1.23+`
- Web 框架：`Gin` 或 `Fiber`
- ORM：`GORM`（或直接 `sqlc`）
- 调度器：`robfig/cron`
- 数据库：`PostgreSQL 16`
- 日志：`zap`

### 6.2 前端

- `Vue 3 + TypeScript + Vite`
- UI：`shadcn-vue`（基于 Radix Vue + Tailwind CSS）
- 路由与状态：`Vue Router + Pinia`
- 图表：`ECharts`

### 6.3 Agent

- 语言：`Go`（跨平台单二进制）
- 依赖外部 `iperf3` 可执行文件
- Windows/Linux system service 模式运行

## 7. 数据库模型（初稿）

### 7.1 `nodes`

- `id` (uuid, pk)
- `name` (text)
- `public_ip` (text)
- `private_ip` (text, nullable)
- `region` (text, nullable)
- `tags` (jsonb)
- `status` (text: online/offline)
- `last_heartbeat_at` (timestamptz)
- `agent_version` (text)
- `created_at`, `updated_at`

### 7.2 `test_tasks`

- `id` (uuid, pk)
- `name` (text)
- `mode` (text: one_to_one/many_to_many)
- `protocol` (text: tcp/udp)
- `params` (jsonb)  
  示例：`duration`, `parallel_streams`, `bandwidth`, `packet_len`
- `created_at`, `updated_at`

### 7.3 `schedules`

- `id` (uuid, pk)
- `task_id` (uuid, fk -> test_tasks.id)
- `cron_expr` (text)
- `enabled` (bool)
- `last_run_at` (timestamptz, nullable)
- `next_run_at` (timestamptz, nullable)
- `created_at`, `updated_at`

### 7.4 `test_runs`

- `id` (uuid, pk)
- `task_id` (uuid, fk)
- `trigger_type` (text: manual/schedule)
- `status` (text: pending/running/success/partial_failed/failed)
- `started_at` (timestamptz, nullable)
- `finished_at` (timestamptz, nullable)
- `summary` (jsonb)  
  示例：成功数、失败数、平均吞吐
- `created_at`

### 7.5 `test_run_pairs`

- `id` (uuid, pk)
- `run_id` (uuid, fk -> test_runs.id)
- `source_node_id` (uuid, fk -> nodes.id)
- `target_node_id` (uuid, fk -> nodes.id)
- `status` (text)
- `error_message` (text, nullable)
- `started_at`, `finished_at`

### 7.6 `test_results`

- `id` (uuid, pk)
- `pair_id` (uuid, fk -> test_run_pairs.id)
- `protocol` (text)
- `metrics` (jsonb)  
  TCP 示例：`sender_mbps`, `receiver_mbps`, `retransmits`  
  UDP 示例：`mbps`, `jitter_ms`, `lost_percent`
- `raw_output` (text)  # 原始 iperf3 json/string
- `created_at`

## 8. API 草案（v1）

### 8.1 前端 API

- `POST /api/v1/nodes/register`（Agent 用）
- `POST /api/v1/nodes/heartbeat`（Agent 用）
- `GET /api/v1/nodes`
- `POST /api/v1/tasks`
- `GET /api/v1/tasks`
- `POST /api/v1/runs`（创建即时测速）
- `GET /api/v1/runs/:id`
- `GET /api/v1/runs/:id/pairs`
- `GET /api/v1/results?task_id=&from=&to=`
- `POST /api/v1/schedules`
- `PATCH /api/v1/schedules/:id`
- `GET /api/v1/schedules`

### 8.2 Agent API（拉模式）

- `POST /agent/v1/poll`：Agent 拉取待执行 job
- `POST /agent/v1/job/:id/ack`：接单确认
- `POST /agent/v1/job/:id/result`：上报结果
- `POST /agent/v1/job/:id/log`：分段日志上报

## 9. 执行与并发策略

- 一次 run 可包含多个 pair job
- 并发控制：
- 全局并发上限（例如 20）
- 单节点并发上限（例如 3），避免抢占带宽导致结果失真
- 失败策略：
- server 未就绪：短重试（最多 3 次）
- client 执行失败：标记 pair 失败，不阻断其他 pair
- run 汇总状态按 pair 完成情况计算

## 10. 安全设计（首版最小可用）

- Agent 与 Backend 使用 HTTPS
- 每个 Agent 分配独立 token（可轮换）
- 关键操作写审计日志（任务创建、删除、调度修改）
- 后端接口单用户登录（本地账号 + 会话）

## 11. 非容器化部署草案

### 11.1 后端

- 进程管理：`systemd`（Linux）或 `nssm`（Windows）
- 配置文件：`config.yaml`
- 数据库独立安装 PostgreSQL

### 11.2 前端

- 构建后静态文件部署到 `Nginx`
- 通过反向代理转发 `/api` 到后端

### 11.3 Agent

- 每台 VPS 部署 `cloudperf-agent` + `iperf3`
- 配置 backend 地址与 token
- 注册为系统服务自启动

## 12. 里程碑（建议）

1. M1（1 周）：后端骨架 + Agent 注册心跳 + 节点列表
2. M2（1 周）：一对一即时测速端到端打通（TCP）
3. M3（1 周）：UDP 支持 + 多对多 Grid 勾选执行
4. M4（1 周）：定时任务 + 历史查询 + CSV 导出
5. M5（1 周）：稳定性优化、参数调优、发布文档

## 13. 待确认细节（进入详细设计前）

1. 首版操作系统支持范围（仅 Linux 还是 Linux+Windows）
2. iperf3 参数白名单（哪些开放给前端可配）
3. 多对多时是否允许 `A->B` 与 `B->A` 作为两条独立测试
4. 历史数据查询默认时间窗口（例如最近 7 天）
5. 是否需要任务模板复制功能

