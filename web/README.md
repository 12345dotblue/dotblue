# dotblue web

## 环境变量

- 示例文件：`.env.example`
- 本地开发可复制为 `.env.local`
- 本地开发和 CI 构建仍可通过 Vite 环境变量提供默认值
- 容器部署时，前端会在启动时生成 `/runtime-config.js`，优先读取运行时环境变量
- 关键变量：
  - `VITE_CASDOOR_SERVER_URL`
  - `VITE_CASDOOR_CLIENT_ID`
  - `VITE_CASDOOR_ORG_NAME`
  - `VITE_CASDOOR_APP_NAME`
  - `VITE_CASDOOR_REDIRECT_URL`（可选）
  - `VITE_BACKEND_URL`
- 也支持同义的容器运行时变量：
  - `DOTBLUE_WEB_CASDOOR_SERVER_URL`
  - `DOTBLUE_WEB_CASDOOR_CLIENT_ID`
  - `DOTBLUE_WEB_CASDOOR_ORG_NAME`
  - `DOTBLUE_WEB_CASDOOR_APP_NAME`
  - `DOTBLUE_WEB_CASDOOR_REDIRECT_URL`
  - `DOTBLUE_WEB_BACKEND_URL`

## 启动

```bash
npm install
npm run dev
```

## 说明

- 前端的 Casdoor `organization` 和 `application` 必须与后端 `manifest/config/config.yaml` 以及 `init_data.json` 保持一致。
- 如果后端启用了自动初始化，前端只需要指向正确的 Casdoor 和后端地址即可。
- 私有化部署时，推荐直接复用同一个 `dotblue-web` 镜像，只改容器环境变量，不要为每个环境重新构建前端镜像。
