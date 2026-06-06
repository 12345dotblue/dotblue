import type { DocsArticle, LocalizedDocsMeta } from '../schema';

export const localizedMeta: LocalizedDocsMeta = {
  "sections": {
    "getting-started": {
      "title": "はじめに",
      "description": "dotblue の役割、最初の成功パターン、ローカル環境の立ち上げと初回ログインまでを最短で理解します。"
    },
    "use-dotblue": {
      "title": "dotblue を使う",
      "description": "ログイン後に、アシスタント作成、モデル設定、チャット運用、日常的な管理までどう進めるかを整理します。"
    },
    "advanced": {
      "title": "高度なトピック",
      "description": "デプロイ戦略、本番移行、セキュリティ境界、長期運用の考え方をまとめます。"
    }
  },
  "articles": {
    "dotblue-overview": {
      "title": "dotblue 概要",
      "summary": "dotblue が何であり、どのチーム向けで、何をもって最初の成功とみなすかを整理します。"
    },
    "quick-start": {
      "title": "クイックスタート",
      "summary": "ローカル環境を立ち上げ、最初のログインと最初のアシスタント確認まで最短で進める手順です。"
    },
    "login-and-authentication": {
      "title": "ログインと認証",
      "summary": "現在のローカル認証の流れ、登録を簡素化している理由、安全な拡張方法を説明します。"
    },
    "assistants-and-workspaces": {
      "title": "アシスタントとワークスペース",
      "summary": "アシスタント、企業コンテキスト、ユーザー向けワークスペースが日常運用でどう結びつくかを説明します。"
    },
    "providers-and-models": {
      "title": "プロバイダーとモデル",
      "summary": "dotblue におけるモデル設定の考え方と、モデル候補が欠落または無効になる典型原因を整理します。"
    },
    "chat-and-operations": {
      "title": "チャットと日常運用",
      "summary": "チャット画面で何を確認すべきか、初回検証と日常運用でどう使うかをまとめます。"
    },
    "deployment-architecture": {
      "title": "デプロイ構成",
      "summary": "最小構成、公開 URL の整合、生成設定ファイルがなぜ重要かを説明します。"
    },
    "production-rollout": {
      "title": "本番展開",
      "summary": "ローカル検証から管理された本番展開へ進む際に必要な SEO、ドメイン、認証、運用の前提を整理します。"
    },
    "troubleshooting-and-ops": {
      "title": "トラブルシュートと運用",
      "summary": "セットアップから日常運用に移る中で、実際によく起きる失敗パターンをまとめます。"
    }
  }
};

export const localizedArticles: Record<string, DocsArticle> = {
  "production-rollout": {
    "sectionSlug": "advanced",
    "slug": "production-rollout",
    "title": "本番展開",
    "summary": "ローカル検証から、適切な SEO、ドメイン、認証、運用規律を備えた制御された本番展開へ進む方法を説明します。",
    "seoTitle": "本番展開 | dotblue Docs",
    "seoDescription": "正式ドメイン、HTTPS、reverse proxy、secret 管理、永続依存、信頼できる運用を前提にした dotblue の本番展開を計画します。",
    "readingTime": "8 分",
    "sections": [
      {
        "id": "production-basics",
        "title": "本番運用の基本前提",
        "paragraphs": [
          "本番展開は、たまたま動いている container を並べることから始まるのではなく、安定した公開ドメインと統制された設定から始まります。ユーザーが見るべきなのは、1 つの一貫したブランド、1 つの一貫した認証経路、1 つの一貫した公開 URL 戦略です。"
        ],
        "code": {
          "language": "text",
          "value": "Deployment checkpoints\n1. Use formal domains for app, API, and auth\n2. Terminate TLS at a trusted reverse proxy\n3. Forward X-Forwarded-* headers correctly\n4. Keep internal container addresses separate from public URLs\n5. Rotate admin and provider secrets before launch"
        }
      },
      {
        "id": "security-and-secrets",
        "title": "セキュリティと秘密情報の取り扱い",
        "bullets": [
          "app、API、auth のすべてで HTTPS を使う。",
          "provider key、管理者パスワード、そのほかの credential は image に焼き込まず secret 管理で注入する。",
          "実ユーザーへ公開する前に、database と重要ストレージの backup を整える。",
          "Casdoor の branding と callback 設定は release 管理対象の資産として扱う。"
        ]
      },
      {
        "id": "seo-and-discoverability",
        "title": "SEO に強い製品ドキュメントとランディング",
        "paragraphs": [
          "公開ドキュメントでは、安定した記事 URL が重要です。主要な docs テーマごとに、長い単一ページ内の anchor ではなく独立した page-level path を持たせるべき理由はここにあります。",
          "各記事は独自の title、description、canonical URL、多言語 alternate link を公開し、検索エンジンが独立した資源として index できるようにするべきです。"
        ]
      }
    ]
  },
  "troubleshooting-and-ops": {
    "sectionSlug": "advanced",
    "slug": "troubleshooting-and-ops",
    "title": "トラブルシューティングと運用",
    "summary": "セットアップから日常運用へ移る際に、多くのチームが実際に直面する失敗パターンを整理します。",
    "seoTitle": "トラブルシューティングと運用 | dotblue Docs",
    "seoDescription": "認証 redirect、空の dashboard、model 不足、古い runtime 挙動、branding のずれなど、dotblue の典型問題を切り分けます。",
    "readingTime": "7 分",
    "sections": [
      {
        "id": "auth-issues",
        "title": "認証と redirect の問題",
        "bullets": [
          "ログイン後に誤った host へ飛ぶ: public URL を再確認し、設定を再生成する。",
          "callback は成功するのに session がおかしい: callback path、ブラウザ向けドメイン、token 維持前提を確認する。",
          "更新後も branding が古く見える: 実行中設定と browser cache が旧 asset を返していないか確認する。"
        ]
      },
      {
        "id": "product-issues",
        "title": "製品と model の問題",
        "bullets": [
          "ログイン直後に dashboard が空白: 初期化状態と backend の database 接続を確認する。",
          "assistant 作成で model が選べない: platform または enterprise model を先に保存する。",
          "設定変更後も Chat が古い挙動のまま: runtime container を再作成し、単純な prompt で再検証する。"
        ]
      },
      {
        "id": "ops-checklist",
        "title": "リリース前の運用チェックリスト",
        "bullets": [
          "home、docs、login、dashboard、chat がすべて整合した branding asset を使っている。",
          "login、callback、registration、logout が同一の公開ドメイン戦略で動作している。",
          "導入要件に含まれるなら、非管理者経路でも最初の assistant を作成し Chat で利用できる。",
          "認証失敗、API error、runtime 健全性、リリース後の環境 drift を監視できる。"
        ]
      }
    ]
  },
  "chat-and-operations": {
    "sectionSlug": "use-dotblue",
    "slug": "chat-and-operations",
    "title": "Chat と日常運用",
    "summary": "Chat 内で何を確認すべきか、初回検証にどう使うか、運用担当が失敗をどう読むべきかを説明します。",
    "seoTitle": "Chat と日常運用 | dotblue Docs",
    "seoDescription": "dotblue Chat を運用検証の画面として使い、初回チェックを理解し、よくある応答や runtime の問題を切り分けます。",
    "readingTime": "6 分",
    "sections": [
      {
        "id": "chat-role",
        "title": "なぜ Chat が運用上の証拠になるのか",
        "paragraphs": [
          "Chat は、認証、assistant 設定、model 構成、runtime 挙動、ユーザー体験が最終的に合流する場所です。",
          "そのため、Chat で 1 回きちんと成功したやり取りができることは、dotblue における最も強い初回受け入れ確認の 1 つです。"
        ]
      },
      {
        "id": "daily-checks",
        "title": "運用担当の日常チェック",
        "bullets": [
          "新しい会話を問題なく作成できる。",
          "意図した assistant が見えており、選択できる。",
          "最初の応答が想定範囲の時間で返る。",
          "失敗した場合でも、実行経路の表示や platform 設定から原因を追える。"
        ]
      },
      {
        "id": "support-playbook",
        "title": "基本的なサポート手順",
        "steps": [
          {
            "title": "単純なメッセージで再現する",
            "desc": "広すぎる依頼や曖昧な依頼ではなく、再現しやすい決定的な prompt を使います。"
          },
          {
            "title": "model 設定を確認する",
            "desc": "選択中の assistant が本当に到達可能で有効な model を使っているか確認します。"
          },
          {
            "title": "runtime の新しさを確認する",
            "desc": "設定を最近変更したなら、古い runtime container を再作成してから再検証します。"
          },
          {
            "title": "認証とセッション継続性を確認する",
            "desc": "ログイン変更後に Chat の挙動がおかしい場合は、callback、token 処理、redirect の整合も確認します。"
          }
        ]
      }
    ]
  },
  "deployment-architecture": {
    "sectionSlug": "advanced",
    "slug": "deployment-architecture",
    "title": "デプロイ構成",
    "summary": "最小スタックの中身、公開 URL をどう揃えるべきか、生成設定に何を含めるべきかを説明します。",
    "seoTitle": "デプロイ構成 | dotblue Docs",
    "seoDescription": "dotblue のデプロイ構成、公開 URL 戦略、最小サービス構成、web・backend・Casdoor 間での生成設定の整合を理解します。",
    "readingTime": "7 分",
    "sections": [
      {
        "id": "minimal-stack",
        "title": "最小サービス構成",
        "paragraphs": [
          "実用的な最小デプロイには postgres、redis、casdoor、dotblue、web が含まれます。これらは永続化、セッションとキュー支援、認証、backend API、ブラウザ向け製品面をそれぞれ担当します。"
        ],
        "code": {
          "language": "text",
          "value": "Services\n- postgres\n- redis\n- casdoor\n- dotblue\n- web\n\nBrowser-facing ports\n- Web: 19000\n- Backend: 18080\n- Casdoor: 18000"
        }
      },
      {
        "id": "public-urls",
        "title": "公開 URL 戦略",
        "bullets": [
          "Casdoor はユーザー向けログインフローが直接到達するため、ブラウザから到達可能な公開 URL を使う必要があります。",
          "frontend の公開 URL は、認証 callback やブランド資産に埋め込まれる URL と一致していなければなりません。",
          "backend の公開 URL は、ブラウザ呼び出しと callback 処理が実際に API 面へ到達する経路を正しく表す必要があります。"
        ]
      },
      {
        "id": "generated-config",
        "title": "生成設定は製品の一部",
        "paragraphs": [
          "生成ファイルを単なる補助物と考えないでください。dotblue では、公開 URL の整合、ブランド設定、認証挙動をサービス間で揃えるために、生成された runtime 設定が重要な役割を持ちます。",
          "ブランド、callback URL、hostname にずれがある場合は、より深いアプリケーションコードを疑う前にまず設定を再生成してください。"
        ]
      }
    ]
  },
  "assistants-and-workspaces": {
    "sectionSlug": "use-dotblue",
    "slug": "assistants-and-workspaces",
    "title": "アシスタントとワークスペース",
    "summary": "アシスタント、企業コンテキスト、ユーザー向けワークスペースが日常利用の中でどう連携するかを説明します。",
    "seoTitle": "アシスタントとワークスペース | dotblue Docs",
    "seoDescription": "dotblue がアシスタント、ワークスペース、チーム境界、初回設定をどのように整理して実運用に結び付けるかを理解します。",
    "readingTime": "6 分",
    "sections": [
      {
        "id": "assistant-model",
        "title": "アシスタントの構成",
        "paragraphs": [
          "各アシスタントは独自の役割、prompt、runtime 挙動を持つ製品面です。最初の設計判断で重要なのは model ではなく scope です。まずは狭い業務に絞り、workflow が安定してから範囲を広げてください。",
          "アシスタント一覧は、これらの製品面を広く公開する前に作成、調整、検証するための運用中心です。"
        ]
      },
      {
        "id": "workspace-boundaries",
        "title": "ワークスペースと組織境界",
        "bullets": [
          "チームや tenant ごとにアシスタントのアクセス権や設定を分ける必要がある場合は、organization や enterprise context を使って分離します。",
          "既定の LLM provider のような共有インフラ判断は platform レベル設定に置きます。",
          "他のアシスタントへ影響させたくない業務差分は assistant 固有設定で扱います。"
        ]
      },
      {
        "id": "first-assistant-guidance",
        "title": "最初のアシスタントを成功させるポイント",
        "steps": [
          {
            "title": "明確な業務を 1 つ選ぶ",
            "desc": "サポート検索、営業の一次判定、知識回答のような用途は、「何でもする会社アシスタント」より良い最初の一歩です。"
          },
          {
            "title": "範囲の狭い system prompt を書く",
            "desc": "何をするのか、何をしないのか、どのような出力を期待するのかを明確に伝えます。"
          },
          {
            "title": "実際の Chat で試す",
            "desc": "利用範囲を広げる前に、信号の強い質問をいくつか送り、挙動が予測可能かを確認します。"
          }
        ]
      }
    ]
  },
  "providers-and-models": {
    "sectionSlug": "use-dotblue",
    "slug": "providers-and-models",
    "title": "プロバイダーとモデル",
    "summary": "dotblue におけるモデル設定の考え方と、モデル候補が出ない・使えないときによくある原因を整理します。",
    "seoTitle": "プロバイダーとモデル | dotblue Docs",
    "seoDescription": "dotblue で LLM provider と model を設定し、platform レベル設定を理解し、アシスタントが model を選択または利用できない典型問題を避けます。",
    "readingTime": "7 分",
    "sections": [
      {
        "id": "platform-models",
        "title": "プラットフォームレベルのモデル設定",
        "paragraphs": [
          "dotblue では、アシスタントが使う前に model 設定が存在している前提です。実際には、assistant 作成体験が成立する前に platform または enterprise レイヤーで有効な provider 設定が必要になります。",
          "アシスタント側で model が見えない場合、原因は assistant UI そのものではないことがほとんどです。多くは provider credential の不足、API base の誤り、または model 定義の未保存です。"
        ]
      },
      {
        "id": "provider-checklist",
        "title": "プロバイダーチェックリスト",
        "bullets": [
          "provider type が実際に使う API と一致している。",
          "API base が backend runtime から到達可能である。",
          "API key が有効で、実際の runtime 環境に読み込まれている。",
          "model 名が provider 側で利用可能な実在 model と一致している。",
          "大きな設定変更後に既存 runtime container を再作成または再起動している。"
        ]
      },
      {
        "id": "failure-patterns",
        "title": "よくある失敗パターン",
        "bullets": [
          "設定上は model が見えるのに assistant で使えない: 保存 scope の不一致または runtime 状態の古さ。",
          "Chat は開くが返答が来ない: provider key または model 名の不一致。",
          "変更前は動いていたのに今は動かない: 古い runtime container が以前の設定を保持している可能性が高い。"
        ]
      }
    ]
  },
  "dotblue-overview": {
    "sectionSlug": "getting-started",
    "slug": "dotblue-overview",
    "title": "dotblue 概要",
    "summary": "dotblue が何であり、どのチーム向けに作られているか、成功した初回導入がどのような状態かを説明します。",
    "seoTitle": "dotblue 概要 | dotblue Docs",
    "seoDescription": "dotblue が何であり、企業チームがどのように使うのか、製品化されたアシスタント、認証、ランタイム運用、デプロイをどう結び付けるのかを理解します。",
    "readingTime": "6 分",
    "sections": [
      {
        "id": "what-it-is",
        "title": "dotblue とは何か",
        "paragraphs": [
          "dotblue は企業向け AI アシスタントの提供面をまとめた製品レイヤーです。単なる chat UI でも単なる model wrapper でもなく、ブランド化されたサインイン、プラットフォームレベルのモデル設定、アシスタント管理、チーム向けアクセス制御、ランタイム運用を 1 つの製品体験に統合します。",
          "この製品は、アイデアから実運用可能なアシスタント体験までを素早く進めたいチーム向けに設計されています。同時に、実環境、複数ユーザー、組織境界を支えるための十分な制御も維持します。"
        ]
      },
      {
        "id": "core-capabilities",
        "title": "中核機能",
        "bullets": [
          "Casdoor を通じたブランド化認証、コールバック、ログアウト整合。",
          "prompt、model、runtime 設定を含むアシスタントのライフサイクル管理。",
          "共有された LLM ガバナンスのためのプラットフォーム級および企業級モデル設定。",
          "実行可視性と会話継続性を備えた chat 画面。",
          "ローカル起動、staging 検証、本番展開を支えるデプロイ資産。"
        ]
      },
      {
        "id": "who-uses-it",
        "title": "この製品が向いているチーム",
        "paragraphs": [
          "dotblue は、社内アシスタントを立ち上げる製品チーム、顧客環境を納品する実装チーム、組織横断で AI アシスタント導入を標準化したいプラットフォームチームに適しています。",
          "最初のユースケースとして最適なのは、明確な業務価値を持つ絞られたアシスタントです。たとえば顧客対応、知識検索、営業支援、社内運用支援などです。"
        ]
      },
      {
        "id": "first-success",
        "title": "最初の成功とは何か",
        "steps": [
          {
            "title": "製品サイトを開く",
            "desc": "ローカライズ済みホームとドキュメントが同じ公開 URL 戦略で利用できることを確認します。"
          },
          {
            "title": "Casdoor 経由でサインインする",
            "desc": "ブランド化されたログイン、コールバック、Dashboard へのセッション確立を確認します。"
          },
          {
            "title": "モデルを設定する",
            "desc": "少なくとも 1 つのプラットフォーム級または企業級 LLM モデルを保存し、アシスタントが応答できるようにします。"
          },
          {
            "title": "アシスタントを作成する",
            "desc": "役割を絞り、明確な system prompt と予測可能な出力期待を持つアシスタントを 1 つ定義します。"
          },
          {
            "title": "Chat を開いてメッセージを送る",
            "desc": "ユーザー向け会話フローとランタイム挙動が end to end で通ることを確認します。"
          }
        ]
      }
    ]
  },
  "quick-start": {
    "sectionSlug": "getting-started",
    "slug": "quick-start",
    "title": "クイックスタート",
    "summary": "ローカル環境を立ち上げ、最初のログインと最初のアシスタント検証までを最短で進める手順です。",
    "seoTitle": "クイックスタート | dotblue Docs",
    "seoDescription": "Compose を使ったローカル起動、設定生成、最初のログイン、最初のアシスタント検証までを practical に進める dotblue のクイックスタートです。",
    "readingTime": "8 分",
    "sections": [
      {
        "id": "before-you-run",
        "title": "起動前の準備",
        "bullets": [
          "実際にローカル起動で使う環境に Docker と Docker Compose を用意します。",
          "設定生成の前に、ブラウザから見える公開 URL を決めます。特に host IP や WSL 公開アドレスでアクセスする場合は先に揃えてください。",
          "利用可能な管理者アカウントと 1 つ以上の有効な LLM API Key を準備し、最初の疎通確認で実際のモデルに到達できるようにします。"
        ],
        "code": {
          "language": "bash",
          "value": "CASDOOR_PUBLIC_URL=https://auth.example.com\nDOTBLUE_PUBLIC_URL=https://app.example.com\nDOTBLUE_BACKEND_PUBLIC_URL=https://api.example.com\n\nDOTBLUE_ADMIN_USERNAME=admin\nDOTBLUE_ADMIN_EMAIL=admin@example.com\nDOTBLUE_ADMIN_PASSWORD=replace-with-a-strong-password\n\nDOTBLUE_LLM_PROVIDER_TYPE=openai\nDOTBLUE_LLM_API_BASE=https://api.openai.com/v1\nDOTBLUE_LLM_API_KEY=replace-with-provider-key\nDOTBLUE_LLM_MODEL=gpt-4.1-mini"
        }
      },
      {
        "id": "compose-path",
        "title": "Compose でスタックを起動する",
        "paragraphs": [
          "ローカルのクイックスタートは、生成済み設定と 1 回の Compose 実行を前提にしています。重要なのは、設定生成後に Casdoor、backend、web が同じ公開 URL 戦略を使っていることです。"
        ],
        "code": {
          "language": "bash",
          "value": "cd deploy/compose\ncp .env.example .env\n./prepare-config.sh\ndocker compose up -d --build"
        }
      },
      {
        "id": "windows-path",
        "title": "Windows 環境",
        "paragraphs": [
          "Windows 中心のローカル運用なら PowerShell の prepare スクリプトを使って構いませんが、生成される公開 URL はブラウザで実際に開くアドレスと必ず一致させてください。"
        ],
        "code": {
          "language": "powershell",
          "value": "cd deploy\\compose\ncopy .env.example .env\n.\\prepare-config.ps1\ndocker compose up -d --build"
        }
      },
      {
        "id": "first-validation",
        "title": "最初の成功を確認する",
        "steps": [
          {
            "title": "`/ja` またはローカライズ済みホームを開く",
            "desc": "今設定したブラウザ向けアドレスで製品ホームが正しく表示されることを確認します。"
          },
          {
            "title": "ログインフローを開く",
            "desc": "Casdoor に到達でき、ブランド資産も同じ公開ドメイン戦略で読み込まれることを確認します。"
          },
          {
            "title": "サインインを完了する",
            "desc": "ホスト不一致や誤ったリダイレクトなしで Dashboard まで戻れることを確認します。"
          },
          {
            "title": "最初のアシスタントを作成する",
            "desc": "モデル候補が出ない場合は、先にプラットフォーム側のモデル設定を保存してください。"
          }
        ]
      }
    ]
  },
  "login-and-authentication": {
    "sectionSlug": "getting-started",
    "slug": "login-and-authentication",
    "title": "ログインと認証",
    "summary": "現在のローカルサインインの仕組み、登録を既定で簡素化している理由、安全に拡張する方法を説明します。",
    "seoTitle": "ログインと認証 | dotblue Docs",
    "seoDescription": "dotblue と Casdoor のログインフロー、ローカル利用向けの最小登録パス、より高度なサインイン検証を設定する場所を理解します。",
    "readingTime": "7 分",
    "sections": [
      {
        "id": "default-flow",
        "title": "既定のローカル認証フロー",
        "paragraphs": [
          "現在のローカルセットアップでは、登録を意図的に最小構成にしています。サインアップでは Username、Display name、Password、Confirm password のみを扱い、SMTP、SMS、各種プロバイダー固有の検証依存なしでチームがスタックを立ち上げられるようにしています。",
          "これにより、ローカル検証は単純になります。ひとつのスタック、ひとつのログイン経路、ひとつのコールバック経路、そしてブラウザ向けの公開ドメイン戦略に集中できます。"
        ]
      },
      {
        "id": "why-simplified",
        "title": "ローカル登録を簡素化している理由",
        "bullets": [
          "メール検証には SMTP 配信、送信者設定、テンプレート、到達性確認が必要です。",
          "電話番号検証には SMS プロバイダー、テンプレート、送信枠、失敗時の処理が必要です。",
          "まず製品フローを検証するチームにとって、初日から高度な ID 展開を行うより、安定したサインインの方が重要なことがほとんどです。"
        ]
      },
      {
        "id": "advanced-options",
        "title": "高度なサインインと登録オプション",
        "note": "高度な認証展開は、ローカル既定値ではなく本番レベルの認証タスクとして扱ってください。",
        "bullets": [
          "メール検証は、SMTP の設定とテストが完了してから有効化してください。",
          "電話番号検証は、SMS 配信が実際の展開計画に含まれる場合にのみ有効化してください。",
          "ソーシャルログイン、WebAuthn、LDAP、企業向け SSO は、段階的な展開の一部として検証するのが適切です。"
        ],
        "links": [
          {
            "label": "Casdoor Sign-up Items",
            "url": "https://casdoor.ai/docs/application/signup-items-table",
            "description": "登録項目と検証要件を設定します。"
          },
          {
            "label": "Casdoor Sign-in Methods",
            "url": "https://casdoor.ai/docs/application/signin-methods",
            "description": "Password、verification code、WebAuthn、LDAP などのログイン方式を選択します。"
          },
          {
            "label": "Casdoor Application Config",
            "url": "https://casdoor.ai/docs/application/config",
            "description": "リダイレクト URL、再送タイムアウト、アプリ単位の認証挙動を確認します。"
          },
          {
            "label": "Casdoor Email Provider",
            "url": "https://casdoor.ai/docs/provider/email/overview",
            "description": "SMTP を設定し、検証メールやパスワードリセットが実際に送信できるようにします。"
          }
        ]
      }
    ]
  }
};
