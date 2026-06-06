import type { DocsLibrary } from '../schema';

export const library: DocsLibrary = {
  "requestedLanguage": "en",
  "contentLanguage": "en",
  "isFallbackContent": false,
  "homeSeoTitle": "dotblue Docs | Getting Started, Product Guides, and Deployment",
  "homeSeoDescription": "Explore dotblue product documentation with structured guides for getting started, workspace operations, authentication, deployment, and production rollout.",
  "homeSeoKeywords": "dotblue docs, enterprise AI assistant docs, getting started, deployment guide, Casdoor login, self-hosted AI workspace, production rollout",
  "eyebrow": "PRODUCT DOCS",
  "title": "dotblue docs for fast setup, rollout, and production delivery",
  "subtitle": "Go straight from first install to login, assistant setup, deployment, and production rollout without digging through a single oversized page.",
  "categoriesLabel": "Documentation categories",
  "popularLabel": "Popular pages",
  "sectionDescriptionLabel": "What you will learn",
  "quickLinksTitle": "Quick links",
  "quickLinks": [
    {
      "label": "Open product home",
      "url": "/en"
    },
    {
      "label": "Open sign-in",
      "url": "/en/login"
    },
    {
      "label": "GitHub",
      "url": "https://github.com/12345dotblue/dotblue",
      "description": "Inspect source code, deployment assets, and issues."
    }
  ],
  "repoTitle": "Open source and implementation reference",
  "repoDescription": "The official repository is still the best place to inspect deployment templates, understand implementation details, and connect the product docs to the actual codebase.",
  "sections": [
    {
      "slug": "getting-started",
      "title": "Getting started",
      "description": "Understand what dotblue is, how the first successful setup looks, and how to get from first boot to first working assistant quickly.",
      "articles": [
        {
          "sectionSlug": "getting-started",
          "slug": "dotblue-overview",
          "title": "dotblue overview",
          "summary": "What dotblue is, who it is built for, and what a successful rollout usually looks like.",
          "seoTitle": "dotblue Overview | Product Documentation",
          "seoDescription": "Learn what dotblue is, how enterprise teams use it, and how it connects productized assistants, authentication, runtime operations, and deployment.",
          "readingTime": "6 min read",
          "sections": [
            {
              "id": "what-it-is",
              "title": "What dotblue is",
              "paragraphs": [
                "dotblue is an enterprise AI assistant delivery surface. It is not just a chat UI and not just a model wrapper. It brings together branded sign-in, platform-level model configuration, assistant management, team-oriented access control, and runtime operations in one product experience.",
                "The product is designed for teams that need to move from idea to deployable assistant experience quickly, while still keeping enough control to support real environments, multiple users, and organization boundaries."
              ]
            },
            {
              "id": "core-capabilities",
              "title": "Core capabilities",
              "bullets": [
                "Branded authentication through Casdoor, including callback and logout alignment.",
                "Assistant lifecycle management with prompt, model, and runtime settings.",
                "Platform and enterprise model configuration for shared LLM governance.",
                "Chat surfaces with execution visibility and conversation continuity.",
                "Deployment assets that support local bring-up, staging validation, and production rollout."
              ]
            },
            {
              "id": "who-uses-it",
              "title": "Who this product is for",
              "paragraphs": [
                "dotblue fits product teams launching internal assistants, implementation teams delivering customer environments, and platform teams standardizing AI assistant rollout across organizations.",
                "The best first use case is a focused assistant with clear business value, such as customer support, knowledge lookup, sales copilot, or internal operations assistance."
              ]
            },
            {
              "id": "first-success",
              "title": "What first success looks like",
              "steps": [
                {
                  "title": "Open the product site",
                  "desc": "Confirm the localized home page and documentation are available from the same public URL strategy."
                },
                {
                  "title": "Sign in through Casdoor",
                  "desc": "Verify the branded login flow, callback, and session establishment into the dashboard."
                },
                {
                  "title": "Configure a model",
                  "desc": "Save at least one platform or enterprise LLM model so assistants can respond."
                },
                {
                  "title": "Create an assistant",
                  "desc": "Define one assistant with a narrow job, a clear system prompt, and a predictable output expectation."
                },
                {
                  "title": "Open Chat and send a message",
                  "desc": "Confirm the user-facing conversation flow and runtime behavior end to end."
                }
              ]
            }
          ]
        },
        {
          "sectionSlug": "getting-started",
          "slug": "quick-start",
          "title": "Quick start",
          "summary": "The fastest practical path to a running local stack and a successful first login.",
          "seoTitle": "Quick Start | dotblue Product Documentation",
          "seoDescription": "Follow the practical dotblue quick start for Compose-based local bring-up, aligned runtime config generation, first login, and first assistant validation.",
          "readingTime": "8 min read",
          "sections": [
            {
              "id": "before-you-run",
              "title": "Before you run the stack",
              "bullets": [
                "Prepare Docker and Docker Compose in the environment you actually use for local bring-up.",
                "Decide browser-facing public URLs before generating config, especially if you access the stack by host IP or WSL-exposed addresses.",
                "Prepare a usable admin account and one valid LLM API key so the first end-to-end test reaches a real model."
              ],
              "code": {
                "language": "bash",
                "value": "CASDOOR_PUBLIC_URL=https://auth.example.com\nDOTBLUE_PUBLIC_URL=https://app.example.com\nDOTBLUE_BACKEND_PUBLIC_URL=https://api.example.com\n\nDOTBLUE_ADMIN_USERNAME=admin\nDOTBLUE_ADMIN_EMAIL=admin@example.com\nDOTBLUE_ADMIN_PASSWORD=replace-with-a-strong-password\n\nDOTBLUE_LLM_PROVIDER_TYPE=openai\nDOTBLUE_LLM_API_BASE=https://api.openai.com/v1\nDOTBLUE_LLM_API_KEY=replace-with-provider-key\nDOTBLUE_LLM_MODEL=gpt-4.1-mini"
              }
            },
            {
              "id": "compose-path",
              "title": "Bring up the stack with Compose",
              "paragraphs": [
                "The local quick start is based on generated config plus a single Compose command. The important part is that Casdoor, backend, and web all use the same public URL strategy after config generation."
              ],
              "code": {
                "language": "bash",
                "value": "cd deploy/compose\ncp .env.example .env\n./prepare-config.sh\ndocker compose up -d --build"
              }
            },
            {
              "id": "windows-path",
              "title": "Windows path",
              "paragraphs": [
                "If your local workflow is Windows-first, use the PowerShell prepare script, but keep the generated public URLs consistent with the address you will open in the browser."
              ],
              "code": {
                "language": "powershell",
                "value": "cd deploy\\compose\ncopy .env.example .env\n.\\prepare-config.ps1\ndocker compose up -d --build"
              }
            },
            {
              "id": "first-validation",
              "title": "Validate the first successful run",
              "steps": [
                {
                  "title": "Open `/en` or your localized home page",
                  "desc": "Confirm the product home page loads through the browser-facing address you just configured."
                },
                {
                  "title": "Open the login flow",
                  "desc": "Verify Casdoor is reachable and its branding assets load from the same public web domain strategy."
                },
                {
                  "title": "Complete sign-in",
                  "desc": "Confirm callback success into the dashboard without host mismatch or redirect drift."
                },
                {
                  "title": "Create a first assistant",
                  "desc": "If model choices are missing, go back and save the platform model first."
                }
              ]
            }
          ]
        },
        {
          "sectionSlug": "getting-started",
          "slug": "login-and-authentication",
          "title": "Login and authentication",
          "summary": "How local sign-in works today, why registration is simplified by default, and how to expand it safely.",
          "seoTitle": "Login and Authentication | dotblue Docs",
          "seoDescription": "Understand dotblue login flows with Casdoor, the default minimal registration path for local usage, and where to configure advanced sign-in and verification.",
          "readingTime": "7 min read",
          "sections": [
            {
              "id": "default-flow",
              "title": "Default local auth flow",
              "paragraphs": [
                "The current local setup intentionally keeps registration minimal. That means sign-up focuses on Username, Display name, Password, and Confirm password so teams can bring up the stack without SMTP, SMS, or provider-specific verification dependencies.",
                "This keeps local validation simple: one stack, one login path, one callback path, and one browser-facing domain strategy."
              ]
            },
            {
              "id": "why-simplified",
              "title": "Why local registration is simplified",
              "bullets": [
                "Email verification requires SMTP delivery, sender configuration, templates, and message reachability checks.",
                "Phone verification requires SMS providers, templates, quotas, and failure handling.",
                "Teams validating product flow first usually need reliable sign-in more than advanced identity rollout on day one."
              ]
            },
            {
              "id": "advanced-options",
              "title": "Advanced sign-in and sign-up options",
              "note": "Treat advanced identity rollout as a production-grade auth task, not a quick local default.",
              "bullets": [
                "Enable email verification only when SMTP is configured and tested.",
                "Enable phone verification only when SMS delivery is part of your real rollout plan.",
                "Social login, WebAuthn, LDAP, or enterprise SSO should be staged and validated as part of a controlled rollout."
              ],
              "links": [
                {
                  "label": "Casdoor Sign-up Items",
                  "url": "https://casdoor.ai/docs/application/signup-items-table",
                  "description": "Configure registration fields and verification requirements."
                },
                {
                  "label": "Casdoor Sign-in Methods",
                  "url": "https://casdoor.ai/docs/application/signin-methods",
                  "description": "Choose password, verification code, WebAuthn, LDAP, and other sign-in methods."
                },
                {
                  "label": "Casdoor Application Config",
                  "url": "https://casdoor.ai/docs/application/config",
                  "description": "Review redirect URLs, resend timeouts, and app-level auth behavior."
                },
                {
                  "label": "Casdoor Email Provider",
                  "url": "https://casdoor.ai/docs/provider/email/overview",
                  "description": "Configure SMTP so verification and password reset can actually deliver messages."
                }
              ]
            }
          ]
        }
      ]
    },
    {
      "slug": "use-dotblue",
      "title": "Use dotblue",
      "description": "Move from basic access to real product operation: assistants, model settings, chat behavior, enterprise structure, and usage patterns.",
      "articles": [
        {
          "sectionSlug": "use-dotblue",
          "slug": "assistants-and-workspaces",
          "title": "Assistants and workspaces",
          "summary": "How assistants, enterprise context, and user-facing workspaces fit together in daily product use.",
          "seoTitle": "Assistants and Workspaces | dotblue Docs",
          "seoDescription": "Learn how dotblue structures assistants, workspaces, team boundaries, and first-run configuration for practical product usage.",
          "readingTime": "6 min read",
          "sections": [
            {
              "id": "assistant-model",
              "title": "How assistants are structured",
              "paragraphs": [
                "Each assistant is a product surface with its own job, prompt, and runtime behavior. That means your first design decision should be scope, not model. Start with a narrow job and only add breadth when the workflow is stable.",
                "The assistant list is the operational center for creating, adjusting, and verifying these product surfaces before they are widely used."
              ]
            },
            {
              "id": "workspace-boundaries",
              "title": "Workspace and organization boundaries",
              "bullets": [
                "Use organizations and enterprise context when assistant access or configuration should differ by team or tenant.",
                "Keep platform-level settings for shared infrastructure decisions such as the default LLM provider.",
                "Use assistant-specific configuration for business behavior that should not affect other assistants."
              ]
            },
            {
              "id": "first-assistant-guidance",
              "title": "What a good first assistant looks like",
              "steps": [
                {
                  "title": "Pick one clear business job",
                  "desc": "Support lookup, sales qualification, or knowledge answering is a better first step than “general company agent”."
                },
                {
                  "title": "Write a narrow system prompt",
                  "desc": "Tell the assistant exactly what it should do, what it should not do, and what kind of answer shape is expected."
                },
                {
                  "title": "Test in real chat",
                  "desc": "Send a few high-signal queries and verify the assistant behaves predictably before expanding usage."
                }
              ]
            }
          ]
        },
        {
          "sectionSlug": "use-dotblue",
          "slug": "providers-and-models",
          "title": "Providers and models",
          "summary": "How to think about model setup in dotblue and what usually causes missing or invalid model options.",
          "seoTitle": "Providers and Models | dotblue Docs",
          "seoDescription": "Configure LLM providers and models in dotblue, understand platform-level settings, and avoid common issues when assistants cannot select or use models.",
          "readingTime": "7 min read",
          "sections": [
            {
              "id": "platform-models",
              "title": "Platform-level model configuration",
              "paragraphs": [
                "dotblue expects model configuration to be available before assistants can use it. In practice, that means the platform or enterprise layer needs a valid provider setup before the assistant creation experience is complete.",
                "If assistants cannot see a model, the problem is usually not the assistant UI. It is usually missing provider credentials, wrong API base, or an unsaved model definition."
              ]
            },
            {
              "id": "provider-checklist",
              "title": "Provider checklist",
              "bullets": [
                "Provider type matches the actual API you are using.",
                "API base is reachable from the backend runtime.",
                "API key is valid and loaded into the real runtime environment.",
                "Model name matches a deployable or available model from the provider.",
                "Any existing runtime containers are recycled after major configuration changes."
              ]
            },
            {
              "id": "failure-patterns",
              "title": "Common failure patterns",
              "bullets": [
                "Model appears in config but assistants cannot use it: save scope mismatch or stale runtime state.",
                "Chat opens but no answer comes back: provider key or model name mismatch.",
                "Everything worked before a change: old runtime containers may still hold previous config."
              ]
            }
          ]
        },
        {
          "sectionSlug": "use-dotblue",
          "slug": "chat-and-operations",
          "title": "Chat and daily operations",
          "summary": "What to verify inside chat, how to use it for first-run validation, and how operators should read failures.",
          "seoTitle": "Chat and Daily Operations | dotblue Docs",
          "seoDescription": "Use dotblue Chat as an operational validation surface, understand first-run checks, and diagnose common response and runtime issues.",
          "readingTime": "6 min read",
          "sections": [
            {
              "id": "chat-role",
              "title": "Why chat is the operational proof point",
              "paragraphs": [
                "Chat is where multiple parts of the product finally meet: authentication, assistant configuration, model setup, runtime behavior, and user-facing experience.",
                "That is why a successful chat exchange is one of the strongest first-run acceptance checks in dotblue."
              ]
            },
            {
              "id": "daily-checks",
              "title": "Daily checks for operators",
              "bullets": [
                "A new conversation can be created cleanly.",
                "The intended assistant is visible and selectable.",
                "The first reply arrives within an expected time window.",
                "Failures are diagnosable through the visible execution path or platform settings."
              ]
            },
            {
              "id": "support-playbook",
              "title": "Basic support playbook",
              "steps": [
                {
                  "title": "Reproduce with a simple message",
                  "desc": "Use a deterministic prompt rather than a broad or ambiguous user request."
                },
                {
                  "title": "Check model configuration",
                  "desc": "Make sure the selected assistant is actually backed by a reachable and valid model."
                },
                {
                  "title": "Check runtime freshness",
                  "desc": "If configuration changed recently, recycle stale runtime containers and retest."
                },
                {
                  "title": "Check auth and session continuity",
                  "desc": "If chat opens strangely after login changes, validate callback, token handling, and redirect consistency."
                }
              ]
            }
          ]
        }
      ]
    },
    {
      "slug": "advanced",
      "title": "Advanced",
      "description": "Go deeper into deployment strategy, production readiness, security boundaries, and operational reliability before broader rollout.",
      "articles": [
        {
          "sectionSlug": "advanced",
          "slug": "deployment-architecture",
          "title": "Deployment architecture",
          "summary": "What the minimum stack contains, how the public URLs should align, and what belongs in generated config.",
          "seoTitle": "Deployment Architecture | dotblue Docs",
          "seoDescription": "Understand the dotblue deployment architecture, public URL strategy, minimal service stack, and generated configuration alignment across web, backend, and Casdoor.",
          "readingTime": "7 min read",
          "sections": [
            {
              "id": "minimal-stack",
              "title": "Minimal service stack",
              "paragraphs": [
                "A practical minimum deployment includes postgres, redis, casdoor, dotblue, and web. These services cover persistence, session and queue support, identity, backend APIs, and the browser-facing product surface."
              ],
              "code": {
                "language": "text",
                "value": "Services\n- postgres\n- redis\n- casdoor\n- dotblue\n- web\n\nBrowser-facing ports\n- Web: 19000\n- Backend: 18080\n- Casdoor: 18000"
              }
            },
            {
              "id": "public-urls",
              "title": "Public URL strategy",
              "bullets": [
                "Casdoor must use a browser-reachable public URL because the user-facing login flow lands there directly.",
                "The frontend public URL must match the URLs embedded into auth callback logic and branding assets.",
                "The backend public URL should reflect how browser calls and callback logic actually reach the API surface."
              ]
            },
            {
              "id": "generated-config",
              "title": "Generated config is part of the product",
              "paragraphs": [
                "Do not think of generated files as a side detail. In dotblue, generated runtime config is how public URL alignment, branding settings, and auth behavior stay consistent across services.",
                "If branding, callback URLs, or hostnames drift, regenerate config before debugging deeper application code."
              ]
            }
          ]
        },
        {
          "sectionSlug": "advanced",
          "slug": "production-rollout",
          "title": "Production rollout",
          "summary": "How to move from local validation to a controlled production deployment with proper SEO, domains, auth, and ops discipline.",
          "seoTitle": "Production Rollout | dotblue Docs",
          "seoDescription": "Plan a production rollout for dotblue with formal domains, HTTPS, reverse proxy, secret management, durable dependencies, and reliable operations.",
          "readingTime": "8 min read",
          "sections": [
            {
              "id": "production-basics",
              "title": "Production basics",
              "paragraphs": [
                "Production rollout starts with stable public domains and disciplined configuration, not just containers that happen to be running. Users should see one coherent brand, one coherent auth path, and one coherent public URL strategy."
              ],
              "code": {
                "language": "text",
                "value": "Deployment checkpoints\n1. Use formal domains for app, API, and auth\n2. Terminate TLS at a trusted reverse proxy\n3. Forward X-Forwarded-* headers correctly\n4. Keep internal container addresses separate from public URLs\n5. Rotate admin and provider secrets before launch"
              }
            },
            {
              "id": "security-and-secrets",
              "title": "Security and secret handling",
              "bullets": [
                "Use HTTPS for app, API, and auth.",
                "Inject provider keys, admin passwords, and other credentials through secret management rather than baked images.",
                "Back up databases and critical storage before exposing the environment to real users.",
                "Treat Casdoor branding and callback configuration as release-controlled assets."
              ]
            },
            {
              "id": "seo-and-discoverability",
              "title": "SEO-friendly product docs and landing pages",
              "paragraphs": [
                "For public documentation, stable article URLs matter. That is why each major docs topic should have a permanent page-level path rather than only an anchor on a large single page.",
                "Each article should expose its own title, description, canonical URL, and alternate language links so search engines can index it as an independent resource."
              ]
            }
          ]
        },
        {
          "sectionSlug": "advanced",
          "slug": "troubleshooting-and-ops",
          "title": "Troubleshooting and operations",
          "summary": "The failure patterns most teams actually hit when moving from setup to daily operation.",
          "seoTitle": "Troubleshooting and Operations | dotblue Docs",
          "seoDescription": "Troubleshoot common dotblue issues around auth redirects, empty dashboards, missing models, stale runtime behavior, and branding drift.",
          "readingTime": "7 min read",
          "sections": [
            {
              "id": "auth-issues",
              "title": "Auth and redirect issues",
              "bullets": [
                "Login jumps to the wrong host: re-check public URLs and regenerate config.",
                "Callback succeeds but the session looks broken: verify callback path, browser-facing domain, and token persistence assumptions.",
                "Branding looks stale after updates: confirm the running config and browser cache are not serving old assets."
              ]
            },
            {
              "id": "product-issues",
              "title": "Product and model issues",
              "bullets": [
                "Dashboard is empty right after login: confirm initialization state and backend access to the database.",
                "Assistant creation has no model options: save the platform or enterprise model first.",
                "Chat still uses old behavior after config changes: recycle runtime containers and retest with a simple prompt."
              ]
            },
            {
              "id": "ops-checklist",
              "title": "Pre-release operations checklist",
              "bullets": [
                "Home, docs, login, dashboard, and chat all share aligned branding assets.",
                "Login, callback, registration, and logout all work from the same public domain strategy.",
                "The first assistant can be created and used in chat by a non-admin path if your rollout requires it.",
                "Monitoring covers auth failures, API errors, runtime health, and environment drift after deploys."
              ]
            }
          ]
        }
      ]
    }
  ]
};
