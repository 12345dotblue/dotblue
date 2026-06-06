import type { DocsArticle, LocalizedDocsMeta } from '../schema';

export const localizedMeta: LocalizedDocsMeta = {
  "sections": {
    "getting-started": {
      "title": "Prise en main",
      "description": "Comprendre ce qu est dotblue, a quoi ressemble le premier succes et comment arriver vite au premier assistant fonctionnel."
    },
    "use-dotblue": {
      "title": "Utiliser dotblue",
      "description": "Passer de la simple connexion a la creation d assistants, au parametrage des modeles et aux operations quotidiennes."
    },
    "advanced": {
      "title": "Sujets avances",
      "description": "Approfondir la strategie de deploiement, la preparation production, la securite et la fiabilite operationnelle."
    }
  },
  "articles": {
    "dotblue-overview": {
      "title": "Vue d ensemble de dotblue",
      "summary": "Explique ce qu est dotblue, pour quelles equipes il est concu et a quoi ressemble un premier succes reel."
    },
    "quick-start": {
      "title": "Demarrage rapide",
      "summary": "Le chemin le plus court pour lancer la stack locale, reussir la premiere connexion et valider le premier assistant."
    },
    "login-and-authentication": {
      "title": "Connexion et authentification",
      "summary": "Decrit le flux de connexion local actuel, pourquoi l inscription est simplifiee par defaut et comment l etendre en securite."
    },
    "assistants-and-workspaces": {
      "title": "Assistants et espaces de travail",
      "summary": "Explique comment assistants, contexte entreprise et espaces utilisateurs s articulent dans l usage quotidien."
    },
    "providers-and-models": {
      "title": "Fournisseurs et modeles",
      "summary": "Comment penser la configuration des modeles dans dotblue et d ou viennent en general les options manquantes ou invalides."
    },
    "chat-and-operations": {
      "title": "Chat et operations quotidiennes",
      "summary": "Ce qu il faut verifier dans le chat et comment l utiliser pour la validation initiale et l exploitation courante."
    },
    "deployment-architecture": {
      "title": "Architecture de deploiement",
      "summary": "Explique la stack minimale, l alignement des URL publiques et le role des fichiers de configuration generes."
    },
    "production-rollout": {
      "title": "Mise en production",
      "summary": "Comment passer de la validation locale a un deploiement production maitrise avec SEO, domaines, auth et operations solides."
    },
    "troubleshooting-and-ops": {
      "title": "Depannage et operations",
      "summary": "Resume les pannes les plus frequentes quand une equipe passe de l installation aux operations quotidiennes."
    }
  }
};

export const localizedArticles: Record<string, DocsArticle> = {
  "production-rollout": {
    "sectionSlug": "advanced",
    "slug": "production-rollout",
    "title": "Passage en production",
    "summary": "Comment passer d une validation locale a un deploiement production maitrise avec SEO, domaines, authentification et discipline operationnelle.",
    "seoTitle": "Passage en production | dotblue Docs",
    "seoDescription": "Planifiez un passage en production dotblue avec domaines formels, HTTPS, reverse proxy, gestion des secrets, dependances durables et operations fiables.",
    "readingTime": "8 min de lecture",
    "sections": [
      {
        "id": "production-basics",
        "title": "Bases de la production",
        "paragraphs": [
          "Le passage en production commence par des domaines publics stables et une configuration disciplinee, pas seulement par des containers qui tournent par hasard. Les utilisateurs doivent voir une marque coherente, un chemin d authentification coherent et une strategie d URL publique coherente."
        ],
        "code": {
          "language": "text",
          "value": "Deployment checkpoints\n1. Use formal domains for app, API, and auth\n2. Terminate TLS at a trusted reverse proxy\n3. Forward X-Forwarded-* headers correctly\n4. Keep internal container addresses separate from public URLs\n5. Rotate admin and provider secrets before launch"
        }
      },
      {
        "id": "security-and-secrets",
        "title": "Securite et gestion des secrets",
        "bullets": [
          "Utilisez HTTPS pour l application, l API et l authentification.",
          "Injectez provider keys, mots de passe administrateur et autres credentials via une gestion des secrets plutot que dans des images bakees.",
          "Sauvegardez bases de donnees et stockages critiques avant d exposer l environnement a de vrais utilisateurs.",
          "Traitez le branding Casdoor et la configuration callback comme des actifs controles par la release."
        ]
      },
      {
        "id": "seo-and-discoverability",
        "title": "Docs produit et pages d accueil favorables au SEO",
        "paragraphs": [
          "Pour une documentation publique, des URL d article stables comptent. C est pourquoi chaque grand sujet docs doit disposer d un chemin de page permanent plutot que d un simple anchor sur une longue page unique.",
          "Chaque article doit exposer son propre title, description, canonical URL et alternate language links afin que les moteurs de recherche puissent l indexer comme une ressource independante."
        ]
      }
    ]
  },
  "troubleshooting-and-ops": {
    "sectionSlug": "advanced",
    "slug": "troubleshooting-and-ops",
    "title": "Depannage et operations",
    "summary": "Les patterns d echec que les equipes rencontrent reellement lorsqu elles passent de l installation a l exploitation quotidienne.",
    "seoTitle": "Depannage et operations | dotblue Docs",
    "seoDescription": "Depannez les problemes courants de dotblue autour des redirects auth, dashboards vides, modeles manquants, runtimes obsoletes et derives de branding.",
    "readingTime": "7 min de lecture",
    "sections": [
      {
        "id": "auth-issues",
        "title": "Problemes d authentification et de redirection",
        "bullets": [
          "Le login saute vers le mauvais host: reverifiez les public URLs puis regenerez la configuration.",
          "Le callback reussit mais la session semble casse: verifiez callback path, domaine navigateur et hypothese de persistance du token.",
          "Le branding semble stale apres mise a jour: confirmez que la configuration en cours et le cache navigateur ne servent pas d anciens assets."
        ]
      },
      {
        "id": "product-issues",
        "title": "Problemes produit et modele",
        "bullets": [
          "Le dashboard est vide juste apres connexion: confirmez l etat d initialisation et l acces backend a la base de donnees.",
          "La creation d assistant ne propose aucun modele: sauvegardez d abord le modele platform ou enterprise.",
          "Chat garde un ancien comportement apres changement de configuration: recreez les runtime containers puis retestez avec un prompt simple."
        ]
      },
      {
        "id": "ops-checklist",
        "title": "Checklist operations avant mise en ligne",
        "bullets": [
          "home, docs, login, dashboard et chat partagent des assets de branding alignes.",
          "login, callback, registration et logout fonctionnent avec la meme strategie de domaine public.",
          "Si votre rollout l exige, le premier assistant peut etre cree et utilise dans Chat via un parcours non admin.",
          "Le monitoring couvre echecs auth, erreurs API, sante runtime et derive de l environnement apres deploiement."
        ]
      }
    ]
  },
  "chat-and-operations": {
    "sectionSlug": "use-dotblue",
    "slug": "chat-and-operations",
    "title": "Chat et operations quotidiennes",
    "summary": "Que verifier dans Chat, comment l utiliser pour la premiere validation et comment les operateurs doivent lire les echecs.",
    "seoTitle": "Chat et operations quotidiennes | dotblue Docs",
    "seoDescription": "Utiliser dotblue Chat comme surface de validation operationnelle, comprendre les controles initiaux et diagnostiquer les problemes frequents de reponse et de runtime.",
    "readingTime": "6 min de lecture",
    "sections": [
      {
        "id": "chat-role",
        "title": "Pourquoi Chat est le vrai point de preuve operationnel",
        "paragraphs": [
          "Chat est l endroit ou plusieurs parties du produit se rejoignent enfin: authentification, configuration des assistants, configuration des modeles, comportement runtime et experience visible par l utilisateur.",
          "C est pourquoi un echange Chat reussi constitue l un des controles d acceptation initiaux les plus solides dans dotblue."
        ]
      },
      {
        "id": "daily-checks",
        "title": "Controles quotidiens pour les operateurs",
        "bullets": [
          "Une nouvelle conversation peut etre creee proprement.",
          "L assistant vise est visible et selectionnable.",
          "La premiere reponse arrive dans une fenetre de temps attendue.",
          "Les echecs restent diagnostiquables via le chemin d execution visible ou les reglages platform."
        ]
      },
      {
        "id": "support-playbook",
        "title": "Parcours de support de base",
        "steps": [
          {
            "title": "Reproduire avec un message simple",
            "desc": "Utilisez un prompt deterministe plutot qu une demande large ou ambigue."
          },
          {
            "title": "Verifier la configuration du modele",
            "desc": "Assurez-vous que l assistant selectionne repose bien sur un modele joignable et valide."
          },
          {
            "title": "Verifier la fraicheur du runtime",
            "desc": "Si la configuration a change recemment, recreez les anciens runtime containers puis testez a nouveau."
          },
          {
            "title": "Verifier l authentification et la continuite de session",
            "desc": "Si Chat se comporte etrangement apres des changements de login, validez callback, gestion du token et coherence des redirections."
          }
        ]
      }
    ]
  },
  "deployment-architecture": {
    "sectionSlug": "advanced",
    "slug": "deployment-architecture",
    "title": "Architecture de deploiement",
    "summary": "Ce que contient la stack minimale, comment aligner les URL publiques et ce qui doit appartenir a la configuration generee.",
    "seoTitle": "Architecture de deploiement | dotblue Docs",
    "seoDescription": "Comprendre l architecture de deploiement dotblue, la strategie d URL publiques, la pile de services minimale et l alignement de configuration generee entre web, backend et Casdoor.",
    "readingTime": "7 min de lecture",
    "sections": [
      {
        "id": "minimal-stack",
        "title": "Pile minimale de services",
        "paragraphs": [
          "Un deploiement minimal vraiment utile comprend postgres, redis, casdoor, dotblue et web. Ces services couvrent respectivement la persistance, le support des sessions et files, l identite, les API backend et la surface produit cote navigateur."
        ],
        "code": {
          "language": "text",
          "value": "Services\n- postgres\n- redis\n- casdoor\n- dotblue\n- web\n\nBrowser-facing ports\n- Web: 19000\n- Backend: 18080\n- Casdoor: 18000"
        }
      },
      {
        "id": "public-urls",
        "title": "Strategie d URL publiques",
        "bullets": [
          "Casdoor doit utiliser une URL publique accessible par le navigateur car le parcours de connexion utilisateur y aboutit directement.",
          "L URL publique du frontend doit correspondre a celle integree dans la logique de callback et les assets de branding.",
          "L URL publique du backend doit refleter la maniere dont les appels navigateur et les callbacks atteignent reellement la surface API."
        ]
      },
      {
        "id": "generated-config",
        "title": "La configuration generee fait partie du produit",
        "paragraphs": [
          "Ne traitez pas les fichiers generes comme un detail secondaire. Dans dotblue, la configuration runtime generee est la maniere de garder coherents l alignement des URL publiques, les reglages de branding et le comportement d authentification entre services.",
          "Si le branding, les URL de callback ou les hostnames divergent, regenerez la configuration avant d aller debugger plus profondement le code applicatif."
        ]
      }
    ]
  },
  "assistants-and-workspaces": {
    "sectionSlug": "use-dotblue",
    "slug": "assistants-and-workspaces",
    "title": "Assistants et espaces de travail",
    "summary": "Explique comment les assistants, le contexte entreprise et les espaces de travail visibles par les utilisateurs s articulent dans l usage quotidien.",
    "seoTitle": "Assistants et espaces de travail | dotblue Docs",
    "seoDescription": "Comprendre comment dotblue organise assistants, espaces de travail, frontieres d equipe et configuration initiale pour un usage produit concret.",
    "readingTime": "6 min de lecture",
    "sections": [
      {
        "id": "assistant-model",
        "title": "Comment les assistants sont structures",
        "paragraphs": [
          "Chaque assistant est une surface produit avec sa propre mission, son prompt et son comportement runtime. Votre premier choix de conception doit donc porter sur le scope, pas sur le model. Commencez par un travail etroit puis n elargissez que lorsque le workflow est stable.",
          "La liste des assistants est le centre operationnel pour creer, ajuster et verifier ces surfaces produit avant un usage plus large."
        ]
      },
      {
        "id": "workspace-boundaries",
        "title": "Frontieres entre espaces de travail et organisations",
        "bullets": [
          "Utilisez organizations et enterprise context lorsque l acces ou la configuration des assistants doit varier selon l equipe ou le tenant.",
          "Gardez les decisions d infrastructure partagee, comme le LLM provider par defaut, au niveau platform.",
          "Placez les differences de comportement metier dans la configuration propre a l assistant afin de ne pas affecter les autres assistants."
        ]
      },
      {
        "id": "first-assistant-guidance",
        "title": "A quoi ressemble un bon premier assistant",
        "steps": [
          {
            "title": "Choisir une tache metier claire",
            "desc": "La recherche support, la qualification commerciale ou la reponse de connaissance constituent un meilleur point de depart qu un “agent d entreprise generaliste”."
          },
          {
            "title": "Ecrire un system prompt resserre",
            "desc": "Indiquez exactement ce que l assistant doit faire, ce qu il ne doit pas faire et la forme de reponse attendue."
          },
          {
            "title": "Tester dans un vrai Chat",
            "desc": "Envoyez quelques requetes a fort signal et verifiez que le comportement reste previsible avant d etendre l usage."
          }
        ]
      }
    ]
  },
  "providers-and-models": {
    "sectionSlug": "use-dotblue",
    "slug": "providers-and-models",
    "title": "Fournisseurs et modeles",
    "summary": "Explique comment penser la configuration des modeles dans dotblue et ce qui provoque le plus souvent des options de modele manquantes ou invalides.",
    "seoTitle": "Fournisseurs et modeles | dotblue Docs",
    "seoDescription": "Configurer les fournisseurs et modeles LLM dans dotblue, comprendre les reglages au niveau plateforme et eviter les problemes frequents lorsque les assistants ne peuvent pas selectionner ou utiliser un modele.",
    "readingTime": "7 min de lecture",
    "sections": [
      {
        "id": "platform-models",
        "title": "Configuration des modeles au niveau plateforme",
        "paragraphs": [
          "dotblue suppose que la configuration des modeles existe avant que les assistants puissent l utiliser. En pratique, cela signifie qu une configuration fournisseur valide doit etre presente au niveau platform ou enterprise avant que la creation d assistant soit complete.",
          "Si les assistants ne voient aucun modele, le probleme vient rarement de l assistant UI. Il s agit le plus souvent d identifiants fournisseur manquants, d un API base incorrect ou d une definition de modele non sauvegardee."
        ]
      },
      {
        "id": "provider-checklist",
        "title": "Checklist fournisseur",
        "bullets": [
          "Le provider type correspond a l API reellement utilisee.",
          "L API base est joignable depuis le backend runtime.",
          "L API key est valide et chargee dans le veritable environnement runtime.",
          "Le nom du model correspond a un modele reellement disponible chez le fournisseur.",
          "Les anciens runtime containers sont recrees ou redemarres apres les gros changements de configuration."
        ]
      },
      {
        "id": "failure-patterns",
        "title": "Schemas d echec frequents",
        "bullets": [
          "Le modele apparait dans la configuration mais les assistants ne peuvent pas l utiliser: decalage de scope de sauvegarde ou etat runtime obsolete.",
          "Chat s ouvre mais aucune reponse ne revient: incoherence entre provider key et nom de modele.",
          "Tout fonctionnait avant une modification: d anciens runtime containers conservent peut etre encore l ancienne configuration."
        ]
      }
    ]
  },
  "dotblue-overview": {
    "sectionSlug": "getting-started",
    "slug": "dotblue-overview",
    "title": "Vue d ensemble de dotblue",
    "summary": "Explique ce qu est dotblue, pour quelles equipes il est concu et a quoi ressemble en general un premier deploiement reussi.",
    "seoTitle": "Vue d ensemble de dotblue | dotblue Docs",
    "seoDescription": "Comprendre ce qu est dotblue, comment les equipes entreprise l utilisent et comment il relie assistants produits, authentification, operations runtime et deploiement.",
    "readingTime": "6 min de lecture",
    "sections": [
      {
        "id": "what-it-is",
        "title": "Ce qu est dotblue",
        "paragraphs": [
          "dotblue est une surface de livraison pour assistants IA en entreprise. Ce n est ni une simple interface chat ni un simple emballage autour d un modele. Il rassemble la connexion brandee, la configuration des modeles au niveau plateforme, la gestion des assistants, le controle d acces par equipe et les operations runtime dans une seule experience produit.",
          "Le produit est concu pour les equipes qui doivent passer rapidement de l idee a une experience assistant deployable, tout en gardant assez de controle pour des environnements reels, plusieurs utilisateurs et des frontieres organisationnelles."
        ]
      },
      {
        "id": "core-capabilities",
        "title": "Capacites principales",
        "bullets": [
          "Authentification brandee via Casdoor avec alignement du callback et du logout.",
          "Gestion du cycle de vie des assistants avec prompt, modele et parametrage runtime.",
          "Configuration des modeles au niveau plateforme et entreprise pour une gouvernance LLM partagee.",
          "Surfaces chat avec visibilite sur l execution et continuite conversationnelle.",
          "Assets de deploiement pour le lancement local, la validation en staging et le passage en production."
        ]
      },
      {
        "id": "who-uses-it",
        "title": "Pour quelles equipes",
        "paragraphs": [
          "dotblue convient aux equipes produit qui lancent des assistants internes, aux equipes d implementation qui livrent des environnements clients et aux equipes plateforme qui standardisent le deploiement d assistants IA a travers plusieurs organisations.",
          "Le meilleur premier cas d usage reste un assistant cible avec une valeur metier claire, comme le support client, la recherche de connaissances, un copilote commercial ou l assistance aux operations internes."
        ]
      },
      {
        "id": "first-success",
        "title": "A quoi ressemble une premiere reussite",
        "steps": [
          {
            "title": "Ouvrir le site produit",
            "desc": "Verifiez que la page d accueil localisee et la documentation sont disponibles avec la meme strategie d URL publique."
          },
          {
            "title": "Se connecter via Casdoor",
            "desc": "Verifiez le flux de connexion brandee, le callback et l etablissement de session vers le Dashboard."
          },
          {
            "title": "Configurer un modele",
            "desc": "Sauvegardez au moins un modele LLM au niveau plateforme ou entreprise afin que les assistants puissent repondre."
          },
          {
            "title": "Creer un assistant",
            "desc": "Definissez un assistant avec un perimetre etroit, un system prompt clair et une attente de sortie previsible."
          },
          {
            "title": "Ouvrir Chat et envoyer un message",
            "desc": "Confirmez que le parcours conversationnel cote utilisateur et le comportement runtime fonctionnent de bout en bout."
          }
        ]
      }
    ]
  },
  "quick-start": {
    "sectionSlug": "getting-started",
    "slug": "quick-start",
    "title": "Demarrage rapide",
    "summary": "Le chemin le plus court pour lancer la stack locale, reussir la premiere connexion et valider le premier assistant.",
    "seoTitle": "Demarrage rapide | dotblue Docs",
    "seoDescription": "Suivez le parcours pratique dotblue pour un lancement local via Compose, une generation de configuration coherente, la premiere connexion et la validation du premier assistant.",
    "readingTime": "8 min de lecture",
    "sections": [
      {
        "id": "before-you-run",
        "title": "Avant de lancer la stack",
        "bullets": [
          "Preparez Docker et Docker Compose dans l environnement utilise pour le lancement local.",
          "Choisissez les URL publiques visibles par le navigateur avant de generer la configuration, surtout si vous accedez a la stack via une IP hote ou une adresse exposee par WSL.",
          "Preparez un compte administrateur utilisable et au moins une cle API LLM valide afin que le premier test de bout en bout atteigne un vrai modele."
        ],
        "code": {
          "language": "bash",
          "value": "CASDOOR_PUBLIC_URL=https://auth.example.com\nDOTBLUE_PUBLIC_URL=https://app.example.com\nDOTBLUE_BACKEND_PUBLIC_URL=https://api.example.com\n\nDOTBLUE_ADMIN_USERNAME=admin\nDOTBLUE_ADMIN_EMAIL=admin@example.com\nDOTBLUE_ADMIN_PASSWORD=replace-with-a-strong-password\n\nDOTBLUE_LLM_PROVIDER_TYPE=openai\nDOTBLUE_LLM_API_BASE=https://api.openai.com/v1\nDOTBLUE_LLM_API_KEY=replace-with-provider-key\nDOTBLUE_LLM_MODEL=gpt-4.1-mini"
        }
      },
      {
        "id": "compose-path",
        "title": "Lancer la stack avec Compose",
        "paragraphs": [
          "Le demarrage local repose sur une configuration generee puis sur une seule commande Compose. Le point important est que Casdoor, backend et web utilisent la meme strategie d URL publique apres la generation."
        ],
        "code": {
          "language": "bash",
          "value": "cd deploy/compose\ncp .env.example .env\n./prepare-config.sh\ndocker compose up -d --build"
        }
      },
      {
        "id": "windows-path",
        "title": "Chemin Windows",
        "paragraphs": [
          "Si votre workflow local est principalement sous Windows, utilisez le script PowerShell de preparation, mais gardez les URL publiques generees strictement coherentes avec l adresse que vous ouvrirez dans le navigateur."
        ],
        "code": {
          "language": "powershell",
          "value": "cd deploy\\compose\ncopy .env.example .env\n.\\prepare-config.ps1\ndocker compose up -d --build"
        }
      },
      {
        "id": "first-validation",
        "title": "Valider la premiere execution reussie",
        "steps": [
          {
            "title": "Ouvrir `/fr` ou votre page d accueil localisee",
            "desc": "Verifiez que la page d accueil du produit se charge bien via l adresse navigateur que vous venez de configurer."
          },
          {
            "title": "Ouvrir le parcours de connexion",
            "desc": "Verifiez que Casdoor est joignable et que les assets de branding se chargent avec la meme strategie de domaine public."
          },
          {
            "title": "Terminer la connexion",
            "desc": "Confirmez le retour vers le Dashboard sans mauvaise redirection ni probleme de correspondance d hote."
          },
          {
            "title": "Creer un premier assistant",
            "desc": "Si les choix de modele n apparaissent pas, revenez d abord sauvegarder le modele au niveau plateforme."
          }
        ]
      }
    ]
  },
  "login-and-authentication": {
    "sectionSlug": "getting-started",
    "slug": "login-and-authentication",
    "title": "Connexion et authentification",
    "summary": "Explique le fonctionnement actuel de la connexion locale, pourquoi l inscription est simplifiee par defaut et comment l etendre en toute securite.",
    "seoTitle": "Connexion et authentification | dotblue Docs",
    "seoDescription": "Comprendre le flux de connexion dotblue avec Casdoor, le parcours d inscription minimal pour un usage local et l emplacement des options avancees de verification.",
    "readingTime": "7 min de lecture",
    "sections": [
      {
        "id": "default-flow",
        "title": "Flux d authentification local par defaut",
        "paragraphs": [
          "La configuration locale actuelle garde volontairement l inscription au strict minimum. Le formulaire se concentre sur Username, Display name, Password et Confirm password afin de permettre aux equipes de lancer la stack sans dependances SMTP, SMS ou verification specifique a un fournisseur.",
          "Cela simplifie la validation locale: une seule stack, un seul chemin de connexion, un seul callback et une seule strategie de domaine public cote navigateur."
        ]
      },
      {
        "id": "why-simplified",
        "title": "Pourquoi l inscription locale est simplifiee",
        "bullets": [
          "La verification email exige une livraison SMTP, une configuration expediteur, des templates et des controles de joignabilite.",
          "La verification par telephone demande un fournisseur SMS, des templates, des quotas et une gestion des echecs.",
          "Les equipes qui valident d abord le parcours produit ont generalement plus besoin d une connexion fiable que d un deploiement identite avance des le premier jour."
        ]
      },
      {
        "id": "advanced-options",
        "title": "Options avancees de connexion et d inscription",
        "note": "Traitez le deploiement identite avance comme un chantier d authentification de niveau production, pas comme un comportement local par defaut.",
        "bullets": [
          "Activez la verification email uniquement lorsque SMTP est configure et teste.",
          "Activez la verification telephone uniquement si la livraison SMS fait partie du vrai plan de deploiement.",
          "Le social login, WebAuthn, LDAP ou le SSO entreprise doivent etre valides par etapes dans un deploiement controle."
        ],
        "links": [
          {
            "label": "Casdoor Sign-up Items",
            "url": "https://casdoor.ai/docs/application/signup-items-table",
            "description": "Configurer les champs d inscription et les exigences de verification."
          },
          {
            "label": "Casdoor Sign-in Methods",
            "url": "https://casdoor.ai/docs/application/signin-methods",
            "description": "Choisir Password, verification code, WebAuthn, LDAP et les autres methodes de connexion."
          },
          {
            "label": "Casdoor Application Config",
            "url": "https://casdoor.ai/docs/application/config",
            "description": "Verifier les URL de redirection, les delais de renvoi et le comportement d authentification de l application."
          },
          {
            "label": "Casdoor Email Provider",
            "url": "https://casdoor.ai/docs/provider/email/overview",
            "description": "Configurer SMTP pour permettre l envoi effectif des messages de verification et de reinitialisation du mot de passe."
          }
        ]
      }
    ]
  }
};
