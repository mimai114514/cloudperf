# CloudPerf (P1)

CloudPerf 是一个用于多 VPS 间 iperf3 测速的系统，包含：

- `web/`：Vue3 + shadcn-vue 风格前端
- `backend/`：Go API + WebSocket Agent 网关 + 编排执行
- `agent/`：Go Agent，执行 iperf3 并回传结果
- `deploy/`：非容器化部署模板

## 功能

- 单用户登录
- 节点创建、Token 下发、在线状态展示
- TCP/UDP 一对一/多对多即时测速
- 历史结果查询与 CSV 导出

## 本地启动（开发）

默认在 `WSL2` 中进行本地开发。

建议先在 WSL2 内拉取最新代码，再开始本地联调：

```bash
mkdir -p ~/code
cd ~/code
git clone <你的仓库地址> cloudperf
# 已有仓库时
cd ~/code/cloudperf
git pull --ff-only
```

1. 准备 PostgreSQL 并创建数据库：`deploy/postgres-init.sql`
2. 启动后端：
   - 进入 `backend/`
   - 配置环境变量（参考 `deploy/backend.env.example`）
   - 运行 `go run ./cmd/cloudperf-backend`
3. 启动 Agent：
   - 进入 `agent/`
   - 先在 UI 创建节点拿到 `NODE_ID` 和 `NODE_TOKEN`
   - 配置环境变量（参考 `deploy/agent.env.example`）
   - 运行 `go run ./cmd/cloudperf-agent`
4. 启动前端：
   - 进入 `web/`
   - `bun install`
   - `unset VITE_API_BASE`
   - `bun run dev`

更完整的单机联调说明见 [`docs/single-machine-dev.md`](/C:/Users/Infinity/Documents/Code/cloudperf/docs/single-machine-dev.md)。

## 默认账号

- 用户名：`admin`
- 密码：`admin123`

请在生产环境修改 `ADMIN_PASSWORD`，并启用 HTTPS + `COOKIE_SECURE=true`。
