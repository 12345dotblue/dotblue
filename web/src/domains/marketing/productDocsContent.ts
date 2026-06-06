import { resolveSupportedLanguage } from '../../i18n/config';

export type DocsLocale = {
  seoTitle: string;
  seoDescription: string;
  seoKeywords: string;
  eyebrow: string;
  title: string;
  subtitle: string;
  primaryCta: string;
  secondaryCta: string;
  githubCta: string;
  anchorTitle: string;
  sections: {
    id: string;
    title: string;
    intro?: string;
    paragraphs?: string[];
    bullets?: string[];
    cards?: Array<{ title: string; desc: string; tag?: string }>;
    steps?: Array<{ title: string; desc: string }>;
    links?: Array<{ label: string; url: string; description?: string }>;
    code?: string;
  }[];
  repoTitle: string;
  repoDescription: string;
  repoPoints: string[];
};

export type DocsContentBundle = {
  requestedLanguage: string;
  contentLanguage: string;
  isFallbackContent: boolean;
  locale: DocsLocale;
};

const CASDOOR_SIGNUP_ITEMS_URL = 'https://casdoor.ai/docs/application/signup-items-table';
const CASDOOR_SIGNIN_METHODS_URL = 'https://casdoor.ai/docs/application/signin-methods';
const CASDOOR_APP_CONFIG_URL = 'https://casdoor.ai/docs/application/config';
const CASDOOR_EMAIL_PROVIDER_URL = 'https://casdoor.ai/docs/provider/email/overview';

const QUICK_START_CODE = `cd deploy/compose
cp .env.example .env
./prepare-config.sh
docker compose up -d --build`;

const ENV_EXAMPLE_CODE = `CASDOOR_PUBLIC_URL=https://auth.example.com
DOTBLUE_PUBLIC_URL=https://app.example.com
DOTBLUE_BACKEND_PUBLIC_URL=https://api.example.com

DOTBLUE_ADMIN_USERNAME=admin
DOTBLUE_ADMIN_EMAIL=admin@example.com
DOTBLUE_ADMIN_PASSWORD=replace-with-a-strong-password

DOTBLUE_LLM_PROVIDER_TYPE=openai
DOTBLUE_LLM_API_BASE=https://ark.cn-beijing.volces.com/api/v3
DOTBLUE_LLM_API_KEY=replace-with-provider-key
DOTBLUE_LLM_MODEL=doubao-seed-2-0-mini-260428`;

const MINIMAL_DEPLOY_CODE = `Services:
- postgres
- redis
- casdoor
- dotblue
- web

Default ports:
- Web: 19000
- Backend: 18080
- Casdoor: 18000`;

const locales: Record<string, DocsLocale> = {
  en: {
    seoTitle: 'dotblue Product Docs | Quick Start, Deployment, and Operations',
    seoDescription:
      'Learn how to use dotblue with a real quick start, minimal deployment checklist, production deployment guidance, and troubleshooting for enterprise AI assistants.',
    seoKeywords:
      'dotblue product docs, quick start, enterprise AI assistants, minimal deployment, production deployment, Casdoor, troubleshooting, self-hosted AI platform',
    eyebrow: 'PRODUCT DOCUMENTATION',
    title: 'dotblue Product Docs',
    subtitle:
      'This guide is written for product, implementation, platform, and operations teams that need a practical path from first install to production rollout.',
    primaryCta: 'Open Sign-in',
    secondaryCta: 'Go to Dashboard',
    githubCta: 'View on GitHub',
    anchorTitle: 'On this page',
    sections: [
      {
        id: 'overview',
        title: 'Detailed Product Overview',
        intro: 'dotblue is an enterprise AI assistant delivery layer, not just a chat frontend.',
        cards: [
          { title: 'Assistant Management', desc: 'Create and manage assistants with independent prompts, models, and runtime settings.', tag: 'Agents' },
          { title: 'Real-time Chat', desc: 'Support streaming responses, conversation history, and visible execution states.', tag: 'Chat' },
          { title: 'Enterprise Structure', desc: 'Use organizations, members, invitations, and roles for real multi-user rollout.', tag: 'Enterprise' },
          { title: 'Platform LLM Control', desc: 'Centralize LLM provider, model, and API credentials at the platform layer.', tag: 'LLM Ops' },
          { title: 'Runtime Isolation', desc: 'Run assistants in isolated container runtime environments for stronger operational boundaries.', tag: 'Runtime' },
          { title: 'Casdoor Authentication', desc: 'Use branded sign-in, callback, and logout flows through Casdoor.', tag: 'Auth' },
        ],
      },
      {
        id: 'prerequisites',
        title: 'Before You Start',
        intro: 'Teams move faster when they prepare the environment and account information before the first boot.',
        bullets: [
          'Prepare Docker and Docker Compose in a working environment, or use WSL if your local workflow depends on it.',
          'Decide the public access addresses for Web, Backend, and Casdoor before generating config files.',
          'Prepare one valid LLM API key so the first assistant can actually answer after login.',
          'Prepare a strong admin username, email, and password for the first platform owner.',
          'Keep in mind that public URLs must use the real reachable host or domain instead of `localhost` when accessed from another machine.',
        ],
        code: ENV_EXAMPLE_CODE,
      },
      {
        id: 'quick-start',
        title: '5-Minute Quick Start',
        intro: 'Use the Compose stack first. It is the fastest path to a full end-to-end validation.',
        steps: [
          { title: 'Copy `.env.example` to `.env`', desc: 'Fill admin credentials, LLM provider values, and browser-facing public URLs.' },
          { title: 'Generate local runtime files', desc: 'Run `prepare-config.sh` or `prepare-config.ps1` so Casdoor and backend use aligned config.' },
          { title: 'Start the stack', desc: 'Run Compose to start postgres, redis, casdoor, dotblue, and web together.' },
          { title: 'Open the product home page', desc: 'Confirm the product site loads before continuing to sign-in.' },
          { title: 'Sign in with the admin account', desc: 'Complete the Casdoor login flow and verify callback success into the dashboard.' },
          { title: 'Create the first assistant', desc: 'Set the platform model first if needed, then create one assistant and open Chat.' },
        ],
        code: QUICK_START_CODE,
      },
      {
        id: 'advanced-auth',
        title: 'Advanced Login and Sign-up Options',
        intro: 'The local quick-start intentionally uses the simplest registration path so teams can boot the stack without email or SMS dependencies.',
        bullets: [
          'The default local initialization keeps sign-in on Password and keeps sign-up focused on Username, Display name, Password, and Confirm password.',
          'Email and phone verification are not enabled by default for quick local bring-up, because they require extra provider setup, templates, and delivery testing.',
          'When you need email verification, SMS verification, social login, WebAuthn, or more advanced sign-in policies, configure them in Casdoor itself and treat that as a production-grade auth rollout task.',
        ],
        links: [
          { label: 'Casdoor Sign-up Items', url: CASDOOR_SIGNUP_ITEMS_URL, description: 'Configure registration fields, optional email verification, and signup item rules.' },
          { label: 'Casdoor Sign-in Methods', url: CASDOOR_SIGNIN_METHODS_URL, description: 'Enable password, verification code, WebAuthn, LDAP, and other sign-in options.' },
          { label: 'Casdoor Application Config', url: CASDOOR_APP_CONFIG_URL, description: 'Review redirect URLs, resend timeout, and application-level auth settings.' },
          { label: 'Casdoor Email Provider', url: CASDOOR_EMAIL_PROVIDER_URL, description: 'Set up SMTP so email verification and password reset flows can actually send messages.' },
        ],
      },
      {
        id: 'first-run',
        title: 'What To Do Right After Login',
        intro: 'This is the real first-use sequence for a new team.',
        bullets: [
          'Open platform settings and confirm the LLM provider and default model are saved successfully.',
          'Open the assistant list and create one assistant with a clear system prompt.',
          'Open Chat, send a first message, and verify that the response path works end to end.',
          'If your rollout is team-based, create or switch enterprise context before inviting other members.',
          'Check that logout and sign-in both work cleanly before sharing the environment with other users.',
        ],
      },
      {
        id: 'minimal-deploy',
        title: 'Minimal Deployment',
        intro: 'Minimal deployment is enough for demos, POCs, QA, and first integration checks.',
        bullets: [
          'The baseline stack includes `postgres`, `redis`, `casdoor`, `dotblue`, and `web`.',
          'The frontend is built as a static web image and the backend handles APIs plus the embedded worker loop.',
          'Casdoor owns authentication and branded login pages, so its public URL must be browser-reachable.',
          'Generated runtime files should come from the prepare scripts rather than manual edits.',
          'When the stack is accessed through WSL or host IP, all public URLs must point to that reachable address.',
        ],
        code: MINIMAL_DEPLOY_CODE,
      },
      {
        id: 'production',
        title: 'Production Deployment',
        intro: 'Production readiness is mostly about stable URLs, durable dependencies, secure secrets, and operational visibility.',
        paragraphs: [
          'Use formal domains for web, backend, and auth instead of development hostnames. Keep browser-facing URLs separate from internal service addresses.',
          'Move PostgreSQL and Redis to durable environments with backup, restore, and monitoring. The local quick-start defaults are not long-term production architecture.',
          'Treat Casdoor branding, callback URLs, generated config files, and release sequencing as part of your deployment process.',
        ],
        bullets: [
          'Use HTTPS for `CASDOOR_PUBLIC_URL`, `DOTBLUE_PUBLIC_URL`, and `DOTBLUE_BACKEND_PUBLIC_URL`.',
          'Add reverse proxy, TLS, and baseline security headers.',
          'Inject API keys, admin passwords, and client secrets through secret management.',
          'Back up data directories, databases, and any business-critical storage paths.',
          'Monitor login success, callback failure, API error rates, and runtime container health.',
        ],
      },
      {
        id: 'troubleshooting',
        title: 'Common Troubleshooting',
        intro: 'These are the issues most teams hit during the first setup cycle.',
        bullets: [
          'Login jumps to the wrong host: check `CASDOOR_PUBLIC_URL`, `DOTBLUE_PUBLIC_URL`, and callback-related generated files.',
          'Login succeeds but dashboard is empty: confirm the backend can reach the database and the platform setup is complete.',
          'No model options when creating an assistant: configure the platform-level or enterprise-level model first.',
          'Chat does not answer after provider changes: recycle existing runtime containers so they pick up the latest model config.',
          'Branding looks inconsistent: regenerate config and verify Casdoor branding assets use the same public domain as the product site.',
        ],
      },
      {
        id: 'operations',
        title: 'Pre-launch Checklist',
        intro: 'Use this list before handing the system to real users.',
        bullets: [
          'Home page, docs page, login page, and dashboard all show aligned brand visuals and metadata.',
          'Casdoor sign-in, callback, and logout all work from the same public domain strategy.',
          'Platform model configuration saves successfully and assistants can choose the expected model.',
          'Create assistant, open chat, send message, and inspect runtime steps in one continuous flow.',
          'Permission boundaries for admins and non-admins are clear and organization data loads correctly.',
        ],
      },
    ],
    repoTitle: 'Open Source and GitHub',
    repoDescription: 'Use the official repository when you want to customize the product, inspect deployment details, or contribute fixes.',
    repoPoints: [
      'The repository contains frontend, backend, and deployment assets together.',
      'It is the right place to inspect real config templates and Compose examples.',
      'Issue reporting and customization should start from the official GitHub project.',
    ],
  },
  'zh-CN': {
    seoTitle: 'dotblue 产品文档 | 快速上手、部署与运维指南',
    seoDescription:
      '查看 dotblue 产品文档，了解真实可执行的快速上手、最小部署、生产环境部署与常见问题排查，帮助企业级 AI 助手平台顺利落地。',
    seoKeywords:
      'dotblue 产品文档,快速上手,企业级 AI 助手,最小部署,生产环境部署,Casdoor,排障,私有化部署',
    eyebrow: 'PRODUCT DOCUMENTATION',
    title: 'dotblue 产品使用文档',
    subtitle: '这份文档面向产品、实施、平台和运维团队，帮助你从第一次安装快速走到可交付的正式环境。',
    primaryCta: '前往登录',
    secondaryCta: '进入控制台',
    githubCta: '访问 GitHub',
    anchorTitle: '文档目录',
    sections: [
      {
        id: 'overview',
        title: '产品功能详细介绍',
        intro: 'dotblue 不是单一聊天页面，而是一套面向企业交付的 AI 助手能力层。',
        cards: [
          { title: '助手管理', desc: '为不同业务角色创建和管理独立助手，分别配置提示词、模型和运行方式。', tag: 'Agents' },
          { title: '实时对话', desc: '支持流式回复、会话历史和执行状态展示，便于业务与技术团队一起排查。', tag: 'Chat' },
          { title: '企业组织能力', desc: '支持企业、成员、邀请和角色边界，适合真实多人场景落地。', tag: 'Enterprise' },
          { title: '平台级模型管理', desc: '把模型供应商、模型名和凭据统一收敛到平台层管理。', tag: 'LLM Ops' },
          { title: '运行时隔离', desc: '助手运行在隔离容器环境中，更适合治理、扩展与稳定交付。', tag: 'Runtime' },
          { title: 'Casdoor 认证', desc: '通过 Casdoor 统一处理登录、回调、登出和品牌化登录页。', tag: 'Auth' },
        ],
      },
      {
        id: 'prerequisites',
        title: '开始前准备',
        intro: '第一次部署是否顺利，通常取决于环境和账号信息是否提前准备完整。',
        bullets: [
          '准备可用的 Docker 与 Docker Compose 环境，如你的本地开发依赖 WSL，则优先在 WSL 中执行。',
          '提前确认 Web、Backend、Casdoor 的公开访问地址，避免生成配置后再整体返工。',
          '准备至少一组可用的 LLM API Key，确保登录后可以真的创建并验证第一个助手。',
          '准备管理员账号、邮箱和强密码，作为系统首次拥有者。',
          '如果环境需要跨机器或通过宿主机 IP 访问，公开地址必须使用真实可访问的域名或 IP，而不是 `localhost`。',
        ],
        code: ENV_EXAMPLE_CODE,
      },
      {
        id: 'quick-start',
        title: '5 分钟快速上手',
        intro: '推荐先用 Compose 跑通全链路，这是最省心的方式。',
        steps: [
          { title: '复制环境变量模板', desc: '把 `.env.example` 复制成 `.env`，填写管理员信息、LLM 配置和公开访问地址。' },
          { title: '生成运行时配置', desc: '执行 `prepare-config.sh` 或 `prepare-config.ps1`，让 Casdoor 和后端配置保持一致。' },
          { title: '启动整套服务', desc: '通过 Compose 一次启动 postgres、redis、casdoor、dotblue 和 web。' },
          { title: '打开产品首页', desc: '先确认首页可访问，再继续进入登录流程。' },
          { title: '使用管理员账号登录', desc: '完成 Casdoor 登录并确认能正确回调到 Dashboard。' },
          { title: '创建第一个助手', desc: '必要时先配置平台模型，然后创建助手并进入 Chat 验证效果。' },
        ],
        code: QUICK_START_CODE,
      },
      {
        id: 'advanced-auth',
        title: '高阶登录注册配置',
        intro: '本地快速拉起默认使用最简注册路径，优先保证服务能直接跑起来，而不是一开始就依赖邮箱或短信能力。',
        bullets: [
          '当前默认初始化会把注册页收敛为 Username、Display name、Password、Confirm password，不默认启用邮箱或手机验证注册。',
          '这样做是为了降低本地体验门槛，避免首次启动时还要额外配置 SMTP、短信服务商、验证码模板和投递连通性。',
          '如果你需要邮箱验证码注册、手机验证码注册、第三方社交登录、WebAuthn 或更复杂的登录策略，建议直接参考 Casdoor 官方文档完成认证体系配置。',
        ],
        links: [
          { label: 'Casdoor 注册项配置', url: CASDOOR_SIGNUP_ITEMS_URL, description: '配置注册字段、邮箱验证规则以及注册表单项。' },
          { label: 'Casdoor 登录方式配置', url: CASDOOR_SIGNIN_METHODS_URL, description: '配置 Password、验证码、WebAuthn、LDAP 等登录方式。' },
          { label: 'Casdoor 应用认证配置', url: CASDOOR_APP_CONFIG_URL, description: '查看回调地址、验证码重发超时和应用级认证设置。' },
          { label: 'Casdoor 邮件服务配置', url: CASDOOR_EMAIL_PROVIDER_URL, description: '配置 SMTP 邮件服务，支撑邮箱验证码与找回密码能力。' },
        ],
      },
      {
        id: 'first-run',
        title: '首次登录后该做什么',
        intro: '下面这条路径才是一个新团队真正的首次使用流程。',
        bullets: [
          '先进入平台设置，确认 LLM 供应商和默认模型已经保存成功。',
          '进入助手列表，创建一个带清晰系统提示词的助手。',
          '打开 Chat 页面，发送第一条消息，确认从前端到模型的完整路径可用。',
          '如果要按团队落地，先创建或切换企业，再邀请其他成员。',
          '在共享给其他用户前，先验证注销和重新登录都能正常工作。',
        ],
      },
      {
        id: 'minimal-deploy',
        title: '最小部署',
        intro: '最小部署适合演示、POC、测试联调和第一次评估。',
        bullets: [
          '基础服务包括 `postgres`、`redis`、`casdoor`、`dotblue`、`web`。',
          '前端通过静态 Web 镜像提供页面，后端负责 API 和嵌入式 worker。',
          'Casdoor 负责认证与品牌化登录，所以其公开地址必须可被浏览器访问。',
          '运行中的配置应来自 prepare 脚本生成的结果，而不是手工拼凑。',
          '如果通过 WSL、宿主机 IP 或其他机器访问，公开地址必须统一为真实可访问地址。',
        ],
        code: MINIMAL_DEPLOY_CODE,
      },
      {
        id: 'production',
        title: '生产环境部署',
        intro: '生产部署的重点是稳定地址、持久化依赖、密钥管理和可运维性，而不是仅仅把服务拉起来。',
        paragraphs: [
          'Web、Backend 和 Casdoor 应该使用正式域名，浏览器访问地址和容器内部访问地址要明确分离。',
          'PostgreSQL 和 Redis 应切换到可备份、可恢复、可监控的长期运行环境，本地默认卷配置不适合长期生产。',
          'Casdoor 的品牌配置、回调地址、生成配置和发布顺序都应该进入正式发布流程，否则最容易出现登录和品牌不一致问题。',
        ],
        bullets: [
          '为 `CASDOOR_PUBLIC_URL`、`DOTBLUE_PUBLIC_URL`、`DOTBLUE_BACKEND_PUBLIC_URL` 统一使用 HTTPS。',
          '接入反向代理、TLS 和基础安全头。',
          '通过 Secret 管理注入 API Key、管理员密码和客户端密钥。',
          '为数据库、运行时数据目录和核心存储建立备份与恢复策略。',
          '监控登录成功率、回调失败率、API 错误率和运行时容器状态。',
        ],
      },
      {
        id: 'troubleshooting',
        title: '常见问题排查',
        intro: '第一次部署时最容易卡在下面这些问题上。',
        bullets: [
          '登录跳到了错误地址：优先检查 `CASDOOR_PUBLIC_URL`、`DOTBLUE_PUBLIC_URL` 和生成后的回调配置。',
          '登录成功但控制台为空：检查后端数据库连接和平台初始化状态。',
          '创建助手时没有模型可选：先配置平台级或企业级模型。',
          '修改模型后聊天仍不生效：重建或回收已有运行时容器，让它重新加载最新配置。',
          '品牌样式不一致：重新生成配置，并确认 Casdoor 静态资源与产品页使用同一公开域名。',
        ],
      },
      {
        id: 'operations',
        title: '上线前检查项',
        intro: '在正式交付给业务用户前，建议至少完成这些检查。',
        bullets: [
          '首页、文档页、登录页和 Dashboard 的品牌视觉与标题描述保持一致。',
          'Casdoor 登录、回调、注销都能在同一套公开地址策略下正常工作。',
          '平台模型配置保存成功，创建助手时能看到正确模型选项。',
          '创建助手、进入 Chat、发送消息、查看运行步骤形成闭环。',
          '管理员与普通用户权限边界清晰，组织相关数据可正常读取。',
        ],
      },
    ],
    repoTitle: '开源与 GitHub',
    repoDescription: '如果你需要二次开发、排查部署细节或提交修复，建议从官方仓库开始。',
    repoPoints: [
      '仓库同时包含前端、后端和部署资产。',
      '实际配置模板和 Compose 示例都以仓库内容为准。',
      'Issue 提交和定制开发都建议基于官方 GitHub 项目进行。',
    ],
  },
  ja: {
    seoTitle: 'dotblue ドキュメント | クイックスタート、デプロイ、運用',
    seoDescription:
      'dotblue のクイックスタート、最小デプロイ、本番デプロイ、Casdoor 認証、初期トラブル対応をまとめた製品ドキュメントです。',
    seoKeywords: 'dotblue, ドキュメント, クイックスタート, AI アシスタント, デプロイ, Casdoor, 自社運用',
    eyebrow: 'PRODUCT DOCUMENTATION',
    title: 'dotblue 製品ドキュメント',
    subtitle: '初回セットアップから本番展開までを短い手順でたどれるように整理したガイドです。',
    primaryCta: 'ログインへ',
    secondaryCta: 'ダッシュボードへ',
    githubCta: 'GitHub を開く',
    anchorTitle: '目次',
    sections: [
      {
        id: 'overview',
        title: '機能概要',
        intro: 'dotblue は企業向け AI アシスタントを運用するための製品レイヤーです。',
        cards: [
          { title: 'アシスタント管理', desc: '役割ごとに独立したアシスタント設定を管理できます。', tag: 'Agents' },
          { title: 'リアルタイム対話', desc: 'ストリーミング応答、会話履歴、実行状態を確認できます。', tag: 'Chat' },
          { title: '組織管理', desc: '企業、メンバー、招待、権限を扱えます。', tag: 'Enterprise' },
          { title: 'LLM 設定', desc: 'モデル接続と資格情報をプラットフォームで集中管理します。', tag: 'LLM Ops' },
          { title: 'ランタイム分離', desc: 'コンテナ実行環境でアシスタントを分離します。', tag: 'Runtime' },
          { title: 'Casdoor 認証', desc: 'ログイン、コールバック、ブランド UI を一元化します。', tag: 'Auth' },
        ],
      },
      {
        id: 'prerequisites',
        title: '開始前の準備',
        bullets: [
          'Docker / Docker Compose が利用できることを確認します。',
          'Web、Backend、Casdoor の公開 URL を先に決めます。',
          '最初の検証に使う LLM API Key を準備します。',
          '管理者アカウント用のメールと強いパスワードを用意します。',
          '別マシンからアクセスする場合は `localhost` ではなく到達可能な IP またはドメインを使います。',
        ],
        code: ENV_EXAMPLE_CODE,
      },
      {
        id: 'quick-start',
        title: '5 分クイックスタート',
        steps: [
          { title: '.env を作成', desc: '管理者情報、LLM 設定、公開 URL を入力します。' },
          { title: '設定を生成', desc: '`prepare-config` を実行し Casdoor と backend を揃えます。' },
          { title: 'スタックを起動', desc: 'Compose で全サービスを一括起動します。' },
          { title: 'ホームを確認', desc: '製品ホームが見えることを先に確認します。' },
          { title: '管理者でサインイン', desc: 'Casdoor から Dashboard へ戻れることを確認します。' },
          { title: '最初のアシスタントを作成', desc: 'モデルを選び Chat で応答を確認します。' },
        ],
        code: QUICK_START_CODE,
      },
      {
        id: 'first-run',
        title: '初回ログイン後の流れ',
        bullets: [
          'プラットフォーム設定で LLM 接続を保存します。',
          'アシスタントを 1 つ作成します。',
          'Chat で最初のメッセージを送ります。',
          '必要なら企業コンテキストを作成または切り替えます。',
          '共有前にログアウトと再ログインを確認します。',
        ],
      },
      {
        id: 'minimal-deploy',
        title: '最小構成デプロイ',
        bullets: [
          '基本サービスは postgres、redis、casdoor、dotblue、web です。',
          'Casdoor の公開 URL はブラウザから到達可能である必要があります。',
          '生成設定は手動編集ではなく prepare スクリプトから作るのが前提です。',
          'WSL やホスト IP で使う場合は公開 URL をその到達先へ合わせます。',
        ],
        code: MINIMAL_DEPLOY_CODE,
      },
      {
        id: 'production',
        title: '本番デプロイ',
        bullets: [
          '公開 URL は HTTPS の正式ドメインを利用します。',
          'DB、Redis、データパスにはバックアップ戦略を用意します。',
          'API Key と秘密情報は Secret 管理で注入します。',
          'ログイン、API エラー、ランタイム状態を監視します。',
        ],
      },
      {
        id: 'troubleshooting',
        title: 'よくある問題',
        bullets: [
          '誤ったホストへ遷移する場合は public URL 設定を確認します。',
          'モデル一覧が空ならプラットフォームモデル設定を確認します。',
          'モデル変更後に反映されない場合は既存ランタイムを再作成します。',
          'ブランド表示がずれる場合は Casdoor と製品側の公開ドメインを揃えます。',
        ],
      },
      {
        id: 'operations',
        title: '公開前チェック',
        bullets: [
          'ホーム、ドキュメント、ログイン画面のブランドが一致している。',
          'サインイン、コールバック、ログアウトが正常。',
          'アシスタント作成から Chat まで一連の流れが通る。',
          '権限境界と組織データ表示が正しい。',
        ],
      },
    ],
    repoTitle: 'GitHub と OSS',
    repoDescription: 'カスタマイズや調査は公式リポジトリから始めるのが最短です。',
    repoPoints: ['フロントエンド、バックエンド、デプロイ資産を含みます。', 'Compose と設定テンプレートの参照元です。', 'Issue や Fork も公式リポジトリを起点にします。'],
  },
  ko: {
    seoTitle: 'dotblue 문서 | 빠른 시작, 배포, 운영',
    seoDescription:
      'dotblue 의 빠른 시작, 최소 배포, 운영 배포, Casdoor 로그인, 초기 장애 대응을 정리한 제품 문서입니다.',
    seoKeywords: 'dotblue, 문서, 빠른 시작, AI 어시스턴트, 배포, Casdoor, 셀프호스팅',
    eyebrow: 'PRODUCT DOCUMENTATION',
    title: 'dotblue 제품 문서',
    subtitle: '첫 설치부터 운영 환경 전환까지 필요한 핵심 절차를 짧고 실제적으로 정리했습니다.',
    primaryCta: '로그인으로 이동',
    secondaryCta: '대시보드 열기',
    githubCta: 'GitHub 보기',
    anchorTitle: '문서 목차',
    sections: [
      {
        id: 'overview',
        title: '기능 소개',
        intro: 'dotblue 는 기업용 AI 어시스턴트 운영을 위한 제품 계층입니다.',
        cards: [
          { title: '어시스턴트 관리', desc: '역할별로 독립된 어시스턴트 구성을 관리합니다.', tag: 'Agents' },
          { title: '실시간 채팅', desc: '스트리밍 응답, 대화 기록, 실행 상태를 확인합니다.', tag: 'Chat' },
          { title: '조직 관리', desc: '기업, 멤버, 초대, 권한을 관리합니다.', tag: 'Enterprise' },
          { title: 'LLM 설정', desc: '모델 연결과 자격 증명을 플랫폼에서 통합 관리합니다.', tag: 'LLM Ops' },
          { title: '런타임 격리', desc: '컨테이너 실행 환경으로 어시스턴트를 분리합니다.', tag: 'Runtime' },
          { title: 'Casdoor 인증', desc: '로그인과 콜백, 브랜딩된 인증 화면을 통합합니다.', tag: 'Auth' },
        ],
      },
      {
        id: 'prerequisites',
        title: '시작 전 준비',
        bullets: [
          'Docker 와 Docker Compose 가 동작하는지 확인합니다.',
          'Web, Backend, Casdoor 의 공개 주소를 먼저 정합니다.',
          '첫 검증에 사용할 LLM API Key 를 준비합니다.',
          '관리자 계정용 이메일과 강한 비밀번호를 준비합니다.',
          '다른 장치에서 접근한다면 `localhost` 대신 실제 접근 가능한 IP 또는 도메인을 사용합니다.',
        ],
        code: ENV_EXAMPLE_CODE,
      },
      {
        id: 'quick-start',
        title: '5분 빠른 시작',
        steps: [
          { title: '.env 생성', desc: '관리자 정보, LLM 설정, 공개 URL 을 입력합니다.' },
          { title: '설정 생성', desc: '`prepare-config` 로 Casdoor 와 backend 설정을 맞춥니다.' },
          { title: '스택 기동', desc: 'Compose 로 전체 서비스를 함께 올립니다.' },
          { title: '홈페이지 확인', desc: '제품 홈이 먼저 열리는지 확인합니다.' },
          { title: '관리자로 로그인', desc: 'Casdoor 에서 대시보드로 콜백되는지 확인합니다.' },
          { title: '첫 어시스턴트 생성', desc: '모델을 선택하고 Chat 에서 응답을 검증합니다.' },
        ],
        code: QUICK_START_CODE,
      },
      {
        id: 'first-run',
        title: '첫 로그인 후 해야 할 일',
        bullets: [
          '플랫폼 설정에서 LLM 연결을 저장합니다.',
          '어시스턴트를 하나 생성합니다.',
          'Chat 에서 첫 메시지를 보냅니다.',
          '필요하면 엔터프라이즈 컨텍스트를 생성하거나 전환합니다.',
          '공유 전에 로그아웃/재로그인을 확인합니다.',
        ],
      },
      {
        id: 'minimal-deploy',
        title: '최소 배포',
        bullets: [
          '기본 서비스는 postgres, redis, casdoor, dotblue, web 입니다.',
          'Casdoor 공개 URL 은 브라우저에서 접근 가능해야 합니다.',
          '실행 설정은 prepare 스크립트 결과를 기준으로 사용합니다.',
          'WSL 또는 호스트 IP 로 접근 시 공개 URL 도 그 주소로 맞춰야 합니다.',
        ],
        code: MINIMAL_DEPLOY_CODE,
      },
      {
        id: 'production',
        title: '운영 배포',
        bullets: [
          '공개 URL 은 HTTPS 기반의 실제 도메인을 사용합니다.',
          'DB, Redis, 데이터 경로에 대한 백업 전략을 준비합니다.',
          'API Key 와 비밀값은 Secret 관리로 주입합니다.',
          '로그인, API 오류, 런타임 상태를 모니터링합니다.',
        ],
      },
      {
        id: 'troubleshooting',
        title: '자주 발생하는 문제',
        bullets: [
          '잘못된 호스트로 이동하면 public URL 설정을 확인합니다.',
          '모델 목록이 비어 있으면 플랫폼 모델 설정을 확인합니다.',
          '모델 변경 후 반영되지 않으면 런타임 컨테이너를 재생성합니다.',
          '브랜딩이 다르게 보이면 Casdoor 와 제품 사이트의 공개 도메인을 맞춥니다.',
        ],
      },
      {
        id: 'operations',
        title: '오픈 전 체크리스트',
        bullets: [
          '홈, 문서, 로그인 화면의 브랜딩이 일치한다.',
          '로그인, 콜백, 로그아웃이 정상 동작한다.',
          '어시스턴트 생성부터 Chat 까지 전체 흐름이 통과한다.',
          '권한 경계와 조직 데이터 노출이 올바르다.',
        ],
      },
    ],
    repoTitle: 'GitHub 및 오픈소스',
    repoDescription: '커스터마이징과 조사 작업은 공식 저장소에서 시작하는 것이 가장 빠릅니다.',
    repoPoints: ['프론트엔드, 백엔드, 배포 자산이 함께 포함됩니다.', 'Compose 와 설정 템플릿의 기준 저장소입니다.', '이슈 보고와 포크도 공식 저장소를 기준으로 진행합니다.'],
  },
  fr: {
    seoTitle: 'Documentation dotblue | Demarrage rapide, deploiement et exploitation',
    seoDescription:
      'Consultez la documentation produit dotblue pour le demarrage rapide, le deploiement minimal, le deploiement de production et les verifications de connexion Casdoor.',
    seoKeywords: 'dotblue, documentation, demarrage rapide, assistants IA, deploiement, Casdoor, auto-heberge',
    eyebrow: 'PRODUCT DOCUMENTATION',
    title: 'Documentation produit dotblue',
    subtitle: 'Un guide concis pour passer de la premiere installation a une mise en production exploitable.',
    primaryCta: 'Aller a la connexion',
    secondaryCta: 'Ouvrir le tableau de bord',
    githubCta: 'Voir sur GitHub',
    anchorTitle: 'Sommaire',
    sections: [
      {
        id: 'overview',
        title: 'Vue detaillee du produit',
        intro: 'dotblue est une couche produit pour assistants IA d entreprise, pas seulement une interface de chat.',
        cards: [
          { title: 'Gestion des assistants', desc: 'Creer et gerer des assistants avec des configurations separees.', tag: 'Agents' },
          { title: 'Chat temps reel', desc: 'Reponses en streaming, historique et etats d execution.', tag: 'Chat' },
          { title: 'Structure entreprise', desc: 'Organisations, membres, invitations et roles.', tag: 'Enterprise' },
          { title: 'Controle LLM', desc: 'Centraliser fournisseurs, modeles et credentials.', tag: 'LLM Ops' },
          { title: 'Isolation runtime', desc: 'Executer les assistants dans des environnements conteneurises isoles.', tag: 'Runtime' },
          { title: 'Authentification Casdoor', desc: 'Connexion, callback et branding relies a Casdoor.', tag: 'Auth' },
        ],
      },
      {
        id: 'prerequisites',
        title: 'Avant de commencer',
        bullets: [
          'Verifier Docker et Docker Compose.',
          'Choisir les URLs publiques pour Web, Backend et Casdoor.',
          'Preparer une cle API LLM valide.',
          'Preparer un compte administrateur avec un mot de passe fort.',
          'Si l acces se fait depuis une autre machine, utiliser une IP ou un domaine atteignable au lieu de `localhost`.',
        ],
        code: ENV_EXAMPLE_CODE,
      },
      {
        id: 'quick-start',
        title: 'Demarrage rapide en 5 minutes',
        steps: [
          { title: 'Creer `.env`', desc: 'Renseigner admin, LLM et URLs publiques.' },
          { title: 'Generer la configuration', desc: 'Executer `prepare-config` pour aligner Casdoor et backend.' },
          { title: 'Demarrer la stack', desc: 'Lancer tous les services avec Compose.' },
          { title: 'Verifier la page d accueil', desc: 'Confirmer que le site est accessible.' },
          { title: 'Se connecter comme admin', desc: 'Verifier le retour vers le dashboard apres Casdoor.' },
          { title: 'Creer le premier assistant', desc: 'Choisir un modele puis tester le chat.' },
        ],
        code: QUICK_START_CODE,
      },
      {
        id: 'first-run',
        title: 'Que faire apres la premiere connexion',
        bullets: [
          'Sauvegarder la configuration LLM de la plateforme.',
          'Creer un assistant.',
          'Envoyer un premier message dans Chat.',
          'Creer ou changer de contexte entreprise si necessaire.',
          'Verifier logout et reconnexion avant de partager l environnement.',
        ],
      },
      {
        id: 'minimal-deploy',
        title: 'Deploiement minimal',
        bullets: [
          'La stack minimale comprend postgres, redis, casdoor, dotblue et web.',
          'L URL publique de Casdoor doit etre atteignable depuis le navigateur.',
          'Les fichiers runtime doivent venir des scripts prepare.',
          'En acces WSL ou IP hote, les URLs publiques doivent pointer vers cette adresse.',
        ],
        code: MINIMAL_DEPLOY_CODE,
      },
      {
        id: 'production',
        title: 'Deploiement de production',
        bullets: [
          'Utiliser des domaines HTTPS reellement exposes.',
          'Mettre en place sauvegardes pour DB, Redis et donnees runtime.',
          'Injecter les secrets via un gestionnaire de secrets.',
          'Superviser la connexion, les callbacks, les API et le runtime.',
        ],
      },
      {
        id: 'troubleshooting',
        title: 'Problemes frequents',
        bullets: [
          'Mauvais host au login : verifier les public URLs.',
          'Aucun modele disponible : verifier la configuration de modeles.',
          'Changement de modele non applique : recreer les runtimes existants.',
          'Branding incoherent : aligner le domaine public de Casdoor et du site produit.',
        ],
      },
      {
        id: 'operations',
        title: 'Checklist avant mise en ligne',
        bullets: [
          'Branding coherent entre home, docs et login.',
          'Connexion, callback et logout valides.',
          'Flux complet assistant -> chat valide.',
          'Frontieres de permissions et donnees organisationnelles correctes.',
        ],
      },
    ],
    repoTitle: 'GitHub et open source',
    repoDescription: 'Le depot officiel reste le meilleur point de depart pour personnaliser et auditer le produit.',
    repoPoints: ['Frontend, backend et deploiement sont regroupes.', 'Les templates de configuration y sont references.', 'Issues et forks doivent partir du depot officiel.'],
  },
  es: {
    seoTitle: 'Documentacion dotblue | Inicio rapido, despliegue y operacion',
    seoDescription:
      'Consulta la documentacion de dotblue para inicio rapido, despliegue minimo, produccion, autenticacion con Casdoor y resolucion de problemas iniciales.',
    seoKeywords: 'dotblue, documentacion, inicio rapido, asistentes IA, despliegue, Casdoor, autoalojado',
    eyebrow: 'PRODUCT DOCUMENTATION',
    title: 'Documentacion de producto dotblue',
    subtitle: 'Una guia practica para pasar de la primera instalacion a un entorno listo para usuarios reales.',
    primaryCta: 'Ir al acceso',
    secondaryCta: 'Abrir panel',
    githubCta: 'Ver en GitHub',
    anchorTitle: 'Contenido',
    sections: [
      {
        id: 'overview',
        title: 'Resumen detallado del producto',
        intro: 'dotblue es una capa de producto para asistentes de IA empresariales, no solo una interfaz de chat.',
        cards: [
          { title: 'Gestion de asistentes', desc: 'Crear y gestionar asistentes con configuraciones separadas.', tag: 'Agents' },
          { title: 'Chat en tiempo real', desc: 'Respuestas en streaming, historial y estados de ejecucion.', tag: 'Chat' },
          { title: 'Estructura empresarial', desc: 'Organizaciones, miembros, invitaciones y roles.', tag: 'Enterprise' },
          { title: 'Control LLM', desc: 'Centralizar proveedores, modelos y credenciales.', tag: 'LLM Ops' },
          { title: 'Aislamiento runtime', desc: 'Ejecutar asistentes en entornos contenedorizados aislados.', tag: 'Runtime' },
          { title: 'Autenticacion Casdoor', desc: 'Unificar login, callback y branding con Casdoor.', tag: 'Auth' },
        ],
      },
      {
        id: 'prerequisites',
        title: 'Antes de empezar',
        bullets: [
          'Verifica Docker y Docker Compose.',
          'Define las URLs publicas de Web, Backend y Casdoor.',
          'Prepara una API Key valida para el LLM.',
          'Prepara una cuenta administradora con contrasena fuerte.',
          'Si accedes desde otra maquina, usa una IP o dominio accesible en lugar de `localhost`.',
        ],
        code: ENV_EXAMPLE_CODE,
      },
      {
        id: 'quick-start',
        title: 'Inicio rapido en 5 minutos',
        steps: [
          { title: 'Crear `.env`', desc: 'Completa admin, LLM y URLs publicas.' },
          { title: 'Generar configuracion', desc: 'Ejecuta `prepare-config` para alinear Casdoor y backend.' },
          { title: 'Levantar la stack', desc: 'Inicia todos los servicios con Compose.' },
          { title: 'Comprobar la home', desc: 'Verifica que el sitio de producto responde.' },
          { title: 'Entrar como admin', desc: 'Confirma el callback desde Casdoor al dashboard.' },
          { title: 'Crear el primer asistente', desc: 'Selecciona un modelo y prueba el chat.' },
        ],
        code: QUICK_START_CODE,
      },
      {
        id: 'first-run',
        title: 'Que hacer tras el primer login',
        bullets: [
          'Guardar la configuracion LLM de la plataforma.',
          'Crear un asistente.',
          'Enviar el primer mensaje en Chat.',
          'Crear o cambiar el contexto empresarial si hace falta.',
          'Comprobar logout y nuevo login antes de compartir el entorno.',
        ],
      },
      {
        id: 'minimal-deploy',
        title: 'Despliegue minimo',
        bullets: [
          'La base incluye postgres, redis, casdoor, dotblue y web.',
          'La URL publica de Casdoor debe ser accesible desde el navegador.',
          'Los archivos runtime deben venir de los scripts prepare.',
          'Si se usa WSL o IP del host, las URLs publicas deben apuntar a esa direccion.',
        ],
        code: MINIMAL_DEPLOY_CODE,
      },
      {
        id: 'production',
        title: 'Despliegue en produccion',
        bullets: [
          'Usa dominios HTTPS reales para todas las URLs publicas.',
          'Prepara copias de seguridad para DB, Redis y datos runtime.',
          'Inyecta secretos mediante gestion de secretos.',
          'Supervisa login, callbacks, APIs y estado del runtime.',
        ],
      },
      {
        id: 'troubleshooting',
        title: 'Problemas frecuentes',
        bullets: [
          'Salto a host incorrecto: revisa las public URLs.',
          'Sin modelos disponibles: revisa la configuracion de modelos.',
          'Cambio de modelo no aplicado: recrea los runtimes existentes.',
          'Branding inconsistente: alinea el dominio publico de Casdoor y del sitio.',
        ],
      },
      {
        id: 'operations',
        title: 'Checklist antes del lanzamiento',
        bullets: [
          'Branding consistente entre home, docs y login.',
          'Login, callback y logout validados.',
          'Flujo completo asistente -> chat verificado.',
          'Permisos y datos organizativos correctos.',
        ],
      },
    ],
    repoTitle: 'GitHub y codigo abierto',
    repoDescription: 'El repositorio oficial es el mejor punto de partida para personalizar e inspeccionar el producto.',
    repoPoints: ['Incluye frontend, backend y despliegue.', 'Es la referencia para plantillas de configuracion.', 'Issues y forks deben partir del proyecto oficial.'],
  },
};

export function getDocsContent(language: string) {
  const resolved = resolveSupportedLanguage(language);
  return locales[resolved] || locales.en;
}

export function getDocsContentBundle(language: string): DocsContentBundle {
  const resolved = resolveSupportedLanguage(language);
  const locale = locales[resolved];
  if (locale) {
    return {
      requestedLanguage: resolved,
      contentLanguage: resolved,
      isFallbackContent: false,
      locale,
    };
  }
  return {
    requestedLanguage: resolved,
    contentLanguage: 'en',
    isFallbackContent: true,
    locale: locales.en,
  };
}
