# CloudPerf 单机开发部署文档（WSL2）

本文档用于在 WSL2 环境里跑通 CloudPerf 的完整本地联调，包括：

- PostgreSQL
- backend
- web
- 2 个本地 agent

适用场景：

- 开发页面和接口
- 验证登录、节点上线、即时测速、历史结果导出
- 在提交前做最小可用链路自测

默认约定：

- 开发环境运行在 `WSL2 Ubuntu`
- 仓库位于 `~/code/cloudperf`
- 浏览器仍然使用 Windows 侧访问 `http://127.0.0.1:5173`

## 1. 单机联调拓扑

```text
浏览器 -> Vite(5173) -> Backend(8080) -> PostgreSQL
                               |
                               +-> Agent A（本机）
                               +-> Agent B（本机）
```

说明：

- 前端开发服务器默认运行在 `http://127.0.0.1:5173`
- 后端默认运行在 `http://127.0.0.1:8080`
- Vite 开发服务器会将 `/api` 和 `/agent` 代理到后端
- 两个 Agent 都连接同一个后端
- 两个 Agent 的 `PUBLIC_IP` 在单机联调时都可以填 `127.0.0.1`
- 后端会为每个测速任务动态分配 `5201+` 端口，因此同机运行多个 Agent 不会固定占用同一个 iperf3 端口

## 2. 环境要求

- Go `1.23+`
- Bun `1.1+`
- PostgreSQL `14+`
- `iperf3`

如果你使用 Ubuntu WSL2，可按需安装：

```bash
sudo apt-get update
sudo apt-get install -y postgresql postgresql-client golang-go iperf3 unzip
```

如果 Bun 尚未安装，可在 WSL2 内执行：

```bash
curl -fsSL https://bun.sh/install | bash
source ~/.bashrc
```

建议先确认版本：

```bash
go version
bun --version
psql --version
iperf3 --version
```

## 3. 一次性准备

### 3.0 在 WSL2 拉取最新代码

建议不要直接在 `/mnt/c/...` 下开发，而是在 WSL2 自己的 Linux 文件系统里拉取仓库，文件监听和依赖安装会更稳定。

首次拉取：

```bash
mkdir -p ~/code
cd ~/code
git clone <你的仓库地址> cloudperf
cd ~/code/cloudperf
```

后续更新：

```bash
cd ~/code/cloudperf
git pull --ff-only
```

如果你已经在 WSL2 里有一份仓库，后续所有命令都以 `~/code/cloudperf` 为根目录执行。

### 3.1 初始化数据库

先确保 PostgreSQL 已启动，然后执行：

```bash
cd ~/code/cloudperf
sudo -u postgres psql -f ./deploy/postgres-init.sql
```

默认会创建：

- 数据库：`cloudperf`
- 用户：`cloudperf`
- 密码：`cloudperf`

如果本机 PostgreSQL 不允许本地免密执行，请按你的实际账号补充 `-h`、`-p` 和密码输入。

### 3.2 安装前端依赖

```bash
cd ~/code/cloudperf/web
bun install
```

### 3.3 可选：检查 Go 依赖

```bash
cd ~/code/cloudperf/backend
go mod tidy

cd ~/code/cloudperf/agent
go mod tidy
```

## 4. 启动顺序

建议总共开 5 个 WSL2 终端：

1. `backend`
2. `web`
3. `agent-a`
4. `agent-b`
5. 可选的 `psql` 或日志观察终端

## 5. 启动 backend

必须在 [`backend`](/C:/Users/Infinity/Documents/Code/cloudperf/backend) 目录内启动。当前代码会读取相对路径 `migrations/001_init.sql`，如果你从别的目录运行，迁移会找不到。

```bash
cd ~/code/cloudperf/backend
export POSTGRES_DSN="postgres://cloudperf:cloudperf@127.0.0.1:5432/cloudperf?sslmode=disable"
export BACKEND_ADDR=":8080"
export ADMIN_USERNAME="admin"
export ADMIN_PASSWORD="admin123"
export COOKIE_SECURE="false"
export SESSION_TTL_HOURS="12"
go run ./cmd/cloudperf-backend
```

成功后会看到类似日志：

```text
backend listening on :8080
```

说明：

- 本地开发必须使用 `COOKIE_SECURE=false`
- 默认管理员账号是 `admin / admin123`
- 后端启动时会自动建表并确保管理员账号存在

## 6. 启动 web

WSL2 本地开发默认通过 Vite 反向代理访问后端，不需要设置 `VITE_API_BASE`。如果你之前导出过这个变量，先清掉，避免浏览器继续直连 `8080`。

```bash
cd ~/code/cloudperf/web
unset VITE_API_BASE
bun run dev
```

然后在 Windows 浏览器里打开：

- `http://127.0.0.1:5173`

注意：

- 请统一使用 `127.0.0.1`
- 不要前端用 `localhost`、后端和 Agent 用 `127.0.0.1`
- 由于前端现在通过 Vite 代理访问后端，登录和后续接口请求都会走同一源站 `5173`
- 默认情况下，WSL2 的本地端口会自动映射到 Windows 主机

## 7. 在页面里创建两个节点

打开前端后：

1. 使用 `admin / admin123` 登录
2. 进入“节点”页面
3. 连续创建两个节点，例如 `local-a`、`local-b`
4. 记录页面返回的：
   - `NODE_ID`
   - `NODE_TOKEN`

注意：

- `NODE_TOKEN` 只会在创建时返回一次
- 如果关掉弹窗或刷新页面后丢失，需要重新创建节点

## 8. 启动两个本地 Agent

### 8.1 Agent A

```bash
cd ~/code/cloudperf/agent
export BACKEND_WS_URL="ws://127.0.0.1:8080/agent/v1/ws"
export NODE_ID="node_1772815178274814605_85aabf1ae4932239"
export NODE_TOKEN="a77420a0337697d7fd70af69526a872901d5f5fa9244f85b3e7d50035bbca1c0"
export PUBLIC_IP="127.0.0.1"
export AGENT_VERSION="dev-a"
export HEARTBEAT_INTERVAL_SEC="15"
export RECONNECT_INTERVAL_SEC="5"
export IPERF_BINARY="iperf3"
go run ./cmd/cloudperf-agent
```

### 8.2 Agent B

```bash
cd ~/code/cloudperf/agent
export BACKEND_WS_URL="ws://127.0.0.1:8080/agent/v1/ws"
export NODE_ID="node_1772815210056814243_ece0606ff47ff32a"
export NODE_TOKEN="958f7107b858cb0f45da95f9bd40acee724d3c44a7fcfa66a3565eb7a69afd9d"
export PUBLIC_IP="127.0.0.1"
export AGENT_VERSION="dev-b"
export HEARTBEAT_INTERVAL_SEC="15"
export RECONNECT_INTERVAL_SEC="5"
export IPERF_BINARY="iperf3"
go run ./cmd/cloudperf-agent
```

Agent 成功连接后，日志里通常会看到：

```text
connected as node ...
```

回到“节点”页面，两个节点都应显示为 `online`。

## 9. 执行一次单机测速

进入“即时测速”页面后：

1. 选择协议 `TCP`
2. 选择一个 `A -> B` 组合
3. 保持默认参数，先跑一次
4. 确认任务详情页显示 `success`
5. 再切换成 `UDP` 跑一次

由于两个 Agent 都在本机，测速结果主要用于验证链路是否正常，不代表真实公网吞吐。

## 10. 最小验收标准

满足以下几点即可认为单机开发环境可用：

1. 可以登录后台
2. 两个节点状态都为 `online`
3. 至少一条 TCP 任务执行成功
4. 至少一条 UDP 任务执行成功
5. “历史结果”页面可以查询到记录
6. CSV 导出可下载

## 11. 常见问题

### 11.1 点击登录提示 `Failed to fetch`

优先检查：

- 是否重新启动了前端开发服务器
- 当前终端里是否已经执行 `unset VITE_API_BASE`
- 前端是否通过 `bun run dev` 启动，而不是旧的缓存进程
- backend 是否已经正常监听 `127.0.0.1:8080`

可以在 WSL2 内直接验证：

```bash
curl http://127.0.0.1:8080/api/v1/auth/me
```

只要能收到后端响应，前端代理通常就能正常工作。

### 11.2 登录后报 401 或一直跳回登录页

检查：

- backend 是否以 `COOKIE_SECURE=false` 启动
- 浏览器访问地址是否统一使用 `127.0.0.1`
- 是否误设置了旧的 `VITE_API_BASE`

### 11.3 节点一直离线

检查：

- `NODE_ID` 和 `NODE_TOKEN` 是否填错
- backend 是否已经启动
- Agent 日志中是否有 `connected as node`
- `BACKEND_WS_URL` 是否为 `ws://127.0.0.1:8080/agent/v1/ws`

### 11.4 任务创建成功，但 pair 失败

检查：

- `iperf3` 是否已安装并能直接在命令行执行
- 两个节点是否都有 `PUBLIC_IP=127.0.0.1`
- 两个 Agent 是否都在线

可以直接执行：

```bash
iperf3 --version
```

### 11.5 backend 启动时报 migration 文件不存在

原因通常是启动目录不对。

修正方式：

- 在 WSL2 中进入 `~/code/cloudperf/backend` 再运行
- 或者确保二进制旁边存在 `migrations/001_init.sql`

## 12. 常用重置方式

如果你只是要重新做一次本地联调，最简单的方式通常是：

1. 停掉 backend、web、两个 Agent
2. 清空数据库中的测试数据或直接重建 `cloudperf` 数据库
3. 重启 backend
4. 重新创建设备节点
5. 重新启动两个 Agent

如果需要彻底清库，可手动删除并重建数据库后，再重新执行 [`deploy/postgres-init.sql`](/C:/Users/Infinity/Documents/Code/cloudperf/deploy/postgres-init.sql)。
