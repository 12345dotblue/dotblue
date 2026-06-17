# dotblue 站点统计集成方案

## 1. 方案概述

为 dotblue 官网（营销站 + 文档站）集成前端统计能力，追踪页面访问和关键转化事件，支撑 SEO 优化和产品增长分析。

### 1.1 目标

- 追踪所有公开页面的 PV/UV（首页、文档页、登录页、法律页）
- 追踪关键转化事件（注册、登录、文档阅读深度）
- 支持多语言维度的流量分析
- 支持 SPA 路由的自动 pageview 追踪
- 提供热力图和会话录制能力，分析用户行为路径
- 提供可切换的统计服务抽象层

### 1.2 技术栈现状

| 项目 | 详情 |
|------|------|
| React | 19.2.5 |
| TypeScript | 6.0.2 |
| Vite | 8.0.10 |
| 路由 | react-router-dom 7.14.2（BrowserRouter） |
| 国际化 | i18next，支持 6 种语言 |
| SEO | react-helmet-async |

### 1.3 当前状态

项目无任何统计代码，完全从零开始集成。

---

## 2. 方案选型

### 2.1 候选方案对比

| 方案 | 费用 | 中国可达 | 热力图/录制 | SPA 支持 | 隐私合规 | 推荐度 |
|------|------|----------|-------------|----------|----------|--------|
| **Microsoft Clarity** | 免费无限制 | ✅ 可达 | ✅ 原生 | 支持 | 自动匿名 | ⭐⭐⭐⭐⭐ |
| Google Analytics (GA4) | 免费 | ❌ 被墙 | ❌ 需第三方 | 原生支持 | 需 GDPR 处理 | ⭐⭐⭐ |
| Umami Cloud | $10/月起 | 取决于部署 | ✅ | 支持 | 隐私友好 | ⭐⭐⭐ |
| Plausible | $9/月起 | 取决于部署 | ❌ | 支持 | 极简合规 | ⭐⭐⭐ |
| PostHog | 免费额度 100K | ❌ 不稳定 | ✅ | 支持 | 可自托管 | ⭐⭐⭐ |
| 百度统计 | 免费 | ✅ 稳定 | ✅ | 支持 | 国内合规 | ⭐⭐ |

### 2.2 推荐方案：Microsoft Clarity

**选择理由：**
1. **全球可达**：`clarity.ms` 在中国大陆可正常访问，不会拖慢页面加载
2. **完全免费**：无流量限制、无站点数量限制，适合产品各阶段
3. **热力图 + 会话录制**：原生提供用户行为可视化分析，无需额外工具
4. **脚本轻量**：~20KB（GA4 约 45KB），对页面性能影响最小
5. **隐私友好**：自动匿名处理数据，无 Cookie 依赖，天然 GDPR 合规
6. **Bing SEO 生态**：与 Bing Webmaster Tools 集成，对 Bing 搜索优化有帮助

**注意事项：**
- Clarity 在流量来源分析深度上不如 GA4
- 无 Google Search Console / Google Ads 联动能力
- 如后续需要 GA4 的 SEO 生态能力，可并行添加，两者不冲突

---

## 3. 架构设计

### 3.1 分层架构

```
┌─────────────────────────────────────────────┐
│              业务层 (Components)              │
│  LandingPage / ProductDocsPage / Login ...  │
└─────────────────┬───────────────────────────┘
                  │ track('event_name', params)
┌─────────────────▼───────────────────────────┐
│           抽象层 (analytics.ts)              │
│  trackPageView / trackEvent / identify      │
└─────────────────┬───────────────────────────┘
                  │ 适配
┌─────────────────▼───────────────────────────┐
│       适配层 (clarity-adapter.ts)            │
│  clarity() 调用封装                          │
└─────────────────┬───────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│           clarity.js (外部脚本)               │
│        Microsoft Clarity                     │
└─────────────────────────────────────────────┘
```

### 3.2 核心模块

```
web/src/domains/analytics/
├── analytics.ts            # 统一接口，业务层调用
├── clarity-adapter.ts      # Clarity 适配器
├── types.ts                # 类型定义
├── events.ts               # 事件常量定义
├── useAnalytics.ts         # React Hook
└── AnalyticsProvider.tsx   # 路由监听 + 自动 pageview
```

### 3.3 环境变量

```bash
# .env.example
VITE_CLARITY_PROJECT_ID=        # Clarity Project ID，留空则禁用统计
```

---

## 4. 实现细节

### 4.1 类型定义

```typescript
// types.ts
export interface PageViewParams {
  path: string;
  title: string;
  language: string;
  referrer?: string;
}

export interface EventParams {
  [key: string]: string | number | boolean | undefined;
}

export interface AnalyticsAdapter {
  init: (projectId: string) => void;
  trackPageView: (params: PageViewParams) => void;
  trackEvent: (eventName: string, params?: EventParams) => void;
  identify?: (userId: string, traits?: Record<string, unknown>) => void;
}
```

### 4.2 统一接口

```typescript
// analytics.ts
import type { AnalyticsAdapter, EventParams, PageViewParams } from './types';

let adapter: AnalyticsAdapter | null = null;
let enabled = false;

export function initAnalytics(projectId: string) {
  if (!projectId) {
    console.info('[Analytics] Disabled: no project ID provided');
    return;
  }

  import('./clarity-adapter').then((mod) => {
    adapter = mod.createClarityAdapter();
    adapter.init(projectId);
    enabled = true;
    console.info('[Analytics] Initialized with project', projectId);
  });
}

export function trackPageView(params: PageViewParams) {
  if (!enabled || !adapter) return;
  adapter.trackPageView(params);
}

export function trackEvent(eventName: string, params?: EventParams) {
  if (!enabled || !adapter) return;
  adapter.trackEvent(eventName, params);
}

export function identify(userId: string, traits?: Record<string, unknown>) {
  if (!enabled || !adapter?.identify) return;
  adapter.identify(userId, traits);
}

export function isEnabled() {
  return enabled;
}
```

### 4.3 Clarity 适配器

```typescript
// clarity-adapter.ts
import type { AnalyticsAdapter, EventParams, PageViewParams } from './types';

declare global {
  interface Window {
    clarity: (...args: unknown[]) => void;
  }
}

function loadClarityScript(projectId: string) {
  (function (c: Document, l: string, a: string, r: string, i: string) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const w = window as any;
    w[a] = w[a] || function () {
      (w[a].q = w[a].q || []).push(arguments);
    };
    const script = c.createElement(l) as HTMLScriptElement;
    script.async = true;
    script.src = `https://www.clarity.ms/tag/${i}`;
    const firstScript = c.getElementsByTagName(l)[0];
    firstScript.parentNode?.insertBefore(script, firstScript);
  })(document, 'script', 'clarity', 'script', projectId);
}

export function createClarityAdapter(): AnalyticsAdapter {
  return {
    init(projectId: string) {
      loadClarityScript(projectId);
    },

    trackPageView(params: PageViewParams) {
      window.clarity('set', 'page_language', params.language);
      if (params.referrer) {
        window.clarity('set', 'referrer', params.referrer);
      }
    },

    trackEvent(eventName: string, params?: EventParams) {
      window.clarity('event', eventName, params);
    },

    identify(userId: string, traits?: Record<string, unknown>) {
      window.clarity('identify', userId, traits);
    },
  };
}
```

### 4.4 事件常量

```typescript
// events.ts
export const ANALYTICS_EVENTS = {
  LOGIN_START: 'login_start',
  LOGIN_SUCCESS: 'login_success',
  SIGNUP_START: 'signup_start',
  SIGNUP_SUCCESS: 'signup_success',
  LOGOUT: 'logout',

  DOC_PAGE_VIEW: 'doc_page_view',
  DOC_SECTION_CLICK: 'doc_section_click',
  DOC_EXTERNAL_LINK_CLICK: 'doc_external_link_click',

  NAV_CLICK: 'nav_click',
  LANGUAGE_SWITCH: 'language_switch',

  CTA_CLICK: 'cta_click',
  PRICING_PLAN_CLICK: 'pricing_plan_click',

  ERROR_PAGE_VIEW: 'error_page_view',
} as const;

export type AnalyticsEventName = (typeof ANALYTICS_EVENTS)[keyof typeof ANALYTICS_EVENTS];
```

### 4.5 路由监听 + 自动 PageView

```typescript
// AnalyticsProvider.tsx
import { useEffect, useRef } from 'react';
import { useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { trackPageView } from './analytics';

export function AnalyticsProvider() {
  const location = useLocation();
  const { i18n } = useTranslation();
  const previousPath = useRef(location.pathname);

  useEffect(() => {
    if (location.pathname !== previousPath.current) {
      previousPath.current = location.pathname;

      trackPageView({
        path: location.pathname + location.search,
        title: document.title,
        language: i18n.language,
      });
    }
  }, [location, i18n.language]);

  return null;
}
```

### 4.6 React Hook

```typescript
// useAnalytics.ts
import { useCallback } from 'react';
import { trackEvent, identify } from './analytics';
import type { EventParams } from './types';

export function useAnalytics() {
  const track = useCallback((eventName: string, params?: EventParams) => {
    trackEvent(eventName, params);
  }, []);

  const setIdentify = useCallback((userId: string, traits?: Record<string, unknown>) => {
    identify(userId, traits);
  }, []);

  return { track, identify: setIdentify };
}
```

---

## 5. 集成点

### 5.1 初始化入口

```typescript
// main.tsx 或 App.tsx
import { initAnalytics } from './domains/analytics/analytics';

const clarityId = import.meta.env.VITE_CLARITY_PROJECT_ID;
if (clarityId) {
  initAnalytics(clarityId);
}
```

### 5.2 路由监听挂载

```tsx
// App.tsx
import { AnalyticsProvider } from './domains/analytics/AnalyticsProvider';

function App() {
  return (
    <Router>
      <AnalyticsProvider />
      {/* ... routes ... */}
    </Router>
  );
}
```

### 5.3 登录成功后识别用户

```typescript
// Login.tsx 或 auth 回调处
import { identify } from './domains/analytics/analytics';
import { ANALYTICS_EVENTS } from './domains/analytics/events';

identify(user.id, { username: user.username });
trackEvent(ANALYTICS_EVENTS.LOGIN_SUCCESS, { method: 'casdoor' });
```

### 5.4 文档页追踪

```typescript
// ProductDocsPage.tsx
import { trackEvent } from './domains/analytics/analytics';
import { ANALYTICS_EVENTS } from './domains/analytics/events';

useEffect(() => {
  if (article) {
    trackEvent(ANALYTICS_EVENTS.DOC_PAGE_VIEW, {
      doc_slug: article.slug,
      section_slug: article.sectionSlug,
      language: currentLanguage,
    });
  }
}, [article, currentLanguage]);
```

---

## 6. 追踪事件清单

### 6.1 自动追踪（无需手动埋点）

| 事件 | 触发时机 | 说明 |
|------|----------|------|
| Page View | 每次路由切换 | Clarity 自动捕获，同时更新 language 自定义属性 |
| 热力图 | 持续采集 | 点击热力图、滚动热力图自动生成 |
| 会话录制 | 自动采样 | Clarity 自动录制部分用户会话 |

### 6.2 手动追踪事件

| 事件名 | 触发场景 | 参数 | 优先级 |
|--------|----------|------|--------|
| `login_start` | 点击登录按钮 | source | P0 |
| `login_success` | 登录成功回调 | method | P0 |
| `signup_start` | 点击注册按钮 | source | P0 |
| `signup_success` | 注册成功 | method | P0 |
| `doc_page_view` | 文档文章页加载 | doc_slug, section_slug, language | P0 |
| `language_switch` | 切换语言 | from_language, to_language | P1 |
| `nav_click` | 点击顶部导航 | nav_item | P1 |
| `cta_click` | 点击 CTA 按钮 | cta_name, location | P1 |
| `pricing_plan_click` | 点击定价方案 | plan_name | P2 |
| `doc_external_link_click` | 点击文档外链 | url, label | P2 |
| `error_page_view` | 访问 404 页面 | path | P2 |

---

## 7. 隐私合规

### 7.1 Clarity 隐私特性

Clarity 天然具备较好的隐私保护能力：
- **无 Cookie 依赖**：不使用第三方 Cookie 追踪用户
- **自动匿名**：默认对敏感数据（密码、信用卡号等）自动脱敏
- **无广告追踪**：不将数据用于广告个性化
- **GDPR 友好**：微软承诺遵守 GDPR，数据处理协议公开透明

### 7.2 可选配置

如需进一步控制，可在 Clarity 后台设置：
- 自定义脱敏规则（隐藏特定 DOM 元素内容）
- 设置数据保留期限
- 排除特定页面不录制

### 7.3 环境变量控制

- `VITE_CLARITY_PROJECT_ID` 为空时，不加载任何统计代码
- 本地开发环境默认不启用统计
- 可通过运行时配置动态开关

---

## 8. 部署配置

### 8.1 环境变量

```bash
# .env.production
VITE_CLARITY_PROJECT_ID=XXXXXXXXXX
```

### 8.2 Docker 构建

```dockerfile
# web/Dockerfile
ARG VITE_CLARITY_PROJECT_ID
```

构建时传入：
```bash
docker build \
  --build-arg VITE_CLARITY_PROJECT_ID=XXXXXXXXXX \
  -t dotblue-web:prod \
  ./web
```

### 8.3 逐服务部署

```bash
docker run -d \
  --name dotblue-web \
  --build-arg VITE_CLARITY_PROJECT_ID=XXXXXXXXXX \
  -p 19000:80 \
  dotblue-web:latest
```

---

## 9. 验证清单

### 9.1 开发环境验证

- [ ] 设置 `VITE_CLARITY_PROJECT_ID` 后，`npm run dev` 启动
- [ ] 打开浏览器 DevTools → Network，确认 `clarity.ms/tag/` 脚本加载
- [ ] 确认 `window.clarity` 函数存在
- [ ] 切换路由，确认 Clarity 后台 Dashboard 有 PV 记录
- [ ] 点击登录，确认自定义事件在 Clarity 中可见

### 9.2 生产环境验证

- [ ] 部署后访问 Clarity 后台 Dashboard，确认有流量
- [ ] 访问文档页，确认 `doc_page_view` 事件可见
- [ ] 切换语言，确认 `page_language` 属性正确
- [ ] 确认热力图正常生成
- [ ] 确认会话录制正常工作
- [ ] 确认 `VITE_CLARITY_PROJECT_ID` 为空时不加载 Clarity 代码

---

## 10. 后续演进

### 10.1 短期（1-2 周）

- 完成 Clarity 基础集成
- 实现自动 pageview 追踪
- 添加 P0 事件埋点
- 配置 Clarity 后台的自定义脱敏规则

### 10.2 中期（1-2 月）

- 利用热力图分析首页和文档页的用户行为
- 通过会话录制发现 UX 问题
- 配置转化漏斗（注册 → 登录 → 创建助手）
- 添加文档阅读深度追踪

### 10.3 长期

- 如需要 Google 生态联动（Search Console / Ads），可并行添加 GA4
- 后端 API 调用统计
- 用户行为路径分析

---

## 11. 文件清单

| 文件 | 说明 |
|------|------|
| `src/domains/analytics/types.ts` | 类型定义 |
| `src/domains/analytics/events.ts` | 事件常量 |
| `src/domains/analytics/analytics.ts` | 统一接口 |
| `src/domains/analytics/clarity-adapter.ts` | Clarity 适配器 |
| `src/domains/analytics/AnalyticsProvider.tsx` | 路由监听 |
| `src/domains/analytics/useAnalytics.ts` | React Hook |
| `src/App.tsx` | 挂载 AnalyticsProvider |
| `src/main.tsx` | 初始化 analytics |
| `web/.env.example` | 添加 `VITE_CLARITY_PROJECT_ID` |
| `web/Dockerfile` | 添加 build arg |
