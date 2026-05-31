import { getLocalizedPath, resolveSupportedLanguage } from '../../i18n/config';

export const DOCS_REPO_URL = 'https://github.com/12345dotblue/dotblue';
export const CASDOOR_SIGNUP_ITEMS_URL = 'https://casdoor.ai/docs/application/signup-items-table';
export const CASDOOR_SIGNIN_METHODS_URL = 'https://casdoor.ai/docs/application/signin-methods';
export const CASDOOR_APP_CONFIG_URL = 'https://casdoor.ai/docs/application/config';
export const CASDOOR_EMAIL_PROVIDER_URL = 'https://casdoor.ai/docs/provider/email/overview';

export type DocsLink = {
  label: string;
  url: string;
  description?: string;
};

export type DocsStep = {
  title: string;
  desc: string;
};

export type DocsCodeBlock = {
  language: string;
  value: string;
};

export type DocsArticleSection = {
  id: string;
  title: string;
  intro?: string;
  paragraphs?: string[];
  bullets?: string[];
  steps?: DocsStep[];
  code?: DocsCodeBlock;
  links?: DocsLink[];
  note?: string;
};

export type DocsArticle = {
  sectionSlug: string;
  slug: string;
  title: string;
  summary: string;
  seoTitle: string;
  seoDescription: string;
  readingTime: string;
  sections: DocsArticleSection[];
};

export type DocsSection = {
  slug: string;
  title: string;
  description: string;
  articles: DocsArticle[];
};

export type DocsLibrary = {
  homeSeoTitle: string;
  homeSeoDescription: string;
  homeSeoKeywords: string;
  eyebrow: string;
  title: string;
  subtitle: string;
  categoriesLabel: string;
  popularLabel: string;
  sectionDescriptionLabel: string;
  quickLinksTitle: string;
  quickLinks: DocsLink[];
  repoTitle: string;
  repoDescription: string;
  sections: DocsSection[];
};

const QUICK_START_CODE = `cd deploy/compose
cp .env.example .env
./prepare-config.sh
docker compose up -d --build`;

const WINDOWS_QUICK_START_CODE = `cd deploy\\compose
copy .env.example .env
.\\prepare-config.ps1
docker compose up -d --build`;

const ENV_SAMPLE_CODE = `CASDOOR_PUBLIC_URL=https://auth.example.com
DOTBLUE_PUBLIC_URL=https://app.example.com
DOTBLUE_BACKEND_PUBLIC_URL=https://api.example.com

DOTBLUE_ADMIN_USERNAME=admin
DOTBLUE_ADMIN_EMAIL=admin@example.com
DOTBLUE_ADMIN_PASSWORD=replace-with-a-strong-password

DOTBLUE_LLM_PROVIDER_TYPE=openai
DOTBLUE_LLM_API_BASE=https://api.openai.com/v1
DOTBLUE_LLM_API_KEY=replace-with-provider-key
DOTBLUE_LLM_MODEL=gpt-4.1-mini`;

const DOCKER_SERVICES_CODE = `Services
- postgres
- redis
- casdoor
- dotblue
- web

Browser-facing ports
- Web: 19000
- Backend: 18080
- Casdoor: 18000`;

const REVERSE_PROXY_CHECKLIST = `Deployment checkpoints
1. Use formal domains for app, API, and auth
2. Terminate TLS at a trusted reverse proxy
3. Forward X-Forwarded-* headers correctly
4. Keep internal container addresses separate from public URLs
5. Rotate admin and provider secrets before launch`;

function buildQuickLinks(language: string): DocsLink[] {
  return [
    { label: language === 'zh-CN' ? '打开产品首页' : 'Open Product Home', url: getLocalizedPath('/', language) },
    { label: language === 'zh-CN' ? '进入登录页' : 'Open Sign-in', url: getLocalizedPath('/login', language) },
    { label: 'GitHub', url: DOCS_REPO_URL, description: language === 'zh-CN' ? '查看源码、部署资产与 issue。' : 'Inspect source code, deployment assets, and issues.' },
  ];
}

const englishLibrary: DocsLibrary = {
  homeSeoTitle: 'dotblue Docs | Getting Started, Product Guides, and Deployment',
  homeSeoDescription:
    'Explore dotblue product documentation with structured guides for getting started, workspace operations, authentication, deployment, and production rollout.',
  homeSeoKeywords:
    'dotblue docs, enterprise AI assistant docs, getting started, deployment guide, Casdoor login, self-hosted AI workspace, production rollout',
  eyebrow: 'PRODUCT DOCS',
  title: 'Documentation for enterprise AI assistant delivery',
  subtitle:
    'Read dotblue like a real product manual: start with the overview, move through login and workspace setup, then go deeper into deployment, operations, and production rollout.',
  categoriesLabel: 'Documentation categories',
  popularLabel: 'Popular pages',
  sectionDescriptionLabel: 'What you will learn',
  quickLinksTitle: 'Quick links',
  quickLinks: buildQuickLinks('en'),
  repoTitle: 'Open source and implementation reference',
  repoDescription:
    'The official repository is still the best place to inspect deployment templates, understand implementation details, and connect the product docs to the actual codebase.',
  sections: [
    {
      slug: 'getting-started',
      title: 'Getting started',
      description: 'Understand what dotblue is, how the first successful setup looks, and how to get from first boot to first working assistant quickly.',
      articles: [
        {
          sectionSlug: 'getting-started',
          slug: 'dotblue-overview',
          title: 'dotblue overview',
          summary: 'What dotblue is, who it is built for, and what a successful rollout usually looks like.',
          seoTitle: 'dotblue Overview | Product Documentation',
          seoDescription:
            'Learn what dotblue is, how enterprise teams use it, and how it connects productized assistants, authentication, runtime operations, and deployment.',
          readingTime: '6 min read',
          sections: [
            {
              id: 'what-it-is',
              title: 'What dotblue is',
              paragraphs: [
                'dotblue is an enterprise AI assistant delivery surface. It is not just a chat UI and not just a model wrapper. It brings together branded sign-in, platform-level model configuration, assistant management, team-oriented access control, and runtime operations in one product experience.',
                'The product is designed for teams that need to move from idea to deployable assistant experience quickly, while still keeping enough control to support real environments, multiple users, and organization boundaries.',
              ],
            },
            {
              id: 'core-capabilities',
              title: 'Core capabilities',
              bullets: [
                'Branded authentication through Casdoor, including callback and logout alignment.',
                'Assistant lifecycle management with prompt, model, and runtime settings.',
                'Platform and enterprise model configuration for shared LLM governance.',
                'Chat surfaces with execution visibility and conversation continuity.',
                'Deployment assets that support local bring-up, staging validation, and production rollout.',
              ],
            },
            {
              id: 'who-uses-it',
              title: 'Who this product is for',
              paragraphs: [
                'dotblue fits product teams launching internal assistants, implementation teams delivering customer environments, and platform teams standardizing AI assistant rollout across organizations.',
                'The best first use case is a focused assistant with clear business value, such as customer support, knowledge lookup, sales copilot, or internal operations assistance.',
              ],
            },
            {
              id: 'first-success',
              title: 'What first success looks like',
              steps: [
                { title: 'Open the product site', desc: 'Confirm the localized home page and documentation are available from the same public URL strategy.' },
                { title: 'Sign in through Casdoor', desc: 'Verify the branded login flow, callback, and session establishment into the dashboard.' },
                { title: 'Configure a model', desc: 'Save at least one platform or enterprise LLM model so assistants can respond.' },
                { title: 'Create an assistant', desc: 'Define one assistant with a narrow job, a clear system prompt, and a predictable output expectation.' },
                { title: 'Open Chat and send a message', desc: 'Confirm the user-facing conversation flow and runtime behavior end to end.' },
              ],
            },
          ],
        },
        {
          sectionSlug: 'getting-started',
          slug: 'quick-start',
          title: 'Quick start',
          summary: 'The fastest practical path to a running local stack and a successful first login.',
          seoTitle: 'Quick Start | dotblue Product Documentation',
          seoDescription:
            'Follow the practical dotblue quick start for Compose-based local bring-up, aligned runtime config generation, first login, and first assistant validation.',
          readingTime: '8 min read',
          sections: [
            {
              id: 'before-you-run',
              title: 'Before you run the stack',
              bullets: [
                'Prepare Docker and Docker Compose in the environment you actually use for local bring-up.',
                'Decide browser-facing public URLs before generating config, especially if you access the stack by host IP or WSL-exposed addresses.',
                'Prepare a usable admin account and one valid LLM API key so the first end-to-end test reaches a real model.',
              ],
              code: { language: 'bash', value: ENV_SAMPLE_CODE },
            },
            {
              id: 'compose-path',
              title: 'Bring up the stack with Compose',
              paragraphs: [
                'The local quick start is based on generated config plus a single Compose command. The important part is that Casdoor, backend, and web all use the same public URL strategy after config generation.',
              ],
              code: { language: 'bash', value: QUICK_START_CODE },
            },
            {
              id: 'windows-path',
              title: 'Windows path',
              paragraphs: [
                'If your local workflow is Windows-first, use the PowerShell prepare script, but keep the generated public URLs consistent with the address you will open in the browser.',
              ],
              code: { language: 'powershell', value: WINDOWS_QUICK_START_CODE },
            },
            {
              id: 'first-validation',
              title: 'Validate the first successful run',
              steps: [
                { title: 'Open `/en` or your localized home page', desc: 'Confirm the product home page loads through the browser-facing address you just configured.' },
                { title: 'Open the login flow', desc: 'Verify Casdoor is reachable and its branding assets load from the same public web domain strategy.' },
                { title: 'Complete sign-in', desc: 'Confirm callback success into the dashboard without host mismatch or redirect drift.' },
                { title: 'Create a first assistant', desc: 'If model choices are missing, go back and save the platform model first.' },
              ],
            },
          ],
        },
        {
          sectionSlug: 'getting-started',
          slug: 'login-and-authentication',
          title: 'Login and authentication',
          summary: 'How local sign-in works today, why registration is simplified by default, and how to expand it safely.',
          seoTitle: 'Login and Authentication | dotblue Docs',
          seoDescription:
            'Understand dotblue login flows with Casdoor, the default minimal registration path for local usage, and where to configure advanced sign-in and verification.',
          readingTime: '7 min read',
          sections: [
            {
              id: 'default-flow',
              title: 'Default local auth flow',
              paragraphs: [
                'The current local setup intentionally keeps registration minimal. That means sign-up focuses on Username, Display name, Password, and Confirm password so teams can bring up the stack without SMTP, SMS, or provider-specific verification dependencies.',
                'This keeps local validation simple: one stack, one login path, one callback path, and one browser-facing domain strategy.',
              ],
            },
            {
              id: 'why-simplified',
              title: 'Why local registration is simplified',
              bullets: [
                'Email verification requires SMTP delivery, sender configuration, templates, and message reachability checks.',
                'Phone verification requires SMS providers, templates, quotas, and failure handling.',
                'Teams validating product flow first usually need reliable sign-in more than advanced identity rollout on day one.',
              ],
            },
            {
              id: 'advanced-options',
              title: 'Advanced sign-in and sign-up options',
              note: 'Treat advanced identity rollout as a production-grade auth task, not a quick local default.',
              bullets: [
                'Enable email verification only when SMTP is configured and tested.',
                'Enable phone verification only when SMS delivery is part of your real rollout plan.',
                'Social login, WebAuthn, LDAP, or enterprise SSO should be staged and validated as part of a controlled rollout.',
              ],
              links: [
                { label: 'Casdoor Sign-up Items', url: CASDOOR_SIGNUP_ITEMS_URL, description: 'Configure registration fields and verification requirements.' },
                { label: 'Casdoor Sign-in Methods', url: CASDOOR_SIGNIN_METHODS_URL, description: 'Choose password, verification code, WebAuthn, LDAP, and other sign-in methods.' },
                { label: 'Casdoor Application Config', url: CASDOOR_APP_CONFIG_URL, description: 'Review redirect URLs, resend timeouts, and app-level auth behavior.' },
                { label: 'Casdoor Email Provider', url: CASDOOR_EMAIL_PROVIDER_URL, description: 'Configure SMTP so verification and password reset can actually deliver messages.' },
              ],
            },
          ],
        },
      ],
    },
    {
      slug: 'use-dotblue',
      title: 'Use dotblue',
      description: 'Move from basic access to real product operation: assistants, model settings, chat behavior, enterprise structure, and usage patterns.',
      articles: [
        {
          sectionSlug: 'use-dotblue',
          slug: 'assistants-and-workspaces',
          title: 'Assistants and workspaces',
          summary: 'How assistants, enterprise context, and user-facing workspaces fit together in daily product use.',
          seoTitle: 'Assistants and Workspaces | dotblue Docs',
          seoDescription:
            'Learn how dotblue structures assistants, workspaces, team boundaries, and first-run configuration for practical product usage.',
          readingTime: '6 min read',
          sections: [
            {
              id: 'assistant-model',
              title: 'How assistants are structured',
              paragraphs: [
                'Each assistant is a product surface with its own job, prompt, and runtime behavior. That means your first design decision should be scope, not model. Start with a narrow job and only add breadth when the workflow is stable.',
                'The assistant list is the operational center for creating, adjusting, and verifying these product surfaces before they are widely used.',
              ],
            },
            {
              id: 'workspace-boundaries',
              title: 'Workspace and organization boundaries',
              bullets: [
                'Use organizations and enterprise context when assistant access or configuration should differ by team or tenant.',
                'Keep platform-level settings for shared infrastructure decisions such as the default LLM provider.',
                'Use assistant-specific configuration for business behavior that should not affect other assistants.',
              ],
            },
            {
              id: 'first-assistant-guidance',
              title: 'What a good first assistant looks like',
              steps: [
                { title: 'Pick one clear business job', desc: 'Support lookup, sales qualification, or knowledge answering is a better first step than “general company agent”.' },
                { title: 'Write a narrow system prompt', desc: 'Tell the assistant exactly what it should do, what it should not do, and what kind of answer shape is expected.' },
                { title: 'Test in real chat', desc: 'Send a few high-signal queries and verify the assistant behaves predictably before expanding usage.' },
              ],
            },
          ],
        },
        {
          sectionSlug: 'use-dotblue',
          slug: 'providers-and-models',
          title: 'Providers and models',
          summary: 'How to think about model setup in dotblue and what usually causes missing or invalid model options.',
          seoTitle: 'Providers and Models | dotblue Docs',
          seoDescription:
            'Configure LLM providers and models in dotblue, understand platform-level settings, and avoid common issues when assistants cannot select or use models.',
          readingTime: '7 min read',
          sections: [
            {
              id: 'platform-models',
              title: 'Platform-level model configuration',
              paragraphs: [
                'dotblue expects model configuration to be available before assistants can use it. In practice, that means the platform or enterprise layer needs a valid provider setup before the assistant creation experience is complete.',
                'If assistants cannot see a model, the problem is usually not the assistant UI. It is usually missing provider credentials, wrong API base, or an unsaved model definition.',
              ],
            },
            {
              id: 'provider-checklist',
              title: 'Provider checklist',
              bullets: [
                'Provider type matches the actual API you are using.',
                'API base is reachable from the backend runtime.',
                'API key is valid and loaded into the real runtime environment.',
                'Model name matches a deployable or available model from the provider.',
                'Any existing runtime containers are recycled after major configuration changes.',
              ],
            },
            {
              id: 'failure-patterns',
              title: 'Common failure patterns',
              bullets: [
                'Model appears in config but assistants cannot use it: save scope mismatch or stale runtime state.',
                'Chat opens but no answer comes back: provider key or model name mismatch.',
                'Everything worked before a change: old runtime containers may still hold previous config.',
              ],
            },
          ],
        },
        {
          sectionSlug: 'use-dotblue',
          slug: 'chat-and-operations',
          title: 'Chat and daily operations',
          summary: 'What to verify inside chat, how to use it for first-run validation, and how operators should read failures.',
          seoTitle: 'Chat and Daily Operations | dotblue Docs',
          seoDescription:
            'Use dotblue Chat as an operational validation surface, understand first-run checks, and diagnose common response and runtime issues.',
          readingTime: '6 min read',
          sections: [
            {
              id: 'chat-role',
              title: 'Why chat is the operational proof point',
              paragraphs: [
                'Chat is where multiple parts of the product finally meet: authentication, assistant configuration, model setup, runtime behavior, and user-facing experience.',
                'That is why a successful chat exchange is one of the strongest first-run acceptance checks in dotblue.',
              ],
            },
            {
              id: 'daily-checks',
              title: 'Daily checks for operators',
              bullets: [
                'A new conversation can be created cleanly.',
                'The intended assistant is visible and selectable.',
                'The first reply arrives within an expected time window.',
                'Failures are diagnosable through the visible execution path or platform settings.',
              ],
            },
            {
              id: 'support-playbook',
              title: 'Basic support playbook',
              steps: [
                { title: 'Reproduce with a simple message', desc: 'Use a deterministic prompt rather than a broad or ambiguous user request.' },
                { title: 'Check model configuration', desc: 'Make sure the selected assistant is actually backed by a reachable and valid model.' },
                { title: 'Check runtime freshness', desc: 'If configuration changed recently, recycle stale runtime containers and retest.' },
                { title: 'Check auth and session continuity', desc: 'If chat opens strangely after login changes, validate callback, token handling, and redirect consistency.' },
              ],
            },
          ],
        },
      ],
    },
    {
      slug: 'advanced',
      title: 'Advanced',
      description: 'Go deeper into deployment strategy, production readiness, security boundaries, and operational reliability before broader rollout.',
      articles: [
        {
          sectionSlug: 'advanced',
          slug: 'deployment-architecture',
          title: 'Deployment architecture',
          summary: 'What the minimum stack contains, how the public URLs should align, and what belongs in generated config.',
          seoTitle: 'Deployment Architecture | dotblue Docs',
          seoDescription:
            'Understand the dotblue deployment architecture, public URL strategy, minimal service stack, and generated configuration alignment across web, backend, and Casdoor.',
          readingTime: '7 min read',
          sections: [
            {
              id: 'minimal-stack',
              title: 'Minimal service stack',
              paragraphs: [
                'A practical minimum deployment includes postgres, redis, casdoor, dotblue, and web. These services cover persistence, session and queue support, identity, backend APIs, and the browser-facing product surface.',
              ],
              code: { language: 'text', value: DOCKER_SERVICES_CODE },
            },
            {
              id: 'public-urls',
              title: 'Public URL strategy',
              bullets: [
                'Casdoor must use a browser-reachable public URL because the user-facing login flow lands there directly.',
                'The frontend public URL must match the URLs embedded into auth callback logic and branding assets.',
                'The backend public URL should reflect how browser calls and callback logic actually reach the API surface.',
              ],
            },
            {
              id: 'generated-config',
              title: 'Generated config is part of the product',
              paragraphs: [
                'Do not think of generated files as a side detail. In dotblue, generated runtime config is how public URL alignment, branding settings, and auth behavior stay consistent across services.',
                'If branding, callback URLs, or hostnames drift, regenerate config before debugging deeper application code.',
              ],
            },
          ],
        },
        {
          sectionSlug: 'advanced',
          slug: 'production-rollout',
          title: 'Production rollout',
          summary: 'How to move from local validation to a controlled production deployment with proper SEO, domains, auth, and ops discipline.',
          seoTitle: 'Production Rollout | dotblue Docs',
          seoDescription:
            'Plan a production rollout for dotblue with formal domains, HTTPS, reverse proxy, secret management, durable dependencies, and reliable operations.',
          readingTime: '8 min read',
          sections: [
            {
              id: 'production-basics',
              title: 'Production basics',
              paragraphs: [
                'Production rollout starts with stable public domains and disciplined configuration, not just containers that happen to be running. Users should see one coherent brand, one coherent auth path, and one coherent public URL strategy.',
              ],
              code: { language: 'text', value: REVERSE_PROXY_CHECKLIST },
            },
            {
              id: 'security-and-secrets',
              title: 'Security and secret handling',
              bullets: [
                'Use HTTPS for app, API, and auth.',
                'Inject provider keys, admin passwords, and other credentials through secret management rather than baked images.',
                'Back up databases and critical storage before exposing the environment to real users.',
                'Treat Casdoor branding and callback configuration as release-controlled assets.',
              ],
            },
            {
              id: 'seo-and-discoverability',
              title: 'SEO-friendly product docs and landing pages',
              paragraphs: [
                'For public documentation, stable article URLs matter. That is why each major docs topic should have a permanent page-level path rather than only an anchor on a large single page.',
                'Each article should expose its own title, description, canonical URL, and alternate language links so search engines can index it as an independent resource.',
              ],
            },
          ],
        },
        {
          sectionSlug: 'advanced',
          slug: 'troubleshooting-and-ops',
          title: 'Troubleshooting and operations',
          summary: 'The failure patterns most teams actually hit when moving from setup to daily operation.',
          seoTitle: 'Troubleshooting and Operations | dotblue Docs',
          seoDescription:
            'Troubleshoot common dotblue issues around auth redirects, empty dashboards, missing models, stale runtime behavior, and branding drift.',
          readingTime: '7 min read',
          sections: [
            {
              id: 'auth-issues',
              title: 'Auth and redirect issues',
              bullets: [
                'Login jumps to the wrong host: re-check public URLs and regenerate config.',
                'Callback succeeds but the session looks broken: verify callback path, browser-facing domain, and token persistence assumptions.',
                'Branding looks stale after updates: confirm the running config and browser cache are not serving old assets.',
              ],
            },
            {
              id: 'product-issues',
              title: 'Product and model issues',
              bullets: [
                'Dashboard is empty right after login: confirm initialization state and backend access to the database.',
                'Assistant creation has no model options: save the platform or enterprise model first.',
                'Chat still uses old behavior after config changes: recycle runtime containers and retest with a simple prompt.',
              ],
            },
            {
              id: 'ops-checklist',
              title: 'Pre-release operations checklist',
              bullets: [
                'Home, docs, login, dashboard, and chat all share aligned branding assets.',
                'Login, callback, registration, and logout all work from the same public domain strategy.',
                'The first assistant can be created and used in chat by a non-admin path if your rollout requires it.',
                'Monitoring covers auth failures, API errors, runtime health, and environment drift after deploys.',
              ],
            },
          ],
        },
      ],
    },
  ],
};

const chineseLibrary: DocsLibrary = {
  homeSeoTitle: 'dotblue 文档中心 | 快速上手、产品使用与部署指南',
  homeSeoDescription:
    '进入 dotblue 文档中心，查看按主题拆分的产品文档，包括快速上手、认证登录、助手使用、模型配置、部署架构与生产上线指南。',
  homeSeoKeywords:
    'dotblue 文档,产品文档,快速上手,企业级 AI 助手,部署指南,Casdoor 登录,模型配置,生产环境',
  eyebrow: 'PRODUCT DOCS',
  title: '像正式产品站一样阅读 dotblue 文档',
  subtitle:
    '从产品概览、快速上手、认证登录，到助手配置、聊天验证、部署架构和生产上线，按主题逐篇阅读，而不是在一个长页面里来回查找。',
  categoriesLabel: '文档分类',
  popularLabel: '热门文档',
  sectionDescriptionLabel: '本分类解决什么问题',
  quickLinksTitle: '快捷入口',
  quickLinks: buildQuickLinks('zh-CN'),
  repoTitle: '开源仓库与实现参考',
  repoDescription:
    '产品文档讲“怎么用”，官方仓库负责回答“底层怎么实现、配置模板长什么样、部署链路怎么串起来”。二者结合起来，才是完整可交付的资料体系。',
  sections: [
    {
      slug: 'getting-started',
      title: '快速开始',
      description: '先理解 dotblue 是什么、第一条成功链路应该长什么样，再用最短路径把本地环境拉起来并完成首次登录。',
      articles: [
        {
          sectionSlug: 'getting-started',
          slug: 'dotblue-overview',
          title: 'dotblue 概览',
          summary: '解释 dotblue 的产品定位、适用团队，以及什么才算真正“跑起来了”。',
          seoTitle: 'dotblue 概览 | dotblue 产品文档',
          seoDescription:
            '了解 dotblue 是什么、适合谁使用，以及它如何把企业级 AI 助手、认证登录、运行治理和部署能力整合成一个产品面。',
          readingTime: '6 分钟阅读',
          sections: [
            {
              id: 'what-it-is',
              title: 'dotblue 是什么',
              paragraphs: [
                'dotblue 是一层企业级 AI 助手交付产品面，不是单纯的聊天 UI，也不是单纯的模型封装。它把品牌化登录、平台级模型管理、助手生命周期、团队权限边界和运行时治理串成一个可交付产品。',
                '它的目标不是只让你“接上一个模型”，而是帮助产品、实施和平台团队把 AI 助手真正落到业务环境中，并能持续维护和扩展。',
              ],
            },
            {
              id: 'core-capabilities',
              title: '核心能力',
              bullets: [
                '通过 Casdoor 提供统一、品牌化的登录、回调和退出登录链路。',
                '支持助手创建、配置、迭代和验证，形成真正的产品面。',
                '支持平台级和企业级模型配置，便于统一治理模型入口。',
                '支持聊天页、对话状态和运行过程验证，方便联调与排障。',
                '提供本地拉起、测试验证和生产部署所需的配置与部署资产。',
              ],
            },
            {
              id: 'target-teams',
              title: '适合哪些团队',
              paragraphs: [
                'dotblue 适合需要快速上线内部助手、客户助手或行业场景助手的产品团队，也适合需要交付多租户或多组织环境的实施团队与平台团队。',
                '最好的起步方式不是一上来做“全能助手”，而是先挑一个边界清晰、价值明确的场景，例如客服、知识问答、销售辅助或内部运营支持。',
              ],
            },
            {
              id: 'first-success',
              title: '什么叫第一次成功',
              steps: [
                { title: '打开产品首页', desc: '确认首页、文档页和登录入口都通过同一套公开地址策略可访问。' },
                { title: '完成 Casdoor 登录', desc: '确认品牌化登录、回调和会话建立是通的。' },
                { title: '配置模型', desc: '至少保存一组平台级或企业级模型配置。' },
                { title: '创建一个助手', desc: '先做一个范围明确、可验证的助手，而不是一开始就做复杂大而全能力。' },
                { title: '进入 Chat 发消息', desc: '确认从前端到模型再到回复的完整链路是通的。' },
              ],
            },
          ],
        },
        {
          sectionSlug: 'getting-started',
          slug: 'quick-start',
          title: '快速上手',
          summary: '用最短的可执行路径把本地环境拉起来，并完成首次登录和首次助手验证。',
          seoTitle: '快速上手 | dotblue 产品文档',
          seoDescription:
            '通过 Compose、生成配置和第一次登录验证，快速完成 dotblue 的本地拉起和首条成功链路。',
          readingTime: '8 分钟阅读',
          sections: [
            {
              id: 'before-you-run',
              title: '启动前准备',
              bullets: [
                '准备可用的 Docker 与 Docker Compose 环境。',
                '在真正要访问页面的浏览器地址策略下，先确定 Web、Backend、Casdoor 的公开 URL。',
                '准备管理员账号以及至少一组可用的 LLM API Key，避免拉起后只能看到空壳页面。',
              ],
              code: { language: 'bash', value: ENV_SAMPLE_CODE },
            },
            {
              id: 'compose-path',
              title: 'Compose 启动路径',
              paragraphs: [
                '本地快速上手的关键不是“把容器都跑起来”，而是让 Casdoor、后端和前端在同一套公开地址上保持一致。真正决定这件事的，是 `prepare-config` 生成链而不是单个容器本身。',
              ],
              code: { language: 'bash', value: QUICK_START_CODE },
            },
            {
              id: 'windows-path',
              title: 'Windows 路径',
              paragraphs: [
                '如果你的工作流主要在 Windows 中完成，可以直接走 PowerShell 脚本，但仍要保证最终浏览器访问地址和生成配置中写入的公开 URL 完全一致。',
              ],
              code: { language: 'powershell', value: WINDOWS_QUICK_START_CODE },
            },
            {
              id: 'first-validation',
              title: '首次成功验证',
              steps: [
                { title: '打开产品首页', desc: '先确认首页和文档页正常可访问。' },
                { title: '进入登录流程', desc: '确认 Casdoor 页面能正确加载，并且品牌静态资源可访问。' },
                { title: '完成登录回调', desc: '确认能从 Casdoor 回跳到 Dashboard，而不是跳到错误域名或错误主机。' },
                { title: '创建第一个助手', desc: '如果看不到模型可选项，优先去平台设置里保存模型配置。' },
              ],
            },
          ],
        },
        {
          sectionSlug: 'getting-started',
          slug: 'login-and-authentication',
          title: '登录与认证',
          summary: '说明当前本地默认认证策略、高阶注册配置为什么默认不打开，以及如何安全扩展。',
          seoTitle: '登录与认证 | dotblue 产品文档',
          seoDescription:
            '了解 dotblue 当前的 Casdoor 登录方式、本地最简注册路径，以及如何为邮箱验证、短信验证和企业级认证扩展做准备。',
          readingTime: '7 分钟阅读',
          sections: [
            {
              id: 'default-flow',
              title: '当前默认认证路径',
              paragraphs: [
                '本地快速拉起默认采用最简注册路径，注册页面只保留 Username、Display name、Password、Confirm password，避免第一次启动就依赖 SMTP、短信服务和验证码模板。',
                '这样做的目标很明确：优先让团队把“注册成功 -> 登录成功 -> 回调成功 -> 进入产品”这条主链路跑通。',
              ],
            },
            {
              id: 'why-simplified',
              title: '为什么默认简化注册',
              bullets: [
                '邮箱验证依赖 SMTP 发信、模板配置、投递可达性和回执验证。',
                '短信验证依赖短信供应商、模板审核、额度和失败处理。',
                '大多数团队在本地验证阶段，优先需要的是稳定登录，而不是完整身份系统上线。',
              ],
            },
            {
              id: 'advanced-options',
              title: '高阶登录注册能力',
              note: '邮箱验证码、短信验证码、WebAuthn、社交登录和企业 SSO 都应该被当作正式认证项目来配置，而不是本地默认行为。',
              bullets: [
                '只有在 SMTP 完整配置并验证成功后，才建议开启邮箱验证注册。',
                '只有在短信链路纳入正式交付范围时，才建议开启手机验证码注册。',
                '社交登录、WebAuthn、LDAP 或企业级 SSO 更适合在测试环境和准生产环境中分阶段验证。',
              ],
              links: [
                { label: 'Casdoor 注册项配置', url: CASDOOR_SIGNUP_ITEMS_URL, description: '配置注册字段、邮箱验证规则和表单项。' },
                { label: 'Casdoor 登录方式配置', url: CASDOOR_SIGNIN_METHODS_URL, description: '配置 Password、验证码、WebAuthn、LDAP 等登录方式。' },
                { label: 'Casdoor 应用认证配置', url: CASDOOR_APP_CONFIG_URL, description: '查看回调地址、验证码重发超时和应用级认证设置。' },
                { label: 'Casdoor 邮件服务配置', url: CASDOOR_EMAIL_PROVIDER_URL, description: '配置 SMTP 邮件服务，支持邮箱验证码和找回密码。' },
              ],
            },
          ],
        },
      ],
    },
    {
      slug: 'use-dotblue',
      title: '使用 dotblue',
      description: '从能登录进入系统，走到真正会创建助手、配置模型、使用聊天页和理解日常运维动作。',
      articles: [
        {
          sectionSlug: 'use-dotblue',
          slug: 'assistants-and-workspaces',
          title: '助手与工作空间',
          summary: '解释助手、组织、企业上下文和工作空间是如何在产品里配合工作的。',
          seoTitle: '助手与工作空间 | dotblue 产品文档',
          seoDescription:
            '了解 dotblue 如何组织助手、工作空间和企业上下文，以及首次创建助手时应如何控制边界和范围。',
          readingTime: '6 分钟阅读',
          sections: [
            {
              id: 'assistant-model',
              title: '助手是一个产品面',
              paragraphs: [
                '在 dotblue 里，助手不是一个简单名字，而是一个有明确职责、提示词、模型配置和运行方式的产品面。',
                '因此第一次创建助手时，最重要的不是“模型多强”，而是边界是否清晰。范围越明确，第一版越容易成功。',
              ],
            },
            {
              id: 'workspace-boundaries',
              title: '工作空间与组织边界',
              bullets: [
                '当不同团队或不同租户需要不同配置时，应通过企业或组织上下文进行隔离。',
                '平台级设置更适合放共享基础设施决策，例如默认模型供应商。',
                '助手级配置更适合承载业务行为差异，不应随便污染其他助手。',
              ],
            },
            {
              id: 'first-assistant-guidance',
              title: '第一个助手怎么做更容易成功',
              steps: [
                { title: '先选一个具体业务任务', desc: '客服问答、知识查询或销售辅助都比“全能公司助手”更适合第一版。' },
                { title: '写一个边界清晰的系统提示词', desc: '要明确它该做什么、不该做什么，以及输出应该长什么样。' },
                { title: '直接在 Chat 中验证', desc: '用几个高价值问题验证结果，而不是只看列表页展示。' },
              ],
            },
          ],
        },
        {
          sectionSlug: 'use-dotblue',
          slug: 'providers-and-models',
          title: '模型供应商与模型配置',
          summary: '讲清楚为什么模型配置总是影响助手创建与聊天结果，以及常见失败点在哪里。',
          seoTitle: '模型供应商与模型配置 | dotblue 产品文档',
          seoDescription:
            '学习如何在 dotblue 中配置模型供应商、排查模型选项缺失，以及理解平台级模型配置的重要性。',
          readingTime: '7 分钟阅读',
          sections: [
            {
              id: 'platform-models',
              title: '平台级模型配置的重要性',
              paragraphs: [
                'dotblue 假设模型配置在助手使用前已经准备好。因此当你在助手创建页面看不到模型时，问题通常不在助手页面本身，而是平台级或企业级模型配置尚未完成。',
                '真正常见的问题是：API Key 无效、API Base 写错、模型名不可用，或者配置保存后运行时还没有刷新。',
              ],
            },
            {
              id: 'provider-checklist',
              title: '配置检查清单',
              bullets: [
                '供应商类型和实际 API 协议匹配。',
                'API Base 对后端运行环境可达。',
                'API Key 在真实运行环境中已注入并可用。',
                '模型名称对应的是当前供应商可以使用的真实模型。',
                '大改配置后，旧运行时容器已经刷新或重建。',
              ],
            },
            {
              id: 'failure-patterns',
              title: '常见失败模式',
              bullets: [
                '配置里能看到模型，但助手不能选：通常是保存范围不对或运行时状态过旧。',
                '聊天页能打开，但没有回复：通常是 API Key、API Base 或模型名不对。',
                '之前能用，改完配置就不行：通常是旧运行时仍在吃旧配置。',
              ],
            },
          ],
        },
        {
          sectionSlug: 'use-dotblue',
          slug: 'chat-and-operations',
          title: '聊天页与日常运维',
          summary: '把聊天页当成真正的验收和运维入口，而不是只是发消息的地方。',
          seoTitle: '聊天页与日常运维 | dotblue 产品文档',
          seoDescription:
            '了解如何通过 dotblue 聊天页做首次联调验收、日常检查和问题排查，确保助手真正可用。',
          readingTime: '6 分钟阅读',
          sections: [
            {
              id: 'chat-role',
              title: '为什么聊天页是最关键的验收面',
              paragraphs: [
                '聊天页是整个产品链路真正汇合的地方：认证、助手配置、模型配置、运行时行为和用户体验都会在这里暴露真实状态。',
                '因此，一次真正成功的聊天对话，比单独看登录页或助手列表页更能说明系统是否真的可用。',
              ],
            },
            {
              id: 'daily-checks',
              title: '日常检查重点',
              bullets: [
                '可以正常创建新会话。',
                '目标助手可见且可选。',
                '首条回复在合理时间内返回。',
                '出错时能通过平台配置或运行时线索定位问题。',
              ],
            },
            {
              id: 'support-playbook',
              title: '基础排障路径',
              steps: [
                { title: '先用简单问题复现', desc: '优先用可预测的简单提示词，而不是模糊的大任务。' },
                { title: '检查模型配置', desc: '确认当前助手绑定的模型真实可用。' },
                { title: '检查运行时是否过旧', desc: '刚改完模型或配置时，优先重建旧运行时后再测。' },
                { title: '检查认证与会话连续性', desc: '如果聊天页状态异常，也要回头检查登录回调和会话保持。' },
              ],
            },
          ],
        },
      ],
    },
    {
      slug: 'advanced',
      title: '高级主题',
      description: '覆盖部署架构、生产上线、安全与长期运维问题，帮助团队从“能跑”走向“能交付”。',
      articles: [
        {
          sectionSlug: 'advanced',
          slug: 'deployment-architecture',
          title: '部署架构',
          summary: '解释最小栈包含什么、公开地址应该怎样统一、生成配置为什么属于产品链路的一部分。',
          seoTitle: '部署架构 | dotblue 产品文档',
          seoDescription:
            '了解 dotblue 的最小部署栈、公开 URL 设计、配置生成链路，以及 web、backend、Casdoor 如何在部署架构上保持一致。',
          readingTime: '7 分钟阅读',
          sections: [
            {
              id: 'minimal-stack',
              title: '最小服务栈',
              paragraphs: [
                '一个真实可用的最小部署通常包含 postgres、redis、casdoor、dotblue、web。这五个服务分别承载数据持久化、运行支持、认证、后端 API 和前端产品面。',
              ],
              code: { language: 'text', value: DOCKER_SERVICES_CODE },
            },
            {
              id: 'public-urls',
              title: '公开地址策略',
              bullets: [
                'Casdoor 的公开 URL 必须是浏览器能直接访问到的地址。',
                '前端公开 URL 必须和登录回调、品牌资源引用使用同一套策略。',
                '后端公开 URL 应反映浏览器请求和回调流程实际访问的 API 面。'
              ],
            },
            {
              id: 'generated-config',
              title: '生成配置不是边角料',
              paragraphs: [
                '在 dotblue 里，生成配置文件并不是一个临时产物，而是确保公开地址、品牌资源和认证行为保持一致的关键环节。',
                '当你遇到品牌不一致、回调地址错误或页面跳错主机时，优先应该想到的是重新生成配置，而不是立刻怀疑前端页面本身。',
              ],
            },
          ],
        },
        {
          sectionSlug: 'advanced',
          slug: 'production-rollout',
          title: '生产上线',
          summary: '从本地验证走向正式交付时，哪些事情必须补齐，哪些才是真正的上线前提。',
          seoTitle: '生产上线 | dotblue 产品文档',
          seoDescription:
            '规划 dotblue 的生产上线流程，包括正式域名、HTTPS、反向代理、Secret 管理、持久化依赖和 SEO 友好的文档站策略。',
          readingTime: '8 分钟阅读',
          sections: [
            {
              id: 'production-basics',
              title: '生产上线的基本前提',
              paragraphs: [
                '生产上线并不是把容器从本地搬到服务器那么简单。真正重要的是：用户看到的是一套稳定域名、一套稳定认证路径和一套稳定品牌体验。',
              ],
              code: { language: 'text', value: REVERSE_PROXY_CHECKLIST },
            },
            {
              id: 'security-and-secrets',
              title: '安全与密钥管理',
              bullets: [
                'Web、Backend、Casdoor 全部使用 HTTPS。',
                'API Key、管理员密码和客户端密钥通过 Secret 管理注入。',
                '数据库与核心存储路径建立备份和恢复策略。',
                'Casdoor 品牌配置和回调配置纳入正式发布流程。',
              ],
            },
            {
              id: 'seo-and-discoverability',
              title: 'SEO 友好的文档策略',
              paragraphs: [
                '公开产品文档不应该只是一个超长单页。更好的方式是让每个主题都有自己的稳定 URL，这样搜索引擎才能把“概览”“快速上手”“部署架构”等主题分别索引。',
                '每篇文档都应有独立的标题、描述、canonical 和多语言 alternate，这样才更接近真正的产品文档站。',
              ],
            },
          ],
        },
        {
          sectionSlug: 'advanced',
          slug: 'troubleshooting-and-ops',
          title: '排障与运维',
          summary: '总结从环境拉起、登录回调到模型配置、品牌资源最容易踩到的实际问题。',
          seoTitle: '排障与运维 | dotblue 产品文档',
          seoDescription:
            '快速定位 dotblue 常见问题，包括登录跳错地址、控制台空白、模型缺失、运行时配置未刷新和品牌资源不一致。',
          readingTime: '7 分钟阅读',
          sections: [
            {
              id: 'auth-issues',
              title: '认证与跳转问题',
              bullets: [
                '登录跳到错误地址：优先检查 public URL 并重新生成配置。',
                '回调成功但会话异常：检查 callback 路径、浏览器访问域名和 token 处理策略。',
                '品牌资源似乎没有更新：确认运行时配置和浏览器缓存是否还在吃旧资源。',
              ],
            },
            {
              id: 'product-issues',
              title: '产品与模型问题',
              bullets: [
                '登录成功后控制台空白：检查初始化状态和后端数据库连接。',
                '创建助手没有模型可选：先保存平台级或企业级模型。',
                '改完配置聊天还不生效：重建运行时容器后再测。'
              ],
            },
            {
              id: 'ops-checklist',
              title: '上线前运维检查单',
              bullets: [
                '首页、文档页、登录页、控制台、聊天页都使用一致的品牌资源。',
                '登录、回调、注册、退出登录都在同一套公开地址策略下可正常工作。',
                '非管理员用户路径如果是交付范围，也应完成从登录到聊天的完整验证。',
                '监控覆盖认证失败、API 错误、运行时状态和发布后的环境漂移。',
              ],
            },
          ],
        },
      ],
    },
  ],
};

const libraries: Record<string, DocsLibrary> = {
  en: englishLibrary,
  'zh-CN': chineseLibrary,
};

export function getDocsLibrary(language: string): DocsLibrary {
  const resolved = resolveSupportedLanguage(language);
  const direct = libraries[resolved];
  if (direct) return direct;
  return {
    ...englishLibrary,
    quickLinks: buildQuickLinks(resolved),
  };
}

export function flattenDocsArticles(library: DocsLibrary) {
  return library.sections.flatMap((section) => section.articles.map((article) => ({ ...article, sectionTitle: section.title })));
}

export function findDocsArticle(language: string, sectionSlug: string, articleSlug: string) {
  const library = getDocsLibrary(language);
  for (const section of library.sections) {
    if (section.slug !== sectionSlug) continue;
    for (const article of section.articles) {
      if (article.slug === articleSlug) {
        return { section, article };
      }
    }
  }
  return null;
}

export function getDocsHref(language: string, sectionSlug?: string, articleSlug?: string) {
  if (!sectionSlug || !articleSlug) {
    return getLocalizedPath('/docs', language);
  }
  return getLocalizedPath(`/docs/${sectionSlug}/${articleSlug}`, language);
}

export function listDocsStaticPaths() {
  const seen = new Set<string>();
  const paths: Array<{ sectionSlug: string; articleSlug: string }> = [];

  for (const section of englishLibrary.sections) {
    for (const article of section.articles) {
      const key = `${section.slug}/${article.slug}`;
      if (seen.has(key)) continue;
      seen.add(key);
      paths.push({ sectionSlug: section.slug, articleSlug: article.slug });
    }
  }

  return paths;
}
