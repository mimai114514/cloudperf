# CloudPerf 部署文档（P1）

## 1. 目标与拓扑

CloudPerf 采用非容器化部署，推荐拓扑：

- 控制节点（1 台）
- PostgreSQL
- CloudPerf Backend（Go）
- Nginx（前端静态站点 + API/WS 反向代理）
- 多个 Agent 节点（每台 VPS 1 个 cloudperf-agent + iperf3）

## 2. 环境要求

### 2.1 控制节点

- Linux x86_64
- PostgreSQL 14+
- Nginx 1.20+
- Go 1.23+（仅构建时需要）
- Bun 1.1+（仅前端构建时需要）

### 2.2 Agent 节点

- Linux x86_64
- `iperf3` 已安装并可执行

## 3. 目录约定

本文档默认使用以下目录：

- `/opt/cloudperf/backend`
- `/opt/cloudperf/web`
- `/opt/cloudperf/agent`
- `/etc/cloudperf`

## 4. 构建发布物

在代码仓库根目录执行：

```bash
# backend
cd backend
go mod tidy
go build -o cloudperf-backend ./cmd/cloudperf-backend

# agent
cd ../agent
go mod tidy
go build -o cloudperf-agent ./cmd/cloudperf-agent

# web
cd ../web
bun install
bun run build
```

构建产物：

- `backend/cloudperf-backend`
- `agent/cloudperf-agent`
- `web/dist/`

## 5. 控制节点部署

### 5.1 初始化 PostgreSQL

```bash
sudo -u postgres psql -f deploy/postgres-init.sql
```

默认会创建：

- 用户：`cloudperf`
- 数据库：`cloudperf`

### 5.2 准备目录与文件

```bash
sudo mkdir -p /opt/cloudperf/backend/migrations /opt/cloudperf/web /etc/cloudperf

# 复制后端二进制和迁移
sudo cp backend/cloudperf-backend /opt/cloudperf/backend/
sudo cp backend/migrations/001_init.sql /opt/cloudperf/backend/migrations/

# 复制前端静态文件
sudo cp -r web/dist /opt/cloudperf/web/
```

### 5.3 后端环境变量

创建 `/etc/cloudperf/backend.env`：

```env
POSTGRES_DSN=postgres://cloudperf:cloudperf@127.0.0.1:5432/cloudperf?sslmode=disable
BACKEND_ADDR=:8080
ADMIN_USERNAME=admin
ADMIN_PASSWORD=请替换为强密码
COOKIE_SECURE=true
SESSION_TTL_HOURS=12
```

说明：

- 生产环境务必设置 `COOKIE_SECURE=true` 并启用 HTTPS。
- 后端启动时会自动执行数据库迁移并确保管理员账号存在。

### 5.4 配置并启动 backend systemd

```bash
sudo cp deploy/cloudperf-backend.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable cloudperf-backend
sudo systemctl start cloudperf-backend
sudo systemctl status cloudperf-backend
```

### 5.5 配置 Nginx

```bash
sudo cp deploy/nginx-cloudperf.conf /etc/nginx/conf.d/cloudperf.conf
sudo nginx -t
sudo systemctl reload nginx
```

> 若启用 HTTPS，请在 Nginx 中补充 TLS 证书配置，并将站点协议切换为 `https://`。

## 6. Agent 节点部署

## 6.1 安装 iperf3

Debian/Ubuntu：

```bash
sudo apt-get update
sudo apt-get install -y iperf3
```

CentOS/RHEL：

```bash
sudo yum install -y iperf3
```

## 6.2 准备 agent 文件

```bash
sudo mkdir -p /opt/cloudperf/agent /etc/cloudperf
sudo cp agent/cloudperf-agent /opt/cloudperf/agent/
```

## 6.3 创建节点并获取 Token

在 Web 控制台登录后，进入“节点”页面创建节点，记录：

- `NODE_ID`
- 一次性 `NODE_TOKEN`

## 6.4 配置 agent 环境变量

创建 `/etc/cloudperf/agent.env`：

```env
BACKEND_WS_URL=wss://你的域名/agent/v1/ws
NODE_ID=节点ID
NODE_TOKEN=节点TOKEN
PUBLIC_IP=该VPS公网IP
AGENT_VERSION=0.1.0
HEARTBEAT_INTERVAL_SEC=15
RECONNECT_INTERVAL_SEC=5
IPERF_BINARY=iperf3
```

如果未启用 TLS，可临时使用 `ws://控制节点IP:8080/agent/v1/ws`。

## 6.5 配置并启动 agent systemd

```bash
sudo cp deploy/cloudperf-agent.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable cloudperf-agent
sudo systemctl start cloudperf-agent
sudo systemctl status cloudperf-agent
```

## 7. 验收步骤

1. 打开 Web，使用管理员账号登录。
2. 在“节点”页面确认 Agent 节点状态为 `online`。
3. 在“即时测速”页面选择两台节点，执行一次 TCP 测速。
4. 再执行一次 UDP 测速。
5. 在“历史结果”页面查询并导出 CSV。

## 8. 升级流程

1. 备份数据库。
2. 停止后端服务：`sudo systemctl stop cloudperf-backend`。
3. 替换 backend 二进制与迁移文件。
4. 启动后端服务并确认健康。
5. 滚动替换各 Agent 二进制并重启 agent 服务。
6. 若前端有变更，替换 `web/dist` 并 reload nginx。

## 9. 常见问题

- 节点一直离线：
- 检查 `NODE_ID/NODE_TOKEN` 是否匹配。
- 检查 `BACKEND_WS_URL` 与 Nginx `/agent/` 代理配置。
- 检查防火墙是否允许 80/443 或 8080。

- UDP 结果为空：
- 确认目标端 `iperf3` 可执行。
- 检查带宽参数 `udp_bandwidth` 是否合理。

- 登录后 401：
- 反向代理下确认 Cookie 未被拦截。
- 使用 HTTPS 并开启 `COOKIE_SECURE=true`。

## 10. 安全建议（生产）

- 修改默认管理员密码。
- 全站启用 HTTPS。
- 控制节点与数据库仅开放必要端口。
- 定期备份 PostgreSQL。
- 将 Agent 与控制节点通信限制在可信网络范围内。
