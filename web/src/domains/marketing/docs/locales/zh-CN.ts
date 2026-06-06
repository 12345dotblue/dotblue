import type { DocsLibrary } from '../schema';

export const library: DocsLibrary = {
  "requestedLanguage": "zh-CN",
  "contentLanguage": "zh-CN",
  "isFallbackContent": false,
  "homeSeoTitle": "dotblue 文档中心 | 从产品认知到安装部署上线",
  "homeSeoDescription": "面向第一次接触 dotblue 的团队，系统说明产品能做什么、应该怎么使用，以及如何从零完成快速安装、首次验证和生产环境上线。",
  "homeSeoKeywords": "dotblue 文档,企业级 AI 助手平台,快速上手,安装部署,生产环境,模型配置,Casdoor,运维排障",
  "eyebrow": "PRODUCT DOCS",
  "title": "dotblue 文档：先看懂产品，再把它装起来并上线",
  "subtitle": "这套文档面向第一次接触 dotblue 的团队，先讲清它能解决什么问题，再带你从 0 到 1 完成本地安装、首次登录、首个助手验证，以及生产环境上线准备。",
  "categoriesLabel": "文档分类",
  "popularLabel": "建议先读",
  "sectionDescriptionLabel": "这个分类重点解决什么问题",
  "quickLinksTitle": "快速入口",
  "quickLinks": [
    {
      "label": "打开产品首页",
      "url": "/zh-CN"
    },
    {
      "label": "进入登录页",
      "url": "/zh-CN/login"
    },
    {
      "label": "GitHub",
      "url": "https://github.com/12345dotblue/dotblue",
      "description": "查看源码、Compose 资产、配置模板和发布相关实现。"
    }
  ],
  "repoTitle": "源码仓库与部署资产",
  "repoDescription": "产品文档负责回答“它是什么、怎么装、怎么用、怎么上线”，官方仓库负责回答“配置模板长什么样、脚本到底生成了什么、具体实现怎么落地”。两者配合起来，才是一套真正能交付的资料。",
  "sections": [
    {
      "slug": "getting-started",
      "title": "快速开始",
      "description": "适合第一次接触 dotblue 的人。先建立产品认知，再完成最短可执行安装路径，把首页、登录、控制台、模型、助手和聊天完整跑通。",
      "articles": [
        {
          "sectionSlug": "getting-started",
          "slug": "dotblue-overview",
          "title": "dotblue 概览",
          "summary": "先看懂 dotblue 是什么、适合谁、能解决什么问题，以及第一次成功到底应该长什么样。",
          "seoTitle": "dotblue 概览 | dotblue 产品文档",
          "seoDescription": "面向第一次接触 dotblue 的团队，解释产品定位、核心能力、关键概念和首次成功路径。",
          "readingTime": "10 分钟阅读",
          "sections": [
            {
              "id": "what-it-is",
              "title": "dotblue 到底是什么",
              "paragraphs": [
                "dotblue 是一套企业级 AI 助手交付平台。它不只是聊天页面，也不只是模型转发层，而是把登录认证、工作空间、模型配置、助手配置、聊天验证、运行治理和部署资产放进同一套产品里。",
                "如果你面对的是“我们想把 AI 助手真的交给团队、客户或业务部门用起来”，dotblue 解决的是这件事的全链路问题，而不是只解决其中某一个点。"
              ]
            },
            {
              "id": "what-problems-it-solves",
              "title": "它主要解决什么问题",
              "bullets": [
                "不再把登录、模型配置、助手管理、聊天入口和部署流程拆成多个零散工具分别拼接。",
                "让产品团队和实施团队可以更快把一个业务场景做成可登录、可配置、可验证的 AI 助手产品面。",
                "让平台团队能够统一治理模型入口、管理员账号、组织边界和运行时行为。",
                "让自托管场景不必从零搭认证、前端、后端、数据库、运行时编排和配置生成链。",
                "让上线后的日常运维有明确抓手，而不是出了问题只能猜。"
              ]
            },
            {
              "id": "core-capabilities",
              "title": "核心能力一眼看懂",
              "bullets": [
                "品牌化认证：通过 Casdoor 提供登录、注册、回调和退出登录能力。",
                "助手管理：创建不同业务助手，维护提示词、模型和助手边界。",
                "模型治理：在平台级或企业级统一配置模型供应商、API Base、API Key 和模型名。",
                "聊天验证：通过真实聊天页验证助手是否真正可用，而不是停留在配置层。",
                "组织与隔离：为多团队、多组织、多客户交付预留边界。",
                "部署资产：提供 Compose、配置生成脚本、排障脚本和生产上线所需的基础资产。"
              ]
            },
            {
              "id": "key-concepts",
              "title": "第一次使用前必须搞懂的概念",
              "bullets": [
                "平台级配置：跨整个环境共享的基础能力，例如默认模型供应商、全局认证策略和系统级设置。",
                "企业或组织上下文：当不同团队、客户或租户需要隔离时，用来划分权限和配置边界。",
                "助手：真正交付给用户使用的 AI 能力单元，应该围绕一个明确业务任务来设计。",
                "模型配置：助手能否工作，先取决于平台或企业层是否已经保存了真实可用的模型配置。",
                "Chat：最终验收面。只有在聊天页能选中助手、发出消息并得到合理回复，才说明链路真的通了。",
                "生成配置：`prepare-config` 不是可有可无的小脚本，它决定了 Web、Backend 和 Casdoor 是否使用一致的公开地址和品牌资源。"
              ]
            },
            {
              "id": "who-should-read",
              "title": "这套产品适合谁",
              "paragraphs": [
                "如果你是产品负责人，这套系统适合你把客服、知识库、销售辅助、运营支持这类场景做成真正可上线的 AI 助手产品。",
                "如果你是实施或交付团队，它适合你把一套助手能力交付到不同企业或环境里，而不是每次从零搭壳子。",
                "如果你是平台或运维团队，它适合你统一管理模型接入、运行时和上线流程，让 AI 助手不再是一堆零散实验。"
              ]
            },
            {
              "id": "first-success",
              "title": "什么叫第一次成功",
              "steps": [
                {
                  "title": "打开首页和文档页",
                  "desc": "确认 Web 已经通过同一套公开地址正常对外可访问。"
                },
                {
                  "title": "打开 Casdoor 登录页",
                  "desc": "确认认证服务可访问，品牌资源没有丢，地址没有跳错。"
                },
                {
                  "title": "完成登录并进入控制台",
                  "desc": "确认登录、回调、会话建立和控制台渲染这条链路已经通了。"
                },
                {
                  "title": "保存至少一个模型配置",
                  "desc": "这是后续创建助手和聊天成功的前提。"
                },
                {
                  "title": "创建第一个边界清晰的助手",
                  "desc": "先做一个明确任务，而不是一上来做万能助手。"
                },
                {
                  "title": "进入 Chat 发消息并拿到正确回复",
                  "desc": "这一步完成，才算真正从 0 到 1 跑通了产品。"
                }
              ]
            }
          ]
        },
        {
          "sectionSlug": "getting-started",
          "slug": "quick-start",
          "title": "快速上手",
          "summary": "按步骤把本地环境真正拉起来，并完成首次登录、首次模型保存、首次助手创建和首次聊天验证。",
          "seoTitle": "快速上手 | dotblue 产品文档",
          "seoDescription": "提供从零开始的 dotblue 安装路径，覆盖前置准备、环境变量、Compose 启动、首次验证和常见安装错误。",
          "readingTime": "14 分钟阅读",
          "sections": [
            {
              "id": "before-you-start",
              "title": "启动前准备",
              "bullets": [
                "准备可用的 Docker 和 Docker Compose 环境；如果你在 Windows 上开发，建议把 Compose 运行在 WSL 中，再用浏览器访问可达的主机 IP 或域名。",
                "提前确认浏览器实际访问地址，不要一边用 `localhost` 生成配置，一边又从另一台机器或宿主机 IP 打开页面。",
                "准备管理员账号信息和一组真实可用的 LLM API Key，否则系统虽然能启动，但你做不到完整验收。",
                "建议至少预留 4 CPU、8 GB 内存和足够磁盘空间给本地容器环境，否则首次构建和模型验证会明显不稳定。",
                "首次安装前先确认 19000、18080、18000 这类默认端口没有被别的程序占用。"
              ]
            },
            {
              "id": "decide-public-urls",
              "title": "先定公开地址，再生成配置",
              "paragraphs": [
                "dotblue 的首次安装最容易踩坑的点，不是某个容器没起来，而是 Web、Backend、Casdoor 使用了互相不一致的公开地址。地址一旦不一致，就会出现登录跳错域名、回调失败、品牌资源丢失、页面表面能开但流程不通等问题。",
                "所以第一原则是：你打算在浏览器里用什么地址打开系统，就把这个地址写进 `.env`，然后再执行 `prepare-config`。不要把这一步放在容器启动之后补救。"
              ],
              "code": {
                "language": "bash",
                "value": "# 示例 1：本机单机访问\nCASDOOR_PUBLIC_URL=http://localhost:18000\nDOTBLUE_PUBLIC_URL=http://localhost:19000\nDOTBLUE_BACKEND_PUBLIC_URL=http://localhost:18080\n\n# 示例 2：WSL 或局域网通过宿主机 IP 访问\nCASDOOR_PUBLIC_URL=http://172.22.3.181:18000\nDOTBLUE_PUBLIC_URL=http://172.22.3.181:19000\nDOTBLUE_BACKEND_PUBLIC_URL=http://172.22.3.181:18080"
              }
            },
            {
              "id": "minimum-env-values",
              "title": "最少要改哪些环境变量",
              "paragraphs": [
                "`deploy/compose/.env.example` 里变量不少，但第一次成功只需要先聚焦几类必填项：公开地址、管理员账号、数据库密码、模型供应商参数。其他像 S3、MinIO、运行时高级模式可以先不动。"
              ],
              "bullets": [
                "`CASDOOR_PUBLIC_URL`、`DOTBLUE_PUBLIC_URL`、`DOTBLUE_BACKEND_PUBLIC_URL`：决定登录回调、页面 API 调用和品牌资源引用。",
                "`DOTBLUE_ADMIN_*`：首个管理员账号。后续很多配置动作都依赖它。",
                "`CASDOOR_DB_PASSWORD`、`DOTBLUE_DB_PASSWORD`：数据库口令，至少不要保留默认占位值。",
                "`DOTBLUE_LLM_PROVIDER_TYPE`、`DOTBLUE_LLM_API_BASE`、`DOTBLUE_LLM_API_KEY`、`DOTBLUE_LLM_MODEL`：决定你是否能完成首次聊天。"
              ],
              "code": {
                "language": "bash",
                "value": "CASDOOR_PUBLIC_URL=http://172.22.3.181:18000\nDOTBLUE_PUBLIC_URL=http://172.22.3.181:19000\nDOTBLUE_BACKEND_PUBLIC_URL=http://172.22.3.181:18080\n\nCASDOOR_DB_PASSWORD=replace-with-casdoor-db-password\nDOTBLUE_DB_PASSWORD=replace-with-dotblue-db-password\n\nDOTBLUE_ADMIN_USERNAME=admin\nDOTBLUE_ADMIN_DISPLAY_NAME=Platform Admin\nDOTBLUE_ADMIN_EMAIL=admin@example.com\nDOTBLUE_ADMIN_PASSWORD=replace-with-strong-admin-password\n\nDOTBLUE_LLM_PROVIDER_TYPE=openai\nDOTBLUE_LLM_API_BASE=https://api.openai.com/v1\nDOTBLUE_LLM_API_KEY=replace-with-provider-key\nDOTBLUE_LLM_MODEL=gpt-4.1-mini"
              }
            },
            {
              "id": "launch-steps",
              "title": "一步一步拉起本地环境",
              "steps": [
                {
                  "title": "复制环境变量模板",
                  "desc": "进入 `deploy/compose`，把 `.env.example` 复制成 `.env`。"
                },
                {
                  "title": "修改 `.env` 中的公开地址、管理员账号和模型参数",
                  "desc": "先把真正影响首次成功的变量改对，再考虑高级参数。"
                },
                {
                  "title": "执行 `prepare-config`",
                  "desc": "Linux 或 WSL 用 `./prepare-config.sh`，Windows PowerShell 用 `.\\prepare-config.ps1`。这一步会生成 `.generated/` 下的运行时配置，并把 Casdoor OAuth 相关值写回 `.env`。"
                },
                {
                  "title": "执行 `docker compose up -d --build`",
                  "desc": "让 postgres、redis、casdoor、dotblue、web 这套最小栈一起启动。"
                },
                {
                  "title": "运行健康检查脚本",
                  "desc": "优先跑 `./smoke-test.sh`；如果 agent runtime 起不来，再跑 `runtime-doctor`。"
                }
              ],
              "code": {
                "language": "bash",
                "value": "cd deploy/compose\ncp .env.example .env\n# 编辑 .env\n./prepare-config.sh\ndocker compose up -d --build\n./smoke-test.sh"
              }
            },
            {
              "id": "windows-commands",
              "title": "Windows PowerShell 路径",
              "paragraphs": [
                "如果你在 Windows 下维护文件和命令，也可以直接用 PowerShell 生成配置；但如果容器实际跑在 WSL 中，请确保浏览器访问地址与 `.env` 中写入的地址一致。"
              ],
              "code": {
                "language": "powershell",
                "value": "cd deploy\\compose\nCopy-Item .env.example .env\n# 编辑 .env\n.\\prepare-config.ps1\ndocker compose up -d --build"
              }
            },
            {
              "id": "first-verification",
              "title": "首次验证按这个顺序做",
              "steps": [
                {
                  "title": "打开首页",
                  "desc": "确认 `DOTBLUE_PUBLIC_URL` 对应的产品页可以打开，静态资源没有 404。"
                },
                {
                  "title": "打开登录页并跳转到 Casdoor",
                  "desc": "确认 Casdoor 页面可达，品牌资源正常，地址没有跳去错误的主机名。"
                },
                {
                  "title": "完成注册或登录",
                  "desc": "本地默认是最简注册路径，先验证 Username / Display name / Password 这条链路跑通。"
                },
                {
                  "title": "完成回调并进入 Dashboard",
                  "desc": "确认 `/callback` 没有跨域名漂移，也不会回到错误地址。"
                },
                {
                  "title": "保存一个模型配置",
                  "desc": "没有模型，后面创建助手和聊天都会卡住。"
                },
                {
                  "title": "创建助手并在 Chat 中发出第一条消息",
                  "desc": "只有这一步成功，才算系统真正可用。"
                }
              ]
            },
            {
              "id": "what-success-looks-like",
              "title": "装成功之后你应该看到什么",
              "bullets": [
                "首页、文档页、登录页使用一致的品牌资源。",
                "Casdoor 登录完成后能稳定回到控制台，不会卡在回调页。",
                "控制台不空白，至少能看到可操作的管理界面。",
                "模型配置保存后，创建助手时可以看到可选模型。",
                "聊天页能选中目标助手，并在合理时间内返回第一条回复。"
              ]
            },
            {
              "id": "common-install-mistakes",
              "title": "快速安装最常见的错误",
              "bullets": [
                "`.env` 里写的是 `localhost`，浏览器却从宿主机 IP 或其他域名访问，导致回调和资源错位。",
                "修改了 `.env` 之后没有重新执行 `prepare-config`，结果 Casdoor 和前端还在吃旧配置。",
                "容器都启动了，但没有配置真实模型 Key，所以创建助手后聊天始终不成功。",
                "改了平台模型后没有重建旧的 runtime 容器，导致聊天结果还在用旧配置。",
                "浏览器缓存了旧前端资源，代码虽然已经发布，但页面看起来像没变化。"
              ]
            }
          ]
        },
        {
          "sectionSlug": "getting-started",
          "slug": "login-and-authentication",
          "title": "登录与认证",
          "summary": "理解本地默认认证路径、为什么注册被简化、首次登录后应该做什么，以及什么时候再开启更高级认证能力。",
          "seoTitle": "登录与认证 | dotblue 产品文档",
          "seoDescription": "讲清 dotblue 的 Casdoor 认证路径、本地最简注册逻辑、首次管理员操作和高级登录注册能力配置边界。",
          "readingTime": "9 分钟阅读",
          "sections": [
            {
              "id": "default-flow",
              "title": "当前默认认证路径",
              "paragraphs": [
                "本地快速拉起默认采用最简注册链路，注册页面保留 Username、Display name、Password、Confirm password 四项，目的就是把首次安装依赖压到最低。",
                "这样做不是在弱化认证，而是在本地验证阶段优先保证“注册成功 -> 登录成功 -> 回调成功 -> 进入产品”这条主链路稳定。"
              ]
            },
            {
              "id": "first-login-runbook",
              "title": "第一次登录建议怎么做",
              "steps": [
                {
                  "title": "先用管理员账号完成首次登录",
                  "desc": "不要第一步就尝试复杂多角色验证，先把管理员主链路跑通。"
                },
                {
                  "title": "确认回调后进入控制台",
                  "desc": "如果回到错误地址，优先检查 public URL 并重新生成配置。"
                },
                {
                  "title": "立刻检查模型配置页面",
                  "desc": "很多团队把问题误以为是登录问题，实际上下一步更常见的阻塞点是模型还没配。"
                },
                {
                  "title": "创建一个测试助手并去 Chat 验证",
                  "desc": "登录成功不是终点，要把认证路径与产品使用路径串起来。"
                }
              ]
            },
            {
              "id": "why-simplified",
              "title": "为什么默认不直接开邮箱和手机注册",
              "bullets": [
                "邮箱注册依赖 SMTP 服务、模板、发信可达性和验证码体验，首次安装时失败面太大。",
                "手机注册依赖短信供应商、模板审核、费用额度和失败处理，不适合作为本地默认路径。",
                "大多数团队在第一天真正需要的是稳定进入产品，而不是马上把完整身份体系做完。"
              ]
            },
            {
              "id": "advanced-options",
              "title": "什么时候再开启高级认证能力",
              "note": "邮箱验证码、短信验证码、WebAuthn、LDAP、社交登录、企业 SSO 都应该进入正式认证项目计划，而不是本地默认行为。",
              "bullets": [
                "当 SMTP 已接通并完成真实投递验证后，再开启邮箱验证码注册。",
                "当短信链路被纳入正式交付范围后，再开启手机验证码注册。",
                "当你准备进入测试环境或准生产环境时，再分阶段验证 LDAP、WebAuthn、企业 SSO 等能力。"
              ],
              "links": [
                {
                  "label": "Casdoor 注册项配置",
                  "url": "https://casdoor.ai/docs/application/signup-items-table",
                  "description": "配置注册字段、必填项、验证项和表单结构。"
                },
                {
                  "label": "Casdoor 登录方式配置",
                  "url": "https://casdoor.ai/docs/application/signin-methods",
                  "description": "配置 Password、验证码、WebAuthn、LDAP 等登录方式。"
                },
                {
                  "label": "Casdoor 应用认证配置",
                  "url": "https://casdoor.ai/docs/application/config",
                  "description": "查看回调地址、应用级认证设置和验证码策略。"
                },
                {
                  "label": "Casdoor 邮件服务配置",
                  "url": "https://casdoor.ai/docs/provider/email/overview",
                  "description": "配置 SMTP 邮件服务，用于邮箱验证码和找回密码。"
                }
              ]
            }
          ]
        }
      ]
    },
    {
      "slug": "use-dotblue",
      "title": "使用 dotblue",
      "description": "系统装起来之后，接下来是学会正确使用它：理解页面分工、工作空间边界、模型配置顺序、助手设计方法和聊天验收方式。",
      "articles": [
        {
          "sectionSlug": "use-dotblue",
          "slug": "assistants-and-workspaces",
          "title": "助手与工作空间",
          "summary": "讲清楚登录进入系统以后，你先看什么、先配什么，以及助手、组织、企业上下文之间到底是什么关系。",
          "seoTitle": "助手与工作空间 | dotblue 产品文档",
          "seoDescription": "帮助第一次接触 dotblue 的团队理解控制台结构、助手边界、组织隔离和首个助手设计方法。",
          "readingTime": "10 分钟阅读",
          "sections": [
            {
              "id": "after-login",
              "title": "第一次进入系统先看什么",
              "paragraphs": [
                "第一次登录成功后，不要急着到处点。正确顺序通常是：先确认控制台正常加载，再去看模型配置，然后创建助手，最后到聊天页做真实验证。",
                "如果你跳过模型配置直接去做助手，经常会得到“页面看起来没问题，但就是不能创建或不能聊天”的假成功状态。"
              ]
            },
            {
              "id": "assistant-model",
              "title": "为什么说助手是一个产品面",
              "paragraphs": [
                "在 dotblue 里，助手不是一个随手起的名字，而是一块真正面向用户的 AI 产品面。它应该有明确目标、清晰边界、自己的提示词和自己的行为预期。",
                "第一个助手做得好不好，不取决于模型参数堆得多复杂，而取决于范围是不是清晰、输出是不是可验证。"
              ]
            },
            {
              "id": "workspace-boundaries",
              "title": "工作空间、组织和平台配置怎么分工",
              "bullets": [
                "平台级设置适合放共享基础设施，例如默认模型供应商和全局能力。",
                "企业或组织上下文适合做团队、客户、租户之间的隔离。",
                "助手级配置适合描述具体业务行为，不应该用来承载跨团队共享的基础设置。",
                "如果你发现自己在每个助手里都重复维护同样的底层配置，往往说明应该把它上提到平台级或企业级。"
              ]
            },
            {
              "id": "role-based-guidance",
              "title": "不同角色第一次进入系统应该做什么",
              "steps": [
                {
                  "title": "平台管理员",
                  "desc": "先完成模型配置、管理员验证、基础品牌与公开地址确认。"
                },
                {
                  "title": "企业管理员或实施人员",
                  "desc": "确认组织边界、成员路径、目标业务助手范围，以及是否需要企业级隔离配置。"
                },
                {
                  "title": "普通业务用户",
                  "desc": "重点是能否选中正确助手并完成稳定聊天，而不是进入所有管理页面。"
                }
              ]
            },
            {
              "id": "first-assistant-guidance",
              "title": "第一个助手怎么做更容易成功",
              "steps": [
                {
                  "title": "先挑一个单点业务任务",
                  "desc": "例如客服问答、知识查询、销售线索初筛，而不是“公司万能助手”。"
                },
                {
                  "title": "写清楚助手边界",
                  "desc": "告诉它做什么、不做什么、什么时候拒答、答案要长什么样。"
                },
                {
                  "title": "先用少量高价值问题验证",
                  "desc": "通过真实问题看输出是否稳定，不要只看配置页面。"
                },
                {
                  "title": "确认通过后再扩范围",
                  "desc": "先做成，再做大，这是第一版成功率最高的方式。"
                }
              ]
            },
            {
              "id": "common-mistakes",
              "title": "第一次使用最容易犯的错",
              "bullets": [
                "把助手做得过宽，导致输出不可控，最后误以为系统不稳定。",
                "还没配模型就开始做助手和聊天验证。",
                "把组织隔离、平台配置和助手配置混在一起，后期维护越来越乱。",
                "只看列表页配置成功，不去聊天页做真实验收。"
              ]
            }
          ]
        },
        {
          "sectionSlug": "use-dotblue",
          "slug": "providers-and-models",
          "title": "模型供应商与模型配置",
          "summary": "把最常见的卡点讲清楚：模型在哪里配、先后顺序是什么、为什么助手页面经常只是把问题暴露出来。",
          "seoTitle": "模型供应商与模型配置 | dotblue 产品文档",
          "seoDescription": "面向新手解释 dotblue 的模型配置顺序、关键参数、首个模型配置模板和常见失败原因。",
          "readingTime": "10 分钟阅读",
          "sections": [
            {
              "id": "where-to-configure",
              "title": "模型应该先在哪里配置",
              "paragraphs": [
                "dotblue 假设模型在助手创建之前就已经可用。因此你看到“助手里没有模型可选”，通常不是助手页面坏了，而是平台级或企业级模型配置还没完成。",
                "正确理解顺序很重要：先有模型配置，再有助手；先有可用模型，再谈聊天体验。"
              ]
            },
            {
              "id": "minimum-provider-template",
              "title": "首个模型配置最少需要哪些信息",
              "bullets": [
                "供应商类型：例如 `openai`。",
                "API Base：必须是后端运行环境真正可达的地址。",
                "API Key：不能只在你本地脑子里有，必须真的进入运行环境。",
                "模型名：必须对应供应商真实可用的模型，而不是随便写的展示名。"
              ],
              "code": {
                "language": "bash",
                "value": "DOTBLUE_LLM_PROVIDER_TYPE=openai\nDOTBLUE_LLM_API_BASE=https://api.openai.com/v1\nDOTBLUE_LLM_API_KEY=replace-with-provider-key\nDOTBLUE_LLM_MODEL=gpt-4.1-mini"
              }
            },
            {
              "id": "save-order",
              "title": "建议的配置顺序",
              "steps": [
                {
                  "title": "先保存供应商与模型",
                  "desc": "确保平台或企业层已经有真实可用的模型记录。"
                },
                {
                  "title": "再回到助手配置页",
                  "desc": "确认模型已经能出现在助手配置里。"
                },
                {
                  "title": "最后去 Chat 验证",
                  "desc": "如果能回复，就说明从模型到运行时这条链路已经通了。"
                }
              ]
            },
            {
              "id": "provider-checklist",
              "title": "模型配置检查清单",
              "bullets": [
                "供应商类型和实际 API 协议一致。",
                "API Base 对后端容器可达，而不只是对浏览器可达。",
                "API Key 已注入当前实际运行环境。",
                "模型名拼写正确且权限可用。",
                "重大配置变更后，旧 runtime 已经刷新或重建。"
              ]
            },
            {
              "id": "failure-patterns",
              "title": "最常见的失败表现",
              "bullets": [
                "配置里能看到模型，助手里却选不到：通常是保存层级不对，或者界面还在读旧状态。",
                "助手能创建，聊天却没回复：通常是 API Key、API Base 或模型名不对。",
                "之前能用，改完就坏：很可能旧 runtime 还在吃旧配置。",
                "不同环境表现不一致：通常是测试环境和生产环境的模型参数没有真正对齐。"
              ]
            }
          ]
        },
        {
          "sectionSlug": "use-dotblue",
          "slug": "chat-and-operations",
          "title": "聊天页与日常运维",
          "summary": "把聊天页当作真正的验收面和运维面，知道每天该看什么、出错时先查什么。",
          "seoTitle": "聊天页与日常运维 | dotblue 产品文档",
          "seoDescription": "说明为什么 Chat 是 dotblue 的关键验收入口，并给出首次联调、日常检查和问题分诊路径。",
          "readingTime": "9 分钟阅读",
          "sections": [
            {
              "id": "chat-role",
              "title": "为什么 Chat 才是真正的验收入口",
              "paragraphs": [
                "Chat 是多个能力汇合的地方：认证成功不成功、助手边界清不清、模型可不可用、运行时是否更新、用户最终体验如何，都会在这里暴露。",
                "所以真正可靠的验收不是“我能打开页面”，而是“我能用正确助手稳定完成一段对话”。"
              ]
            },
            {
              "id": "first-acceptance-script",
              "title": "首次联调建议这样验收",
              "steps": [
                {
                  "title": "创建一个新会话",
                  "desc": "确认页面状态和会话状态是正常的。"
                },
                {
                  "title": "选中目标助手",
                  "desc": "确保你测试的是正确的助手，而不是默认空状态。"
                },
                {
                  "title": "输入一个短而可预测的问题",
                  "desc": "先验证稳定性，不要第一条就上复杂业务任务。"
                },
                {
                  "title": "观察返回时间和结果格式",
                  "desc": "看它是否在合理时间内返回，并符合你给助手设定的输出边界。"
                },
                {
                  "title": "再测两三个高价值问题",
                  "desc": "判断这个助手是否具备进入下一轮迭代的基础。"
                }
              ]
            },
            {
              "id": "daily-checks",
              "title": "日常运维最值得盯的点",
              "bullets": [
                "是否可以稳定创建新会话。",
                "目标助手是否持续可见且可选。",
                "首条回复时间是否在预期范围内。",
                "模型改动之后，行为是否已经真正刷新。",
                "出错时是否能快速判断是认证问题、模型问题还是运行时问题。"
              ]
            },
            {
              "id": "support-playbook",
              "title": "出问题时先按这个路径查",
              "steps": [
                {
                  "title": "先用简单问题复现",
                  "desc": "不要用模糊长任务制造额外噪音。"
                },
                {
                  "title": "检查当前助手对应的模型",
                  "desc": "先排除模型不可用或配置不一致。"
                },
                {
                  "title": "检查 runtime 是否陈旧",
                  "desc": "刚改完模型或技能配置时，旧 runtime 很可能仍在工作。"
                },
                {
                  "title": "再回头检查认证和会话",
                  "desc": "如果页面状态异常或上下文丢失，再回看登录回调和会话连续性。"
                }
              ]
            },
            {
              "id": "symptom-reading",
              "title": "按现象快速判断方向",
              "bullets": [
                "页面能开但无法登录：优先看 Casdoor 可达性和公开地址配置。",
                "能登录但控制台空白：优先看后端初始化和数据库连接。",
                "能进控制台但没模型：优先看平台级模型配置。",
                "能配助手但聊天不回复：优先看模型参数和 runtime 是否刷新。"
              ]
            }
          ]
        }
      ]
    },
    {
      "slug": "advanced",
      "title": "高级主题",
      "description": "当你准备从“本地能跑”走向“测试可交付”或“生产可上线”时，这部分会告诉你架构、发布、安全、备份和排障到底该怎么做。",
      "articles": [
        {
          "sectionSlug": "advanced",
          "slug": "deployment-architecture",
          "title": "部署架构",
          "summary": "看清 dotblue 最小栈由什么组成、不同服务分别负责什么，以及公开地址和生成配置为什么必须一致。",
          "seoTitle": "部署架构 | dotblue 产品文档",
          "seoDescription": "解释 dotblue 的最小服务栈、公开 URL 策略、持久化对象和生成配置在部署架构中的作用。",
          "readingTime": "11 分钟阅读",
          "sections": [
            {
              "id": "minimal-stack",
              "title": "最小可用服务栈",
              "paragraphs": [
                "一个真实可用的 dotblue 最小部署通常至少包含 postgres、redis、casdoor、dotblue、web。它们分别负责数据持久化、会话与队列支撑、认证、后端 API 与运行控制、以及浏览器端产品页面。",
                "如果其中任意一个服务只是“名义上存在”但没有和公开地址、配置生成链对齐，整个产品就会变成看似能打开、实际不闭环的状态。"
              ],
              "code": {
                "language": "text",
                "value": "Services\n- postgres: 数据库\n- redis: 会话 / 队列 / 事件支撑\n- casdoor: 认证服务\n- dotblue: 后端 API 与运行治理\n- web: 前端产品页\n\nBrowser-facing ports\n- Web: 19000\n- Backend: 18080\n- Casdoor: 18000"
              }
            },
            {
              "id": "public-url-matrix",
              "title": "公开地址必须这样理解",
              "bullets": [
                "`DOTBLUE_PUBLIC_URL` 是用户最终打开首页、文档页、登录入口和控制台的地址。",
                "`CASDOOR_PUBLIC_URL` 是用户浏览器真正访问登录页和完成认证跳转的地址。",
                "`DOTBLUE_BACKEND_PUBLIC_URL` 是前端浏览器调用后端 API 时所面向的公开 API 地址。",
                "内部容器地址只用于服务间通信，不能拿来替代浏览器访问地址。"
              ],
              "code": {
                "language": "text",
                "value": "Public URL matrix\n- App URL: user opens dotblue web here\n- Auth URL: user is redirected to Casdoor here\n- API URL: browser requests backend here\n\nInternal URL matrix\n- dotblue -> postgres / redis / casdoor by container network\n- web static container does not replace browser-facing public URLs"
              }
            },
            {
              "id": "generated-config",
              "title": "为什么生成配置是架构的一部分",
              "paragraphs": [
                "`prepare-config` 生成的不是临时垃圾文件，而是服务对齐的核心机制。Casdoor 的应用配置、品牌资源 URL、回调地址、前端 `VITE_*` 变量和后端运行配置，都是通过它串起来的。",
                "如果你修改了公开地址、品牌资源、回调策略或 OAuth 客户端相关信息，却没有重新生成配置，那么你测试到的其实不是当前环境，而是混合了旧配置的环境。"
              ]
            },
            {
              "id": "persistent-data",
              "title": "哪些数据必须持久化",
              "bullets": [
                "Postgres：核心业务数据、用户、配置和关系数据。",
                "文件存储目录或对象存储：聊天附件与相关文件资产。",
                "agent runtime 数据目录：如果你的交付依赖运行时落盘，需要明确它的宿主路径和备份策略。",
                "Casdoor 初始化与应用配置：至少要把真正生效的生成结果纳入发布记录。"
              ]
            },
            {
              "id": "from-local-to-production",
              "title": "从本地到生产，架构思路怎么变",
              "paragraphs": [
                "本地环境追求的是最快跑通；生产环境追求的是稳定域名、密钥治理、持久化数据、备份恢复和可观测性。因此你不能简单把“本地 compose 能跑”视为生产方案已经成立。",
                "进入正式环境时，至少要重新审视域名、TLS、代理层、Secret 注入、备份策略、日志留存和发布回滚策略。"
              ]
            }
          ]
        },
        {
          "sectionSlug": "advanced",
          "slug": "production-rollout",
          "title": "生产上线",
          "summary": "基于“数据库、Redis 等基础组件已经存在”的前提，按 Casdoor、dotblue、web 三个服务分别部署，并写清配置文件、初始化数据、端口、访问地址与 Nginx 代理。",
          "seoTitle": "生产上线 | dotblue 产品文档",
          "seoDescription": "提供 dotblue 的逐服务生产部署指南，覆盖 Casdoor、dotblue、web 的配置文件、初始化数据、Docker 启动方式、OSS/S3 配置、访问端口和 Nginx 代理。",
          "readingTime": "22 分钟阅读",
          "sections": [
            {
              "id": "scope",
              "title": "这份部署说明适用什么场景",
              "paragraphs": [
                "这份说明假设 PostgreSQL、Redis 等基础组件已经存在，你现在要做的是把 Casdoor、dotblue backend、web 三个服务作为正式服务独立部署起来，而不是用 Docker Compose 一次性把所有组件拉起。",
                "默认落地方式是：Casdoor 使用官方镜像单独启动；dotblue backend 使用仓库内 Dockerfile 构建镜像单独启动；web 使用仓库内 Dockerfile 构建静态站点镜像；最外层由 Nginx 负责正式域名、TLS 和反向代理。"
              ]
            },
            {
              "id": "target-topology",
              "title": "生产环境目标形态",
              "paragraphs": [
                "生产环境最终应形成三条稳定访问面：`app.example.com` 给用户访问前端，`api.example.com` 给前端调用后端 API，`auth.example.com` 给用户完成 Casdoor 登录与回调。",
                "Casdoor、dotblue backend、web 各自独立启动，各自监听内部端口；数据库、Redis、文件存储作为底层依赖存在；Nginx 只负责把正式域名反代到对应内部端口。"
              ],
              "code": {
                "language": "text",
                "value": "Recommended production shape\n- app.example.com -> web container port 80\n- api.example.com -> dotblue backend port 8000\n- auth.example.com -> Casdoor port 8000\n- postgres -> existing production PostgreSQL\n- redis -> existing production Redis\n- file storage -> local durable path or S3-compatible object storage\n- nginx -> TLS termination and reverse proxy"
              }
            },
            {
              "id": "artifacts",
              "title": "部署前要准备好的文件和信息",
              "bullets": [
                "三个正式域名：例如 `app.example.com`、`api.example.com`、`auth.example.com`。",
                "一台可访问 PostgreSQL 和 Redis 的 Linux 服务器，建议至少 4 CPU、8 GB 内存。",
                "TLS 证书，或者由 Nginx 自动签发证书的方案。",
                "Casdoor 模板文件：`deploy/casdoor/app.conf.example` 和 `deploy/casdoor/init_data.example.json`。",
                "dotblue backend 模板文件：`backend/manifest/config/config.example.yaml` 和 `backend/manifest/config/init_data.example.json`。",
                "web 构建参数来源：`web/.env.example` 与 `web/Dockerfile`。",
                "管理员账号、LLM API Key、文件存储方案，以及数据库和 Redis 的连接地址。"
              ]
            },
            {
              "id": "service-ports",
              "title": "三个服务各自监听什么端口",
              "bullets": [
                "Casdoor 容器内部默认监听 `8000`。",
                "dotblue backend 默认监听 `8000`。",
                "web 容器内部默认监听 `80`。",
                "生产环境里建议只把这三个服务绑定到 `127.0.0.1`，例如 `127.0.0.1:18000 -> casdoor:8000`、`127.0.0.1:18080 -> backend:8000`、`127.0.0.1:19000 -> web:80`，再由 Nginx 对外代理正式域名。"
              ]
            },
            {
              "id": "casdoor-files",
              "title": "先部署 Casdoor：需要哪些文件，哪些必须改",
              "paragraphs": [
                "Casdoor 最关键的两个文件是 `deploy/casdoor/app.conf.example` 和 `deploy/casdoor/init_data.example.json`。前者控制 Casdoor 自己如何连接数据库和对外暴露；后者控制组织、应用、管理员、回调地址、品牌资源等初始化数据。",
                "Casdoor 官方镜像是 `casbin/casdoor:latest`，启动时只要把改好的 `app.conf` 和 `init_data.json` 挂进容器即可。"
              ],
              "bullets": [
                "`app.conf` 里必须改：`dataSourceName`、`dbName`、`origin`、`originFrontend`、`defaultApplication`、`radiusDefaultOrganization`。",
                "`init_data.json` 里必须改：组织名、应用名、`homepageUrl`、`redirectUris`、`clientId`、`clientSecret`、品牌资源 URL、管理员用户名 / 邮箱 / 密码。",
                "如果正式品牌资源放在 `app.example.com/brand/...`，Casdoor 初始化数据里的 `logo`、`favicon`、`formSideHtml`、`formBackgroundUrl` 都要一起替换成正式地址。"
              ],
              "code": {
                "language": "ini",
                "value": "app.conf 关键项示例\nrunmode = prod\ndriverName = postgres\ndataSourceName = \"user=casdoor password=replace-with-casdoor-db-password host=10.0.0.21 port=5432 sslmode=disable dbname=casdoor\"\ndbName = casdoor\norigin = https://auth.example.com\noriginFrontend = https://auth.example.com\ndefaultApplication = dotblue\nradiusDefaultOrganization = dotblue\ninitDataFile = /init_data.json"
              }
            },
            {
              "id": "casdoor-init-data",
              "title": "Casdoor `init_data.json` 里最关键的字段",
              "bullets": [
                "`organizations[0].name`：建议与 dotblue 使用的组织名保持一致，例如 `dotblue`。",
                "`applications[0].name`：建议设为 `dotblue`。",
                "`applications[0].homepageUrl`：应为 `https://app.example.com`。",
                "`applications[0].redirectUris`：至少包含 `https://app.example.com/callback`。",
                "`applications[0].clientId`、`clientSecret`：后续要写回 dotblue backend 配置和 web 构建参数。",
                "`users[0]`：Casdoor 超管账号信息，生产环境必须改成正式用户名、邮箱和强密码。"
              ]
            },
            {
              "id": "casdoor-run",
              "title": "Casdoor 启动命令示例",
              "paragraphs": [
                "下面示例假设你已经把修改后的 `app.conf` 和 `init_data.json` 放在服务器目录 `/opt/dotblue/casdoor/` 下。"
              ],
              "code": {
                "language": "bash",
                "value": "docker run -d \\\n  --name casdoor \\\n  --restart unless-stopped \\\n  -p 127.0.0.1:18000:8000 \\\n  -v /opt/dotblue/casdoor/app.conf:/conf/app.conf:ro \\\n  -v /opt/dotblue/casdoor/init_data.json:/init_data.json:ro \\\n  -v /opt/dotblue/casdoor/logs:/logs \\\n  casbin/casdoor:latest"
              }
            },
            {
              "id": "dotblue-files",
              "title": "再部署 dotblue backend：配置文件与初始数据怎么分工",
              "paragraphs": [
                "dotblue backend 最关键的是两个文件：`config.yaml` 负责运行时连接和存储配置，`init_data.json` 负责首次安装时写入组织、管理员、平台与默认模型供应商信息。",
                "根据后端 README 的约定，服务启动时默认会读 `manifest/config/config.yaml`；如果存在 `manifest/config/init_data.json`，或者通过 `DOTBLUE_INIT_DATA_PATH` 指定初始化文件，就会自动执行首次安装。"
              ],
              "bullets": [
                "`config.yaml` 里必须改：数据库连接、Redis 地址、Casdoor endpoint、Casdoor clientId / clientSecret / jwtSecret、文件存储、engine 路径与模式。",
                "`init_data.json` 里必须改：组织名、管理员用户名 / 邮箱 / 密码、平台 runtime 路径、默认模型供应商参数。",
                "`init_data.json` 中的 `organization.name` 必须与 `config.yaml` 中的 `casdoor.organizationName` 保持一致；`application.name` 必须与 `casdoor.applicationName` 保持一致。"
              ]
            },
            {
              "id": "dotblue-config-example",
              "title": "dotblue `config.yaml` 关键项示例",
              "code": {
                "language": "yaml",
                "value": "server:\n  address: \":8000\"\n\ndatabase:\n  default:\n    link: \"pgsql:dotblue:replace-with-dotblue-db-password@tcp(10.0.0.21:5432)/dotblue\"\n    debug: false\n\ncasdoor:\n  endpoint: \"http://10.0.0.31:18000\"\n  clientId: \"replace-with-runtime-client-id\"\n  clientSecret: \"replace-with-runtime-client-secret\"\n  jwtSecret: |\n    -----BEGIN CERTIFICATE-----\n    replace-with-casdoor-application-certificate\n    -----END CERTIFICATE-----\n  organizationName: \"dotblue\"\n  applicationName: \"dotblue\"\n  bootstrap:\n    endpoint: \"\"\n    clientId: \"\"\n    clientSecret: \"\"\n    jwtSecret: \"\"\n\nfiles:\n  driver: \"local\"\n  localRoot: \"/var/lib/dotblue/chat-files\"\n\nredis:\n  address: \"10.0.0.22:6379\"\n  password: \"\"\n  db: 0\n  keyPrefix: \"dot\"\n\nworker:\n  embedded: true\n\nengine:\n  dataBasePath: \"/var/lib/dotblue/agents\"\n  dataMountPath: \"/var/lib/dotblue/agents\"\n  containerPort: 8642\n  runtimeMode: \"host\"\n  endpointMode: \"host_loopback\"\n  dockerEndpoint: \"unix:///var/run/docker.sock\"\n  dockerNetwork: \"\""
              }
            },
            {
              "id": "oss-config",
              "title": "如果文件存储用 OSS，该怎么改",
              "paragraphs": [
                "当前仓库只有 `local` 和 `s3` 两种文件驱动，没有单独的 `oss` 驱动。所以如果你要接阿里云 OSS，应按 S3 兼容方式配置，而不是写 `driver: oss`。",
                "也就是说：`files.driver` 仍然要写成 `s3`，然后把 `endpoint`、`bucket`、`accessKey`、`secretKey`、`region` 等参数改成你的 OSS 兼容参数。"
              ],
              "code": {
                "language": "yaml",
                "value": "files:\n  driver: \"s3\"\n  localRoot: \"/var/lib/dotblue/chat-files\"\n  s3:\n    endpoint: \"https://oss-cn-hangzhou.aliyuncs.com\"\n    region: \"cn-hangzhou\"\n    bucket: \"your-dotblue-bucket\"\n    accessKey: \"replace-with-oss-access-key\"\n    secretKey: \"replace-with-oss-secret-key\"\n    sessionToken: \"\"\n    forcePathStyle: false\n    autoCreateBucket: false"
              }
            },
            {
              "id": "dotblue-bootstrap",
              "title": "什么时候需要 `casdoor.bootstrap.*`",
              "bullets": [
                "默认情况下，dotblue 会直接复用 `casdoor.*` 配置做初始化。",
                "如果你提供给 dotblue 运行期使用的 Casdoor 应用，只具备登录校验能力，不具备创建 Organization / Application / User 的权限，那么首次安装时就需要额外配置 `casdoor.bootstrap.endpoint`、`clientId`、`clientSecret`、`jwtSecret`。",
                "这样初始化过程使用 bootstrap 应用执行，运行期登录仍然使用主 `casdoor.*` 配置。"
              ]
            },
            {
              "id": "dotblue-init-data",
              "title": "dotblue `init_data.json` 关键项示例",
              "code": {
                "language": "json",
                "value": "{\n  \"version\": 1,\n  \"organization\": {\n    \"name\": \"dotblue\",\n    \"displayName\": \"dotblue\"\n  },\n  \"application\": {\n    \"name\": \"dotblue\",\n    \"displayName\": \"dotblue\",\n    \"homepageUrl\": \"https://app.example.com\",\n    \"redirectUris\": [\"https://app.example.com/callback\"]\n  },\n  \"admin\": {\n    \"username\": \"admin\",\n    \"displayName\": \"Platform Admin\",\n    \"email\": \"admin@example.com\",\n    \"passwordEnv\": \"DOTBLUE_ADMIN_PASSWORD\"\n  },\n  \"platform\": {\n    \"dataBasePath\": \"/var/lib/dotblue/agents\",\n    \"dataMountPath\": \"/var/lib/dotblue/agents\",\n    \"containerPort\": 8642,\n    \"runtimeMode\": \"host\",\n    \"endpointMode\": \"host_loopback\",\n    \"dockerEndpoint\": \"unix:///var/run/docker.sock\",\n    \"dockerNetwork\": \"\"\n  },\n  \"provider\": {\n    \"type\": \"openai\",\n    \"apiBase\": \"https://api.openai.com/v1\",\n    \"apiKeyEnv\": \"DOTBLUE_LLM_API_KEY\",\n    \"model\": \"gpt-4.1-mini\"\n  }\n}"
              }
            },
            {
              "id": "dotblue-run",
              "title": "dotblue backend 构建与启动命令示例",
              "paragraphs": [
                "下面示例假设你把 `config.yaml` 和 `init_data.json` 放到服务器目录 `/opt/dotblue/backend-config/`，并把运行时目录放到 `/var/lib/dotblue/`。",
                "后端镜像来自仓库里的 `backend/Dockerfile`，容器内部默认执行 `./main`，内部监听 `8000`。"
              ],
              "code": {
                "language": "bash",
                "value": "docker build -t dotblue-backend:prod /opt/src/dotblue/backend\n\ndocker run -d \\\n  --name dotblue-backend \\\n  --restart unless-stopped \\\n  -p 127.0.0.1:18080:8000 \\\n  -e DOTBLUE_ADMIN_PASSWORD='replace-with-strong-admin-password' \\\n  -e DOTBLUE_LLM_API_KEY='replace-with-real-provider-key' \\\n  -v /opt/dotblue/backend-config/config.yaml:/app/manifest/config/config.yaml:ro \\\n  -v /opt/dotblue/backend-config/init_data.json:/app/manifest/config/init_data.json:ro \\\n  -v /var/lib/dotblue/chat-files:/var/lib/dotblue/chat-files \\\n  -v /var/lib/dotblue/agents:/var/lib/dotblue/agents \\\n  -v /var/run/docker.sock:/var/run/docker.sock \\\n  dotblue-backend:prod"
              }
            },
            {
              "id": "web-build",
              "title": "最后部署 web：镜像可复用，配置在启动时注入",
              "paragraphs": [
                "web 镜像现在会在容器启动时生成 `/runtime-config.js`，再由浏览器读取 `VITE_CASDOOR_SERVER_URL`、`VITE_CASDOOR_CLIENT_ID`、`VITE_CASDOOR_ORG_NAME`、`VITE_CASDOOR_APP_NAME`、`VITE_BACKEND_URL` 这些运行时配置。",
                "也就是说，私有化部署时可以复用同一个 `dotblue-web` 镜像，只改容器环境变量，不需要为每个环境重新构建前端镜像。"
              ],
              "code": {
                "language": "bash",
                "value": "docker build -t dotblue-web:prod /opt/src/dotblue/web"
              }
            },
            {
              "id": "web-run",
              "title": "web 启动命令示例",
              "code": {
                "language": "bash",
                "value": "docker run -d \\\n  --name dotblue-web \\\n  --restart unless-stopped \\\n  -p 127.0.0.1:19000:80 \\\n  -e VITE_CASDOOR_SERVER_URL=https://auth.example.com \\\n  -e VITE_CASDOOR_CLIENT_ID=replace-with-runtime-client-id \\\n  -e VITE_CASDOOR_ORG_NAME=dotblue \\\n  -e VITE_CASDOOR_APP_NAME=dotblue \\\n  -e VITE_BACKEND_URL=https://api.example.com \\\n  dotblue-web:prod"
              }
            },
            {
              "id": "nginx-proxy",
              "title": "Nginx 代理示例",
              "paragraphs": [
                "建议只让 Nginx 对外监听 80/443，Casdoor、dotblue backend、web 都只绑定到 `127.0.0.1` 的内部端口。",
                "然后把 `auth.example.com` 反代到 `127.0.0.1:18000`，把 `api.example.com` 反代到 `127.0.0.1:18080`，把 `app.example.com` 反代到 `127.0.0.1:19000`。"
              ],
              "code": {
                "language": "nginx",
                "value": "server {\n  listen 443 ssl http2;\n  server_name app.example.com;\n\n  ssl_certificate /etc/letsencrypt/live/app.example.com/fullchain.pem;\n  ssl_certificate_key /etc/letsencrypt/live/app.example.com/privkey.pem;\n\n  location / {\n    proxy_pass http://127.0.0.1:19000;\n    proxy_set_header Host $host;\n    proxy_set_header X-Forwarded-Proto https;\n    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n    proxy_set_header X-Real-IP $remote_addr;\n  }\n}\n\nserver {\n  listen 443 ssl http2;\n  server_name api.example.com;\n\n  ssl_certificate /etc/letsencrypt/live/api.example.com/fullchain.pem;\n  ssl_certificate_key /etc/letsencrypt/live/api.example.com/privkey.pem;\n\n  location / {\n    proxy_pass http://127.0.0.1:18080;\n    proxy_set_header Host $host;\n    proxy_set_header X-Forwarded-Proto https;\n    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n    proxy_set_header X-Real-IP $remote_addr;\n  }\n}\n\nserver {\n  listen 443 ssl http2;\n  server_name auth.example.com;\n\n  ssl_certificate /etc/letsencrypt/live/auth.example.com/fullchain.pem;\n  ssl_certificate_key /etc/letsencrypt/live/auth.example.com/privkey.pem;\n\n  location / {\n    proxy_pass http://127.0.0.1:18000;\n    proxy_set_header Host $host;\n    proxy_set_header X-Forwarded-Proto https;\n    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n    proxy_set_header X-Real-IP $remote_addr;\n  }\n}"
              }
            },
            {
              "id": "deployment-order",
              "title": "推荐的正式部署顺序",
              "steps": [
                {
                  "title": "先准备 Casdoor 配置并启动 Casdoor",
                  "desc": "确保 `auth.example.com` 已能正常打开 Casdoor 首页。"
                },
                {
                  "title": "再准备 dotblue `config.yaml` 和 `init_data.json`",
                  "desc": "确认 Casdoor endpoint、clientId、clientSecret、jwtSecret 和管理员密码环境变量全部可用。"
                },
                {
                  "title": "启动 dotblue backend",
                  "desc": "确认 `api.example.com` 已能被 Nginx 代理到 backend。"
                },
                {
                  "title": "按正式域名构建并启动 web",
                  "desc": "确认 `VITE_*` 构建参数都是正式值，再上线 web。"
                },
                {
                  "title": "最后统一验证登录、回调、控制台和聊天",
                  "desc": "确保三个服务的正式域名和回调已经完全串起来。"
                }
              ]
            },
            {
              "id": "secrets-and-persistence",
              "title": "超管账号密码和关键信息放哪",
              "bullets": [
                "Casdoor 初始化管理员密码在 `deploy/casdoor/init_data.example.json` 的 `users` 段中配置。",
                "dotblue 平台管理员密码在 `backend/manifest/config/init_data.example.json` 的 `admin.passwordEnv` 或部署模板的 `admin.password` 中配置。",
                "LLM API Key 不要写死到仓库，推荐通过环境变量如 `DOTBLUE_LLM_API_KEY` 注入 backend 容器。",
                "Casdoor `clientId` / `clientSecret`、应用证书、backend 运行期 `jwtSecret` 都属于必须留档的核心机密。",
                "数据库、文件存储、运行时目录和当前生效配置文件应纳入备份与版本记录。"
              ]
            },
            {
              "id": "acceptance-runbook",
              "title": "第一次上线后的验收步骤",
              "steps": [
                {
                  "title": "先打开 `https://app.example.com`",
                  "desc": "确认前端首页和静态资源正常。"
                },
                {
                  "title": "从登录页进入 Casdoor",
                  "desc": "确认浏览器跳到 `https://auth.example.com`，页面样式和品牌资源正常。"
                },
                {
                  "title": "用管理员账号完成登录",
                  "desc": "确认回调回到 `https://app.example.com/callback` 并最终进入控制台。"
                },
                {
                  "title": "检查控制台和模型配置",
                  "desc": "确认 backend 能正常返回数据，并且至少有一组可用模型。"
                },
                {
                  "title": "创建一个测试助手",
                  "desc": "先用简单边界的助手验收，不要直接上复杂生产业务助手。"
                },
                {
                  "title": "在 Chat 中发一条简单消息",
                  "desc": "确认能得到第一条回复，说明 Casdoor、backend、模型和前端已经串通。"
                }
              ]
            },
            {
              "id": "ops-commands",
              "title": "上线后建议立刻执行的运维命令",
              "code": {
                "language": "bash",
                "value": "docker ps --filter name=casdoor\ndocker ps --filter name=dotblue-backend\ndocker ps --filter name=dotblue-web\n\ndocker logs --tail=200 casdoor\ndocker logs --tail=200 dotblue-backend\ndocker logs --tail=200 dotblue-web\n\ncurl -I http://127.0.0.1:18000\ncurl -I http://127.0.0.1:18080\ncurl -I http://127.0.0.1:19000\n\ncurl -I https://auth.example.com\ncurl -I https://api.example.com\ncurl -I https://app.example.com"
              }
            },
            {
              "id": "upgrade-and-rollback",
              "title": "升级和回滚怎么做更稳",
              "steps": [
                {
                  "title": "升级前先备份数据库和文件存储",
                  "desc": "不要在无备份情况下直接替换生产配置或镜像。"
                },
                {
                  "title": "保留上一版镜像和配置文件副本",
                  "desc": "Casdoor 的 `app.conf/init_data.json`、dotblue 的 `config.yaml/init_data.json`、web 构建参数都要留档。"
                },
                {
                  "title": "升级时先替换配置，再替换镜像",
                  "desc": "避免代码和配置版本错位。"
                },
                {
                  "title": "回滚时镜像和配置一起回退",
                  "desc": "不要只回退容器镜像，却保留新版 Casdoor 或 dotblue 配置。"
                },
                {
                  "title": "回滚后重新做最小验收",
                  "desc": "至少重验首页、登录、回调、控制台和第一条聊天消息。"
                }
              ]
            },
            {
              "id": "backup-and-monitoring",
              "title": "备份、监控和告警",
              "bullets": [
                "至少为 PostgreSQL 制定备份和恢复流程，并定期演练。",
                "Casdoor 日志、backend 日志和 Nginx 访问日志要保留。",
                "监控登录失败、API 错误、首条回复时延和运行时异常。",
                "把当前正式生效的配置文件版本和镜像版本纳入发布记录。"
              ]
            },
            {
              "id": "go-live-checklist",
              "title": "按这个做最终上线验收",
              "bullets": [
                "Casdoor 已独立可访问，组织、应用、管理员初始化正确。",
                "dotblue backend 已正常连接数据库、Redis 和 Casdoor。",
                "web 构建时使用的 `VITE_*` 参数都是正式域名和正式 clientId。",
                "Nginx 已正确把 `app`、`api`、`auth` 三个域名代理到对应内部端口。",
                "管理员可以完整完成登录、回调、进入控制台、创建助手和首条聊天验证。",
                "文件存储方案、管理员密码、LLM Key、Casdoor clientSecret、证书都已妥善保管。",
                "当前生效配置文件、镜像版本、备份时间点和回滚方案都已留档。"
              ]
            }
          ]
        },        {
          "sectionSlug": "advanced",
          "slug": "troubleshooting-and-ops",
          "title": "排障与运维",
          "summary": "按照问题现象定位：启动失败、登录失败、回调错乱、模型缺失、聊天异常、品牌不一致分别先查什么。",
          "seoTitle": "排障与运维 | dotblue 产品文档",
          "seoDescription": "提供面向真实问题现象的 dotblue 排障路径，覆盖安装、认证、模型、运行时和生产运维问题。",
          "readingTime": "12 分钟阅读",
          "sections": [
            {
              "id": "startup-issues",
              "title": "安装或启动阶段的问题怎么查",
              "steps": [
                {
                  "title": "容器没起来",
                  "desc": "先看 `docker compose ps` 和构建日志，判断是镜像构建失败、端口冲突还是环境变量缺失。"
                },
                {
                  "title": "页面打不开",
                  "desc": "先确认 `DOTBLUE_PUBLIC_URL` 对应地址真的可从当前浏览器访问，而不是只在容器内部可达。"
                },
                {
                  "title": "agent runtime 起不来",
                  "desc": "优先跑 `runtime-doctor`，检查 docker.sock 权限、runtime mode 和网络连通性。"
                }
              ]
            },
            {
              "id": "auth-issues",
              "title": "登录、回调和品牌问题",
              "bullets": [
                "登录页跳去错误地址：先看 `CASDOOR_PUBLIC_URL` 和 `DOTBLUE_PUBLIC_URL` 是否一致地使用了真实访问地址。",
                "回调成功却回不到控制台：优先重新执行 `prepare-config`，确认 Casdoor OAuth 回调地址已经更新。",
                "品牌资源没有更新：确认当前运行配置引用的是新资源，同时排除浏览器缓存。",
                "能看到登录页但会话不稳定：检查浏览器访问域名、回调路径和前端 token 处理链路。"
              ]
            },
            {
              "id": "model-and-runtime-issues",
              "title": "模型与运行时问题",
              "bullets": [
                "创建助手时没有模型：先检查平台级或企业级模型是否已保存。",
                "聊天页没有回复：优先检查 API Key、API Base、模型名以及网络可达性。",
                "改完模型仍表现旧行为：重建旧 runtime 容器后再测。",
                "不同助手结果差异很大：检查是否实际上使用了不同模型或不同系统提示词。"
              ]
            },
            {
              "id": "production-ops-issues",
              "title": "生产环境特别容易漏掉的点",
              "bullets": [
                "域名已经换成正式地址，但没有重新生成配置。",
                "TLS 已加，但反向代理没有正确转发 `X-Forwarded-*`。",
                "数据库有备份，但文件存储和运行时数据没有恢复演练。",
                "发布了新前端资源，但登录和回调仍指向旧 Casdoor 配置。"
              ]
            },
            {
              "id": "ops-checklist",
              "title": "建议长期保留的运维检查单",
              "bullets": [
                "首页、文档页、登录页、控制台、聊天页品牌一致。",
                "登录、回调、注册、退出登录都在同一套公开地址策略下工作正常。",
                "模型变更后有明确的 runtime 刷新动作。",
                "上线后持续观察认证失败、API 错误、聊天成功率和环境漂移。"
              ]
            }
          ]
        }
      ]
    }
  ]
};

