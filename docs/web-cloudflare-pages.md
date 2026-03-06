# Web 自动部署到 Cloudflare Pages

本项目的前端位于 `web/`，技术栈是 Vue 3 + Vite，构建产物目录为 `web/dist`。仓库已添加 GitHub Actions 工作流，可在推送到 `main` 后自动构建并发布到 Cloudflare Pages。

## 已添加内容

- `.github/workflows/web-cloudflare-pages.yml`
  - 监听 `main` 分支下与 `web/**` 相关的提交
  - 执行 `npm install`
  - 执行 `npm run build`
  - 调用 `cloudflare/pages-action@v1` 发布 `web/dist`
- `web/public/_redirects`
  - 为 Vue Router 的 History 模式提供 SPA 回退
  - 避免在 Pages 上直接刷新 `/nodes`、`/results` 等路由时出现 404

## Cloudflare Pages 配置

1. 在 Cloudflare Pages 创建一个新项目，例如 `cloudperf-web`
2. 该项目不必在 Cloudflare 控制台里再配置构建命令，因为发布由 GitHub Actions 完成
3. 记录以下信息：
   - `Account ID`
   - `Pages project name`
4. 创建一个具备 Pages 编辑权限的 API Token
   - 推荐最小权限：`Cloudflare Pages:Edit`

## GitHub 仓库变量与密钥

在 GitHub 仓库 Settings -> Secrets and variables -> Actions 中配置：

### Secrets

- `CLOUDFLARE_API_TOKEN`
- `CLOUDFLARE_ACCOUNT_ID`

### Variables

- `CLOUDFLARE_PAGES_PROJECT`
  - 示例：`cloudperf-web`
- `VITE_API_BASE`
  - 示例：`https://api.example.com`
  - 如果前端与后端走同域反向代理，可留空或不设置

## 部署触发规则

- 推送到 `main` 且变更命中以下路径时自动部署：
  - `web/**`
  - `.github/workflows/web-cloudflare-pages.yml`
- 也支持在 GitHub Actions 页面手动触发

## 后端联通说明

前端代码通过 `VITE_API_BASE` 访问后端：

- 文件：`web/src/api/client.ts`
- 默认逻辑：`const base = import.meta.env.VITE_API_BASE ?? ''`

这意味着有两种典型部署方式：

1. 前后端同域
   - `VITE_API_BASE` 留空
   - 由网关把 `/api/*` 反代到后端
2. 前后端分域
   - 将 `VITE_API_BASE` 设置为后端完整地址
   - 例如 `https://api.cloudperf.example.com`
   - 后端需要正确配置 CORS 和 Cookie 策略

## 建议

- 当前工作流使用 `npm install`，因为仓库里尚未提交 lockfile
- 如果后续提交 `package-lock.json`，建议将工作流改为 `npm ci` 以获得更稳定的构建
