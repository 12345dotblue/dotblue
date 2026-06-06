import type { DocsArticle, LocalizedDocsMeta } from '../schema';

export const localizedMeta: LocalizedDocsMeta = {
  "sections": {
    "getting-started": {
      "title": "시작하기",
      "description": "dotblue 가 무엇인지, 첫 성공 경로가 어떤 모습인지, 로컬 환경과 첫 로그인까지 어떻게 가장 빠르게 도달하는지 설명합니다."
    },
    "use-dotblue": {
      "title": "dotblue 사용하기",
      "description": "로그인 이후 어시스턴트 생성, 모델 설정, 채팅 운영, 일상 관리까지 어떻게 이어지는지 정리합니다."
    },
    "advanced": {
      "title": "고급 주제",
      "description": "배포 전략, 프로덕션 전환, 보안 경계, 장기 운영 안정성을 다룹니다."
    }
  },
  "articles": {
    "dotblue-overview": {
      "title": "dotblue 개요",
      "summary": "dotblue 가 무엇인지, 어떤 팀을 위한 제품인지, 첫 성공을 무엇으로 판단하는지 설명합니다."
    },
    "quick-start": {
      "title": "빠른 시작",
      "summary": "로컬 스택을 실행하고 첫 로그인과 첫 어시스턴트 검증까지 가장 짧은 경로로 진행하는 방법입니다."
    },
    "login-and-authentication": {
      "title": "로그인과 인증",
      "summary": "현재 로컬 로그인 흐름, 등록을 기본적으로 단순화한 이유, 안전하게 확장하는 방법을 설명합니다."
    },
    "assistants-and-workspaces": {
      "title": "어시스턴트와 워크스페이스",
      "summary": "어시스턴트, 엔터프라이즈 컨텍스트, 사용자 워크스페이스가 일상적인 제품 사용에서 어떻게 연결되는지 설명합니다."
    },
    "providers-and-models": {
      "title": "프로바이더와 모델",
      "summary": "dotblue 에서 모델 설정을 어떻게 이해해야 하는지와, 모델 옵션이 비어 있거나 잘못되는 일반적인 원인을 정리합니다."
    },
    "chat-and-operations": {
      "title": "채팅과 일상 운영",
      "summary": "채팅 화면에서 무엇을 검증해야 하는지, 초기 검증과 운영 중에 어떻게 활용해야 하는지 설명합니다."
    },
    "deployment-architecture": {
      "title": "배포 아키텍처",
      "summary": "최소 스택 구성, 공개 URL 정렬, 생성된 설정이 왜 중요한지 설명합니다."
    },
    "production-rollout": {
      "title": "프로덕션 롤아웃",
      "summary": "로컬 검증에서 통제된 프로덕션 배포로 넘어갈 때 필요한 SEO, 도메인, 인증, 운영 원칙을 정리합니다."
    },
    "troubleshooting-and-ops": {
      "title": "문제 해결과 운영",
      "summary": "설정부터 일상 운영까지 팀이 실제로 자주 맞닥뜨리는 실패 패턴을 정리합니다."
    }
  }
};

export const localizedArticles: Record<string, DocsArticle> = {
  "production-rollout": {
    "sectionSlug": "advanced",
    "slug": "production-rollout",
    "title": "운영 배포",
    "summary": "로컬 검증에서 적절한 SEO, 도메인, 인증, 운영 규율을 갖춘 통제된 운영 배포로 옮겨가는 방법을 설명합니다.",
    "seoTitle": "운영 배포 | dotblue Docs",
    "seoDescription": "정식 도메인, HTTPS, reverse proxy, secret 관리, 영속 의존성, 신뢰 가능한 운영을 기반으로 dotblue 운영 배포를 계획합니다.",
    "readingTime": "8분",
    "sections": [
      {
        "id": "production-basics",
        "title": "운영 배포의 기본 전제",
        "paragraphs": [
          "운영 배포는 우연히 떠 있는 container 에서 시작하지 않습니다. 안정적인 공개 도메인과 절제된 설정에서 시작합니다. 사용자가 보게 되는 것은 하나의 일관된 브랜드, 하나의 일관된 인증 경로, 하나의 일관된 공개 URL 전략이어야 합니다."
        ],
        "code": {
          "language": "text",
          "value": "Deployment checkpoints\n1. Use formal domains for app, API, and auth\n2. Terminate TLS at a trusted reverse proxy\n3. Forward X-Forwarded-* headers correctly\n4. Keep internal container addresses separate from public URLs\n5. Rotate admin and provider secrets before launch"
        }
      },
      {
        "id": "security-and-secrets",
        "title": "보안과 비밀정보 관리",
        "bullets": [
          "app, API, auth 전부에 HTTPS 를 사용한다.",
          "provider key, 관리자 비밀번호, 기타 credential 은 image 에 내장하지 말고 secret 관리로 주입한다.",
          "실사용자에게 환경을 공개하기 전에 database 와 핵심 storage 백업을 준비한다.",
          "Casdoor branding 과 callback 설정은 release 통제 자산으로 다룬다."
        ]
      },
      {
        "id": "seo-and-discoverability",
        "title": "SEO 친화적인 제품 문서와 랜딩",
        "paragraphs": [
          "공개 문서에서는 안정적인 article URL 이 중요합니다. 그래서 주요 docs 주제는 긴 단일 페이지의 anchor 만 두는 대신 독립적인 page-level path 를 가져야 합니다.",
          "각 문서는 자체 title, description, canonical URL, alternate language link 를 제공해 검색 엔진이 독립 자원으로 색인할 수 있어야 합니다."
        ]
      }
    ]
  },
  "troubleshooting-and-ops": {
    "sectionSlug": "advanced",
    "slug": "troubleshooting-and-ops",
    "title": "문제 해결과 운영",
    "summary": "설치에서 일상 운영으로 넘어갈 때 팀이 실제로 자주 맞닥뜨리는 실패 패턴을 정리합니다.",
    "seoTitle": "문제 해결과 운영 | dotblue Docs",
    "seoDescription": "인증 redirect, 빈 dashboard, 누락된 model, 오래된 runtime 동작, branding 불일치 등 dotblue 의 흔한 문제를 진단합니다.",
    "readingTime": "7분",
    "sections": [
      {
        "id": "auth-issues",
        "title": "인증과 redirect 문제",
        "bullets": [
          "로그인이 잘못된 host 로 이동함: public URL 을 다시 확인하고 설정을 다시 생성한다.",
          "callback 은 성공하지만 session 이 이상함: callback path, 브라우저 도메인, token 유지 전제를 점검한다.",
          "업데이트 후에도 branding 이 오래된 것처럼 보임: 실행 중 설정과 browser cache 가 낡은 asset 을 제공하는지 확인한다."
        ]
      },
      {
        "id": "product-issues",
        "title": "제품과 model 문제",
        "bullets": [
          "로그인 직후 dashboard 가 비어 있음: 초기화 상태와 backend 의 database 접근을 확인한다.",
          "assistant 생성 시 model 옵션이 없음: platform 또는 enterprise model 을 먼저 저장한다.",
          "설정 변경 후에도 Chat 이 옛 동작을 사용함: runtime container 를 재생성하고 단순한 prompt 로 다시 검증한다."
        ]
      },
      {
        "id": "ops-checklist",
        "title": "배포 전 운영 체크리스트",
        "bullets": [
          "home, docs, login, dashboard, chat 이 모두 정렬된 branding asset 을 사용한다.",
          "login, callback, registration, logout 이 같은 공개 도메인 전략에서 동작한다.",
          "배포 범위에 포함된다면 비관리자 경로에서도 첫 assistant 생성과 Chat 사용이 가능하다.",
          "인증 실패, API error, runtime 상태, 배포 후 환경 drift 를 모니터링한다."
        ]
      }
    ]
  },
  "chat-and-operations": {
    "sectionSlug": "use-dotblue",
    "slug": "chat-and-operations",
    "title": "Chat 과 일상 운영",
    "summary": "Chat 안에서 무엇을 확인해야 하는지, 첫 검증에 어떻게 활용할지, 운영자가 실패를 어떻게 읽어야 하는지 설명합니다.",
    "seoTitle": "Chat 과 일상 운영 | dotblue Docs",
    "seoDescription": "dotblue Chat 을 운영 검증 화면으로 사용하고, 첫 실행 점검을 이해하며, 흔한 응답 및 runtime 문제를 진단합니다.",
    "readingTime": "6분",
    "sections": [
      {
        "id": "chat-role",
        "title": "왜 Chat 이 운영 증명 지점인가",
        "paragraphs": [
          "Chat 은 인증, assistant 설정, model 구성, runtime 동작, 사용자 경험이 최종적으로 만나는 지점입니다.",
          "그래서 Chat 에서 한 번 제대로 성공한 대화는 dotblue 에서 가장 강력한 첫 수용 확인 중 하나입니다."
        ]
      },
      {
        "id": "daily-checks",
        "title": "운영자의 일상 점검",
        "bullets": [
          "새 대화를 문제없이 만들 수 있다.",
          "의도한 assistant 가 보이고 선택 가능하다.",
          "첫 응답이 기대한 시간 안에 도착한다.",
          "실패가 발생해도 실행 경로 표시나 platform 설정을 통해 원인을 추적할 수 있다."
        ]
      },
      {
        "id": "support-playbook",
        "title": "기본 지원 플레이북",
        "steps": [
          {
            "title": "간단한 메시지로 재현",
            "desc": "모호하거나 지나치게 넓은 요청 대신, 결정적이고 재현 가능한 prompt 를 사용합니다."
          },
          {
            "title": "model 설정 확인",
            "desc": "선택된 assistant 가 실제로 접근 가능한 유효한 model 을 사용 중인지 확인합니다."
          },
          {
            "title": "runtime 최신성 확인",
            "desc": "최근 설정을 바꿨다면, 오래된 runtime container 를 교체한 뒤 다시 테스트합니다."
          },
          {
            "title": "인증과 세션 연속성 확인",
            "desc": "로그인 변경 이후 Chat 이 이상하게 열리면 callback, token 처리, redirect 일관성도 점검합니다."
          }
        ]
      }
    ]
  },
  "deployment-architecture": {
    "sectionSlug": "advanced",
    "slug": "deployment-architecture",
    "title": "배포 아키텍처",
    "summary": "최소 스택에 무엇이 포함되는지, 공개 URL 을 어떻게 맞춰야 하는지, 생성 설정에 무엇이 들어가야 하는지 설명합니다.",
    "seoTitle": "배포 아키텍처 | dotblue Docs",
    "seoDescription": "dotblue 배포 아키텍처, 공개 URL 전략, 최소 서비스 스택, web·backend·Casdoor 사이의 생성 설정 정합성을 이해합니다.",
    "readingTime": "7분",
    "sections": [
      {
        "id": "minimal-stack",
        "title": "최소 서비스 스택",
        "paragraphs": [
          "실용적인 최소 배포에는 postgres, redis, casdoor, dotblue, web 이 포함됩니다. 이 서비스들은 각각 영속성, 세션과 큐 지원, 인증, backend API, 브라우저용 제품 표면을 담당합니다."
        ],
        "code": {
          "language": "text",
          "value": "Services\n- postgres\n- redis\n- casdoor\n- dotblue\n- web\n\nBrowser-facing ports\n- Web: 19000\n- Backend: 18080\n- Casdoor: 18000"
        }
      },
      {
        "id": "public-urls",
        "title": "공개 URL 전략",
        "bullets": [
          "Casdoor 는 사용자 로그인 흐름이 직접 도달하므로 브라우저에서 접근 가능한 공개 URL 을 사용해야 합니다.",
          "frontend 공개 URL 은 인증 callback 로직과 브랜딩 자산에 내장되는 URL 과 일치해야 합니다.",
          "backend 공개 URL 은 브라우저 호출과 callback 흐름이 실제로 API 표면에 도달하는 방식을 정확히 반영해야 합니다."
        ]
      },
      {
        "id": "generated-config",
        "title": "생성 설정은 제품의 일부다",
        "paragraphs": [
          "생성 파일을 부수적인 세부사항으로 생각하지 마세요. dotblue 에서는 생성된 runtime 설정이 공개 URL 정렬, 브랜딩 설정, 인증 동작을 서비스 전반에 걸쳐 일관되게 유지하는 핵심 수단입니다.",
          "브랜딩, callback URL, hostname 이 어긋난다면 더 깊은 애플리케이션 코드를 의심하기 전에 먼저 설정을 다시 생성하세요."
        ]
      }
    ]
  },
  "assistants-and-workspaces": {
    "sectionSlug": "use-dotblue",
    "slug": "assistants-and-workspaces",
    "title": "어시스턴트와 워크스페이스",
    "summary": "어시스턴트, 엔터프라이즈 컨텍스트, 사용자용 워크스페이스가 일상적인 제품 사용에서 어떻게 맞물리는지 설명합니다.",
    "seoTitle": "어시스턴트와 워크스페이스 | dotblue Docs",
    "seoDescription": "dotblue 가 어시스턴트, 워크스페이스, 팀 경계, 첫 설정을 어떻게 구성해 실제 제품 운영으로 연결하는지 알아봅니다.",
    "readingTime": "6분",
    "sections": [
      {
        "id": "assistant-model",
        "title": "어시스턴트 구조 이해",
        "paragraphs": [
          "각 어시스턴트는 고유한 역할, prompt, runtime 동작을 가진 제품 표면입니다. 첫 설계 결정에서 중요한 것은 model 이 아니라 scope 입니다. 먼저 좁은 업무부터 시작하고 workflow 가 안정된 뒤에만 범위를 넓히세요.",
          "어시스턴트 목록은 이러한 제품 표면을 널리 공개하기 전에 생성, 조정, 검증하는 운영 중심 지점입니다."
        ]
      },
      {
        "id": "workspace-boundaries",
        "title": "워크스페이스와 조직 경계",
        "bullets": [
          "팀이나 tenant 별로 어시스턴트 접근 또는 설정이 달라야 한다면 organization 과 enterprise context 로 분리합니다.",
          "기본 LLM provider 같은 공유 인프라 판단은 platform 수준 설정에 둡니다.",
          "다른 어시스턴트에 영향을 주면 안 되는 비즈니스 차이는 assistant 전용 설정으로 관리합니다."
        ]
      },
      {
        "id": "first-assistant-guidance",
        "title": "첫 어시스턴트를 잘 만드는 방법",
        "steps": [
          {
            "title": "명확한 비즈니스 작업 하나를 고르기",
            "desc": "지원 조회, 영업 자격 판별, 지식 응답 같은 작업은 “만능 회사 어시스턴트”보다 더 좋은 첫 단계입니다."
          },
          {
            "title": "범위가 좁은 system prompt 작성",
            "desc": "무엇을 해야 하는지, 무엇을 하지 말아야 하는지, 어떤 형태의 답변을 기대하는지 명확히 적습니다."
          },
          {
            "title": "실제 Chat 에서 검증",
            "desc": "사용 범위를 넓히기 전에 신호가 강한 질문 몇 개를 보내고 동작이 예측 가능한지 확인합니다."
          }
        ]
      }
    ]
  },
  "providers-and-models": {
    "sectionSlug": "use-dotblue",
    "slug": "providers-and-models",
    "title": "프로바이더와 모델",
    "summary": "dotblue 에서 모델 구성을 어떻게 바라봐야 하는지와 모델 선택지가 없거나 잘못될 때의 전형적인 원인을 설명합니다.",
    "seoTitle": "프로바이더와 모델 | dotblue Docs",
    "seoDescription": "dotblue 에서 LLM provider 와 model 을 구성하고, platform 수준 설정을 이해하며, 어시스턴트가 model 을 선택하거나 사용할 수 없을 때의 대표 문제를 피합니다.",
    "readingTime": "7분",
    "sections": [
      {
        "id": "platform-models",
        "title": "플랫폼 수준 모델 구성",
        "paragraphs": [
          "dotblue 는 어시스턴트가 사용하기 전에 model 구성이 준비되어 있다고 가정합니다. 실제로는 assistant 생성 경험이 완성되기 전에 platform 또는 enterprise 계층에 유효한 provider 설정이 있어야 합니다.",
          "어시스턴트에서 model 이 보이지 않는다면 문제는 assistant UI 자체가 아닌 경우가 대부분입니다. 보통은 provider credential 누락, 잘못된 API base, 혹은 저장되지 않은 model 정의가 원인입니다."
        ]
      },
      {
        "id": "provider-checklist",
        "title": "프로바이더 점검 목록",
        "bullets": [
          "provider type 이 실제 사용하는 API 와 일치한다.",
          "API base 가 backend runtime 에서 접근 가능하다.",
          "API key 가 유효하고 실제 runtime 환경에 주입되어 있다.",
          "model 이름이 provider 에서 실제로 사용 가능한 모델과 일치한다.",
          "큰 설정 변경 후 기존 runtime container 를 재시작하거나 교체했다."
        ]
      },
      {
        "id": "failure-patterns",
        "title": "자주 보이는 실패 패턴",
        "bullets": [
          "설정에는 model 이 보이지만 assistant 에서 사용할 수 없음: 저장 scope 불일치 또는 오래된 runtime 상태.",
          "Chat 은 열리지만 답변이 돌아오지 않음: provider key 또는 model 이름 불일치.",
          "변경 전에는 됐는데 지금은 안 됨: 오래된 runtime container 가 이전 설정을 계속 쓰고 있을 가능성이 큼."
        ]
      }
    ]
  },
  "dotblue-overview": {
    "sectionSlug": "getting-started",
    "slug": "dotblue-overview",
    "title": "dotblue 개요",
    "summary": "dotblue 가 무엇인지, 어떤 팀을 위해 만들어졌는지, 성공적인 첫 도입이 어떤 모습인지 설명합니다.",
    "seoTitle": "dotblue 개요 | dotblue Docs",
    "seoDescription": "dotblue 가 무엇인지, 엔터프라이즈 팀이 어떻게 사용하는지, 제품형 어시스턴트와 인증, 런타임 운영, 배포를 어떻게 연결하는지 알아봅니다.",
    "readingTime": "6분",
    "sections": [
      {
        "id": "what-it-is",
        "title": "dotblue 는 무엇인가",
        "paragraphs": [
          "dotblue 는 엔터프라이즈 AI 어시스턴트 제공을 위한 제품 계층입니다. 단순한 chat UI 도 아니고 단순한 model wrapper 도 아닙니다. 브랜딩된 로그인, 플랫폼 수준 모델 설정, 어시스턴트 관리, 팀 단위 접근 제어, 런타임 운영을 하나의 제품 경험으로 묶습니다.",
          "이 제품은 아이디어에서 실제 배포 가능한 어시스턴트 경험까지 빠르게 이동해야 하는 팀을 위해 설계되었습니다. 동시에 실제 환경, 다중 사용자, 조직 경계를 지원할 만큼의 제어도 유지합니다."
        ]
      },
      {
        "id": "core-capabilities",
        "title": "핵심 기능",
        "bullets": [
          "Casdoor 기반 브랜딩 인증과 콜백, 로그아웃 정합성.",
          "prompt, model, runtime 설정을 포함한 어시스턴트 라이프사이클 관리.",
          "공유 LLM 거버넌스를 위한 플랫폼 및 엔터프라이즈 모델 설정.",
          "실행 가시성과 대화 연속성을 제공하는 chat 화면.",
          "로컬 기동, staging 검증, 운영 배포를 지원하는 배포 자산."
        ]
      },
      {
        "id": "who-uses-it",
        "title": "이 제품이 적합한 팀",
        "paragraphs": [
          "dotblue 는 내부 어시스턴트를 출시하는 제품 팀, 고객 환경을 전달하는 구현 팀, 여러 조직에 걸친 AI 어시스턴트 도입을 표준화하려는 플랫폼 팀에 적합합니다.",
          "가장 좋은 첫 사용 사례는 경계가 명확하고 사업 가치가 분명한 집중형 어시스턴트입니다. 예를 들면 고객 지원, 지식 검색, 세일즈 코파일럿, 내부 운영 지원 등이 있습니다."
        ]
      },
      {
        "id": "first-success",
        "title": "첫 성공의 기준",
        "steps": [
          {
            "title": "제품 사이트 열기",
            "desc": "현지화된 홈과 문서가 같은 공개 URL 전략으로 제공되는지 확인합니다."
          },
          {
            "title": "Casdoor 로 로그인",
            "desc": "브랜딩된 로그인 흐름, 콜백, Dashboard 세션 수립을 확인합니다."
          },
          {
            "title": "모델 구성",
            "desc": "어시스턴트가 응답할 수 있도록 플랫폼 또는 엔터프라이즈 LLM 모델을 하나 이상 저장합니다."
          },
          {
            "title": "어시스턴트 생성",
            "desc": "범위가 좁고 명확한 system prompt 와 예측 가능한 출력 기대를 가진 어시스턴트를 하나 정의합니다."
          },
          {
            "title": "Chat 에서 메시지 보내기",
            "desc": "사용자 대화 흐름과 런타임 동작이 end to end 로 이어지는지 확인합니다."
          }
        ]
      }
    ]
  },
  "quick-start": {
    "sectionSlug": "getting-started",
    "slug": "quick-start",
    "title": "빠른 시작",
    "summary": "로컬 환경을 띄우고 첫 로그인과 첫 어시스턴트 검증까지 가장 짧게 가는 실행 경로입니다.",
    "seoTitle": "빠른 시작 | dotblue Docs",
    "seoDescription": "Compose 기반 로컬 기동, 설정 생성, 첫 로그인, 첫 어시스턴트 검증까지 이어지는 dotblue 빠른 시작 가이드입니다.",
    "readingTime": "8분",
    "sections": [
      {
        "id": "before-you-run",
        "title": "스택 실행 전 준비",
        "bullets": [
          "실제로 로컬 기동에 사용할 환경에 Docker 와 Docker Compose 를 준비합니다.",
          "설정을 생성하기 전에 브라우저에서 접근할 공개 URL 을 결정하세요. 특히 host IP 나 WSL 노출 주소로 접근한다면 먼저 맞춰야 합니다.",
          "사용 가능한 관리자 계정과 유효한 LLM API Key 하나 이상을 준비해 첫 종단간 검증이 실제 모델까지 도달하게 합니다."
        ],
        "code": {
          "language": "bash",
          "value": "CASDOOR_PUBLIC_URL=https://auth.example.com\nDOTBLUE_PUBLIC_URL=https://app.example.com\nDOTBLUE_BACKEND_PUBLIC_URL=https://api.example.com\n\nDOTBLUE_ADMIN_USERNAME=admin\nDOTBLUE_ADMIN_EMAIL=admin@example.com\nDOTBLUE_ADMIN_PASSWORD=replace-with-a-strong-password\n\nDOTBLUE_LLM_PROVIDER_TYPE=openai\nDOTBLUE_LLM_API_BASE=https://api.openai.com/v1\nDOTBLUE_LLM_API_KEY=replace-with-provider-key\nDOTBLUE_LLM_MODEL=gpt-4.1-mini"
        }
      },
      {
        "id": "compose-path",
        "title": "Compose 로 스택 기동하기",
        "paragraphs": [
          "로컬 빠른 시작은 생성된 설정과 한 번의 Compose 실행을 전제로 합니다. 핵심은 설정 생성 이후 Casdoor, backend, web 이 같은 공개 URL 전략을 사용하도록 맞추는 것입니다."
        ],
        "code": {
          "language": "bash",
          "value": "cd deploy/compose\ncp .env.example .env\n./prepare-config.sh\ndocker compose up -d --build"
        }
      },
      {
        "id": "windows-path",
        "title": "Windows 경로",
        "paragraphs": [
          "Windows 중심 워크플로라면 PowerShell prepare 스크립트를 사용해도 되지만, 생성되는 공개 URL 은 브라우저에서 실제로 열 주소와 반드시 일치해야 합니다."
        ],
        "code": {
          "language": "powershell",
          "value": "cd deploy\\compose\ncopy .env.example .env\n.\\prepare-config.ps1\ndocker compose up -d --build"
        }
      },
      {
        "id": "first-validation",
        "title": "첫 성공 경로 검증",
        "steps": [
          {
            "title": "`/ko` 또는 현지화된 홈 열기",
            "desc": "방금 설정한 브라우저용 주소로 제품 홈이 정상적으로 열리는지 확인합니다."
          },
          {
            "title": "로그인 흐름 열기",
            "desc": "Casdoor 에 접근 가능하고 브랜딩 자산도 같은 공개 도메인 전략으로 로드되는지 확인합니다."
          },
          {
            "title": "로그인 완료",
            "desc": "호스트 불일치나 잘못된 리디렉션 없이 Dashboard 로 돌아오는지 확인합니다."
          },
          {
            "title": "첫 어시스턴트 만들기",
            "desc": "모델 선택지가 보이지 않으면 먼저 플랫폼 모델 설정을 저장한 뒤 다시 시도합니다."
          }
        ]
      }
    ]
  },
  "login-and-authentication": {
    "sectionSlug": "getting-started",
    "slug": "login-and-authentication",
    "title": "로그인과 인증",
    "summary": "현재 로컬 로그인 방식, 기본 등록 절차를 단순화한 이유, 이를 안전하게 확장하는 방법을 설명합니다.",
    "seoTitle": "로그인과 인증 | dotblue Docs",
    "seoDescription": "dotblue 와 Casdoor 로그인 흐름, 로컬 사용을 위한 최소 등록 경로, 고급 로그인 및 검증 설정 위치를 이해합니다.",
    "readingTime": "7분",
    "sections": [
      {
        "id": "default-flow",
        "title": "기본 로컬 인증 흐름",
        "paragraphs": [
          "현재 로컬 환경은 등록 절차를 의도적으로 최소화합니다. 회원가입은 Username, Display name, Password, Confirm password 에 집중해 SMTP, SMS, 공급자별 검증 의존성 없이도 팀이 스택을 올릴 수 있게 합니다.",
          "이렇게 하면 로컬 검증이 단순해집니다. 하나의 스택, 하나의 로그인 경로, 하나의 콜백 경로, 그리고 브라우저에서 접근하는 공개 도메인 전략만 맞추면 됩니다."
        ]
      },
      {
        "id": "why-simplified",
        "title": "로컬 등록을 단순화한 이유",
        "bullets": [
          "이메일 검증에는 SMTP 발송, 발신자 설정, 템플릿, 수신 가능 여부 확인이 필요합니다.",
          "전화번호 검증에는 SMS 공급자, 템플릿, 할당량, 실패 처리 흐름이 필요합니다.",
          "제품 흐름을 먼저 검증하는 팀에게는 첫날부터 고급 ID 체계를 도입하는 것보다 안정적인 로그인 경로가 더 중요합니다."
        ]
      },
      {
        "id": "advanced-options",
        "title": "고급 로그인 및 가입 옵션",
        "note": "고급 인증 도입은 로컬 기본값이 아니라 운영 수준의 인증 과제로 다뤄야 합니다.",
        "bullets": [
          "이메일 검증은 SMTP 설정과 테스트가 끝난 뒤에만 활성화하세요.",
          "전화번호 검증은 실제 배포 계획에 SMS 전달이 포함된 경우에만 활성화하세요.",
          "소셜 로그인, WebAuthn, LDAP, 엔터프라이즈 SSO 는 단계적 배포 과정에서 검증하는 것이 좋습니다."
        ],
        "links": [
          {
            "label": "Casdoor Sign-up Items",
            "url": "https://casdoor.ai/docs/application/signup-items-table",
            "description": "가입 필드와 검증 요구사항을 설정합니다."
          },
          {
            "label": "Casdoor Sign-in Methods",
            "url": "https://casdoor.ai/docs/application/signin-methods",
            "description": "Password, verification code, WebAuthn, LDAP 등 로그인 방식을 선택합니다."
          },
          {
            "label": "Casdoor Application Config",
            "url": "https://casdoor.ai/docs/application/config",
            "description": "리디렉션 URL, 재전송 타임아웃, 애플리케이션 단위 인증 동작을 검토합니다."
          },
          {
            "label": "Casdoor Email Provider",
            "url": "https://casdoor.ai/docs/provider/email/overview",
            "description": "SMTP 를 설정해 인증 메일과 비밀번호 재설정 메일을 실제로 발송할 수 있게 합니다."
          }
        ]
      }
    ]
  }
};
