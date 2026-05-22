# dotblue web

## 环境变量

- 示例文件：`.env.example`
- 本地开发可复制为 `.env.local`
- 关键变量：
  - `VITE_CASDOOR_SERVER_URL`
  - `VITE_CASDOOR_CLIENT_ID`
  - `VITE_CASDOOR_ORG_NAME`
  - `VITE_CASDOOR_APP_NAME`
  - `VITE_BACKEND_URL`

## 启动

```bash
npm install
npm run dev
```

## 说明

- 前端的 Casdoor `organization` 和 `application` 必须与后端 `manifest/config/config.yaml` 以及 `init_data.json` 保持一致。
- 如果后端启用了自动初始化，前端只需要指向正确的 Casdoor 和后端地址即可。
