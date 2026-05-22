# dotblue backend

## 初始化

- 后端启动后会先检查 `sys_settings.initialized`。
- 如果存在 `manifest/config/init_data.json`，或通过环境变量 `DOTBLUE_INIT_DATA_PATH` 指定了初始化文件，服务会自动执行首次安装。
- 如果没有初始化文件，系统仍然保留现有的 Web 安装页 `/setup` 作为兜底流程。

## 配置文件

- 运行时配置模板：`manifest/config/config.example.yaml`
- 初始化配置模板：`manifest/config/init_data.example.json`
- 推荐把你自己的运行时配置放在 `manifest/config/config.yaml`
- 推荐把你自己的初始化配置放在 `manifest/config/init_data.json`
- 这两个真实文件都应该保留在本地，不要提交到仓库

## init_data.json

- 示例文件见 `manifest/config/init_data.example.json`。
- 推荐做法：把示例复制为 `backend/manifest/config/init_data.json`，再按部署环境修改。
- `organization.name` 和 `application.name` 必须与 `manifest/config/config.yaml` 中的 `casdoor.organizationName`、`casdoor.applicationName` 保持一致。
- 敏感信息不要直接提交到仓库：
  - 管理员密码使用 `admin.passwordEnv`
  - LLM API Key 使用 `provider.apiKeyEnv`
  - Casdoor 更高权限安装凭据使用 `casdoor.bootstrap.*`

## 自动安装覆盖范围

- 自动确保 Casdoor organization 存在并同步基础品牌字段。
- 自动确保 Casdoor application 存在并同步登录页样式字段。
- 自动确保 `admin` 组存在。
- 自动创建或更新平台管理员账号。
- 自动写入本地 `platform` 和 `provider` 设置，并将系统标记为已初始化。

## 推荐部署方式

1. 复制 `manifest/config/config.example.yaml` 为 `manifest/config/config.yaml`
2. 复制 `manifest/config/init_data.example.json` 为 `manifest/config/init_data.json`
3. 设置环境变量：

```bash
export DOTBLUE_ADMIN_PASSWORD='replace-with-a-strong-password'
export DOTBLUE_LLM_API_KEY='replace-with-provider-key'
```

4. 根据实际 Casdoor 部署修改 `manifest/config/config.yaml`
5. 根据品牌和回调地址修改 `manifest/config/init_data.json`
6. 启动后端，首次启动会自动完成安装

## bootstrap 凭据

- 默认情况下，初始化流程会直接复用运行期 `casdoor.*` 凭据。
- 如果你的运行期应用还没有创建，或者运行期应用没有创建 Organization / Application 的权限，可以额外配置：
  - `casdoor.bootstrap.clientId`
  - `casdoor.bootstrap.clientSecret`
  - `casdoor.bootstrap.jwtSecret`
- 这样安装器会用 bootstrap 应用执行初始化，但运行期登录仍然使用主 `casdoor.*` 配置。
