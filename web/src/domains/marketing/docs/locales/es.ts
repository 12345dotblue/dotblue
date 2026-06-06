import type { DocsArticle, LocalizedDocsMeta } from '../schema';

export const localizedMeta: LocalizedDocsMeta = {
  "sections": {
    "getting-started": {
      "title": "Primeros pasos",
      "description": "Entender que es dotblue, como se ve el primer exito y como llegar rapido al primer asistente funcionando."
    },
    "use-dotblue": {
      "title": "Usar dotblue",
      "description": "Pasar del acceso basico a la operacion real del producto: asistentes, modelos, chat y acciones diarias."
    },
    "advanced": {
      "title": "Temas avanzados",
      "description": "Profundizar en estrategia de despliegue, preparacion para produccion, seguridad y fiabilidad operativa."
    }
  },
  "articles": {
    "dotblue-overview": {
      "title": "Vision general de dotblue",
      "summary": "Explica que es dotblue, para que equipos fue creado y como luce un primer despliegue exitoso."
    },
    "quick-start": {
      "title": "Inicio rapido",
      "summary": "La ruta mas corta para levantar la stack local, completar el primer inicio de sesion y validar el primer asistente."
    },
    "login-and-authentication": {
      "title": "Inicio de sesion y autenticacion",
      "summary": "Describe como funciona hoy el acceso local, por que el registro se simplifica por defecto y como ampliarlo de forma segura."
    },
    "assistants-and-workspaces": {
      "title": "Asistentes y espacios de trabajo",
      "summary": "Explica como se relacionan los asistentes, el contexto empresarial y los espacios de trabajo orientados al usuario."
    },
    "providers-and-models": {
      "title": "Proveedores y modelos",
      "summary": "Como pensar la configuracion de modelos en dotblue y que suele causar opciones de modelo faltantes o invalidas."
    },
    "chat-and-operations": {
      "title": "Chat y operaciones diarias",
      "summary": "Que verificar dentro del chat y como usarlo para la validacion inicial y la operacion diaria."
    },
    "deployment-architecture": {
      "title": "Arquitectura de despliegue",
      "summary": "Explica que contiene la stack minima, como deben alinearse las URL publicas y por que la configuracion generada forma parte del producto."
    },
    "production-rollout": {
      "title": "Paso a produccion",
      "summary": "Como pasar de la validacion local a un despliegue controlado en produccion con SEO, dominios, autenticacion y operaciones bien resueltas."
    },
    "troubleshooting-and-ops": {
      "title": "Solucion de problemas y operaciones",
      "summary": "Resume los fallos mas comunes cuando un equipo pasa de la instalacion a la operacion diaria."
    }
  }
};

export const localizedArticles: Record<string, DocsArticle> = {
  "production-rollout": {
    "sectionSlug": "advanced",
    "slug": "production-rollout",
    "title": "Paso a produccion",
    "summary": "Como pasar de la validacion local a un despliegue controlado en produccion con SEO, dominios, autenticacion y disciplina operativa.",
    "seoTitle": "Paso a produccion | dotblue Docs",
    "seoDescription": "Planifica el paso a produccion de dotblue con dominios formales, HTTPS, reverse proxy, gestion de secretos, dependencias duraderas y operaciones fiables.",
    "readingTime": "8 min de lectura",
    "sections": [
      {
        "id": "production-basics",
        "title": "Fundamentos de produccion",
        "paragraphs": [
          "El paso a produccion empieza con dominios publicos estables y una configuracion disciplinada, no solo con containers que casualmente estan corriendo. Los usuarios deben ver una sola marca coherente, una sola ruta de autenticacion coherente y una sola estrategia de URL publica coherente."
        ],
        "code": {
          "language": "text",
          "value": "Deployment checkpoints\n1. Use formal domains for app, API, and auth\n2. Terminate TLS at a trusted reverse proxy\n3. Forward X-Forwarded-* headers correctly\n4. Keep internal container addresses separate from public URLs\n5. Rotate admin and provider secrets before launch"
        }
      },
      {
        "id": "security-and-secrets",
        "title": "Seguridad y manejo de secretos",
        "bullets": [
          "Usa HTTPS para la aplicacion, la API y la autenticacion.",
          "Inyecta provider keys, contrasenas de administrador y otras credentials mediante gestion de secretos en lugar de incrustarlas en imagenes.",
          "Respalda bases de datos y almacenamientos criticos antes de exponer el entorno a usuarios reales.",
          "Trata el branding de Casdoor y la configuracion callback como activos controlados por la release."
        ]
      },
      {
        "id": "seo-and-discoverability",
        "title": "Docs de producto y landing pages amigables con SEO",
        "paragraphs": [
          "Para documentacion publica, importan las URL estables por articulo. Por eso cada gran tema docs debe tener una ruta permanente a nivel de pagina y no solo un anchor dentro de una pagina unica muy larga.",
          "Cada articulo debe exponer su propio title, description, canonical URL y alternate language links para que los motores de busqueda puedan indexarlo como un recurso independiente."
        ]
      }
    ]
  },
  "troubleshooting-and-ops": {
    "sectionSlug": "advanced",
    "slug": "troubleshooting-and-ops",
    "title": "Solucion de problemas y operaciones",
    "summary": "Los patrones de fallo que los equipos realmente encuentran al pasar de la instalacion a la operacion diaria.",
    "seoTitle": "Solucion de problemas y operaciones | dotblue Docs",
    "seoDescription": "Resuelve problemas comunes de dotblue relacionados con redirects auth, dashboards vacios, modelos faltantes, comportamientos runtime obsoletos y deriva de branding.",
    "readingTime": "7 min de lectura",
    "sections": [
      {
        "id": "auth-issues",
        "title": "Problemas de autenticacion y redireccion",
        "bullets": [
          "El login salta al host equivocado: vuelve a revisar las public URLs y regenera la configuracion.",
          "El callback funciona pero la session parece rota: verifica callback path, dominio visible en navegador y supuestos de persistencia del token.",
          "El branding parece antiguo tras actualizar: confirma que la configuracion en ejecucion y la cache del navegador no sigan sirviendo assets viejos."
        ]
      },
      {
        "id": "product-issues",
        "title": "Problemas de producto y modelo",
        "bullets": [
          "El dashboard queda vacio justo despues del login: confirma el estado de inicializacion y el acceso del backend a la base de datos.",
          "La creacion del assistant no muestra modelos: guarda antes el modelo a nivel platform o enterprise.",
          "Chat sigue usando el comportamiento anterior tras cambios de configuracion: recicla runtime containers y vuelve a probar con un prompt simple."
        ]
      },
      {
        "id": "ops-checklist",
        "title": "Checklist operativa antes de publicar",
        "bullets": [
          "home, docs, login, dashboard y chat comparten assets de branding alineados.",
          "login, callback, registration y logout funcionan con la misma estrategia de dominio publico.",
          "Si tu rollout lo requiere, el primer assistant puede crearse y usarse en Chat desde una ruta no administradora.",
          "El monitoreo cubre fallos auth, errores API, salud runtime y deriva del entorno despues de desplegar."
        ]
      }
    ]
  },
  "chat-and-operations": {
    "sectionSlug": "use-dotblue",
    "slug": "chat-and-operations",
    "title": "Chat y operaciones diarias",
    "summary": "Que verificar dentro de Chat, como usarlo para la primera validacion y como deben leer los operadores los fallos.",
    "seoTitle": "Chat y operaciones diarias | dotblue Docs",
    "seoDescription": "Usa dotblue Chat como superficie de validacion operativa, entiende las comprobaciones iniciales y diagnostica problemas comunes de respuesta y runtime.",
    "readingTime": "6 min de lectura",
    "sections": [
      {
        "id": "chat-role",
        "title": "Por que Chat es la prueba operativa clave",
        "paragraphs": [
          "Chat es el lugar donde finalmente confluyen varias partes del producto: autenticacion, configuracion del assistant, configuracion del model, comportamiento runtime y experiencia visible para el usuario.",
          "Por eso, un intercambio exitoso en Chat es una de las comprobaciones iniciales de aceptacion mas fuertes dentro de dotblue."
        ]
      },
      {
        "id": "daily-checks",
        "title": "Comprobaciones diarias para operadores",
        "bullets": [
          "Se puede crear una nueva conversacion sin problemas.",
          "El assistant previsto esta visible y puede seleccionarse.",
          "La primera respuesta llega dentro de una ventana de tiempo razonable.",
          "Los fallos pueden diagnosticarse mediante la ruta de ejecucion visible o los ajustes de platform."
        ]
      },
      {
        "id": "support-playbook",
        "title": "Guia basica de soporte",
        "steps": [
          {
            "title": "Reproducir con un mensaje simple",
            "desc": "Usa un prompt determinista en lugar de una solicitud amplia o ambigua."
          },
          {
            "title": "Revisar la configuracion del modelo",
            "desc": "Asegurate de que el assistant seleccionado este realmente respaldado por un modelo valido y accesible."
          },
          {
            "title": "Revisar la frescura del runtime",
            "desc": "Si la configuracion cambio hace poco, recicla runtime containers antiguos y vuelve a probar."
          },
          {
            "title": "Revisar autenticacion y continuidad de sesion",
            "desc": "Si Chat se comporta de forma extrana tras cambios de login, valida callback, manejo de token y coherencia de redireccion."
          }
        ]
      }
    ]
  },
  "deployment-architecture": {
    "sectionSlug": "advanced",
    "slug": "deployment-architecture",
    "title": "Arquitectura de despliegue",
    "summary": "Que contiene la stack minima, como deben alinearse las URL publicas y que debe pertenecer a la configuracion generada.",
    "seoTitle": "Arquitectura de despliegue | dotblue Docs",
    "seoDescription": "Comprende la arquitectura de despliegue de dotblue, la estrategia de URL publicas, la stack minima de servicios y la alineacion de configuracion generada entre web, backend y Casdoor.",
    "readingTime": "7 min de lectura",
    "sections": [
      {
        "id": "minimal-stack",
        "title": "Stack minima de servicios",
        "paragraphs": [
          "Un despliegue minimo realmente util incluye postgres, redis, casdoor, dotblue y web. Estos servicios cubren persistencia, soporte de sesiones y colas, identidad, API backend y la superficie de producto orientada al navegador."
        ],
        "code": {
          "language": "text",
          "value": "Services\n- postgres\n- redis\n- casdoor\n- dotblue\n- web\n\nBrowser-facing ports\n- Web: 19000\n- Backend: 18080\n- Casdoor: 18000"
        }
      },
      {
        "id": "public-urls",
        "title": "Estrategia de URL publicas",
        "bullets": [
          "Casdoor debe usar una URL publica accesible desde el navegador porque el flujo de acceso del usuario termina alli directamente.",
          "La URL publica del frontend debe coincidir con las URL integradas en la logica de callback y en los recursos de marca.",
          "La URL publica del backend debe reflejar como las llamadas del navegador y los callbacks llegan realmente a la superficie API."
        ]
      },
      {
        "id": "generated-config",
        "title": "La configuracion generada es parte del producto",
        "paragraphs": [
          "No trates los archivos generados como un detalle secundario. En dotblue, la configuracion runtime generada es la forma de mantener alineadas entre servicios las URL publicas, los ajustes de marca y el comportamiento de autenticacion.",
          "Si la marca, las URL de callback o los hostnames se desalinean, regenera la configuracion antes de depurar codigo de aplicacion mas profundo."
        ]
      }
    ]
  },
  "assistants-and-workspaces": {
    "sectionSlug": "use-dotblue",
    "slug": "assistants-and-workspaces",
    "title": "Asistentes y espacios de trabajo",
    "summary": "Explica como encajan los asistentes, el contexto empresarial y los espacios de trabajo visibles para el usuario en el uso diario del producto.",
    "seoTitle": "Asistentes y espacios de trabajo | dotblue Docs",
    "seoDescription": "Aprende como dotblue organiza asistentes, espacios de trabajo, limites de equipo y configuracion inicial para un uso practico del producto.",
    "readingTime": "6 min de lectura",
    "sections": [
      {
        "id": "assistant-model",
        "title": "Como se estructuran los asistentes",
        "paragraphs": [
          "Cada asistente es una superficie de producto con su propia tarea, prompt y comportamiento runtime. Eso significa que tu primera decision de diseno debe ser el scope, no el model. Empieza con una tarea acotada y amplia el alcance solo cuando el workflow sea estable.",
          "La lista de asistentes es el centro operativo para crear, ajustar y verificar estas superficies de producto antes de un uso mas amplio."
        ]
      },
      {
        "id": "workspace-boundaries",
        "title": "Limites entre espacios de trabajo y organizacion",
        "bullets": [
          "Usa organizations y enterprise context cuando el acceso o la configuracion del asistente deban variar por equipo o tenant.",
          "Mantén las decisiones de infraestructura compartida, como el LLM provider por defecto, en los ajustes de platform.",
          "Usa la configuracion propia del assistant para diferencias de negocio que no deban afectar a otros asistentes."
        ]
      },
      {
        "id": "first-assistant-guidance",
        "title": "Como debe ser un buen primer asistente",
        "steps": [
          {
            "title": "Elegir una tarea de negocio clara",
            "desc": "Busqueda de soporte, calificacion comercial o respuestas de conocimiento son un mejor primer paso que un “agente general de empresa”."
          },
          {
            "title": "Escribir un system prompt acotado",
            "desc": "Indica exactamente que debe hacer el asistente, que no debe hacer y que forma de respuesta se espera."
          },
          {
            "title": "Probar en Chat real",
            "desc": "Envia algunas consultas de alta senal y verifica que el comportamiento sea predecible antes de ampliar el uso."
          }
        ]
      }
    ]
  },
  "providers-and-models": {
    "sectionSlug": "use-dotblue",
    "slug": "providers-and-models",
    "title": "Proveedores y modelos",
    "summary": "Explica como pensar la configuracion de modelos en dotblue y que suele causar opciones de modelo faltantes o invalidas.",
    "seoTitle": "Proveedores y modelos | dotblue Docs",
    "seoDescription": "Configura proveedores y modelos LLM en dotblue, entiende los ajustes a nivel plataforma y evita problemas comunes cuando los asistentes no pueden seleccionar o usar un modelo.",
    "readingTime": "7 min de lectura",
    "sections": [
      {
        "id": "platform-models",
        "title": "Configuracion de modelos a nivel plataforma",
        "paragraphs": [
          "dotblue asume que la configuracion de modelos esta disponible antes de que los asistentes puedan usarla. En la practica, eso significa que la capa platform o enterprise necesita una configuracion valida del provider antes de completar la creacion del assistant.",
          "Si los asistentes no ven un modelo, el problema normalmente no esta en la assistant UI. Lo habitual es que falten credenciales del provider, que el API base sea incorrecto o que la definicion del modelo no se haya guardado."
        ]
      },
      {
        "id": "provider-checklist",
        "title": "Checklist del proveedor",
        "bullets": [
          "El provider type coincide con la API real que estas usando.",
          "El API base es accesible desde el backend runtime.",
          "La API key es valida y esta cargada en el entorno runtime real.",
          "El nombre del model coincide con un modelo real y disponible del proveedor.",
          "Los runtime containers anteriores se reciclan o reinician despues de cambios grandes de configuracion."
        ]
      },
      {
        "id": "failure-patterns",
        "title": "Patrones comunes de fallo",
        "bullets": [
          "El modelo aparece en la configuracion pero los asistentes no pueden usarlo: desajuste de scope al guardar o estado runtime obsoleto.",
          "Chat abre pero no llega ninguna respuesta: desajuste entre provider key y nombre del modelo.",
          "Todo funcionaba antes de un cambio: es posible que runtime containers antiguos sigan reteniendo la configuracion previa."
        ]
      }
    ]
  },
  "dotblue-overview": {
    "sectionSlug": "getting-started",
    "slug": "dotblue-overview",
    "title": "Vision general de dotblue",
    "summary": "Explica que es dotblue, para que equipos fue creado y como suele verse un primer despliegue exitoso.",
    "seoTitle": "Vision general de dotblue | dotblue Docs",
    "seoDescription": "Aprende que es dotblue, como lo usan los equipos empresariales y como conecta asistentes productizados, autenticacion, operaciones runtime y despliegue.",
    "readingTime": "6 min de lectura",
    "sections": [
      {
        "id": "what-it-is",
        "title": "Que es dotblue",
        "paragraphs": [
          "dotblue es una capa de entrega para asistentes de IA empresariales. No es solo una interfaz chat ni solo un envoltorio de modelo. Reune acceso con marca, configuracion de modelos a nivel plataforma, gestion de asistentes, control de acceso orientado a equipos y operaciones runtime en una sola experiencia de producto.",
          "El producto esta disenado para equipos que necesitan pasar rapido de una idea a una experiencia de asistente desplegable, manteniendo suficiente control para entornos reales, multiples usuarios y limites organizativos."
        ]
      },
      {
        "id": "core-capabilities",
        "title": "Capacidades principales",
        "bullets": [
          "Autenticacion con marca mediante Casdoor, incluyendo alineacion de callback y logout.",
          "Gestion del ciclo de vida del asistente con prompt, modelo y ajustes de runtime.",
          "Configuracion de modelos a nivel plataforma y empresa para una gobernanza LLM compartida.",
          "Superficies chat con visibilidad de ejecucion y continuidad conversacional.",
          "Activos de despliegue para arranque local, validacion en staging y paso a produccion."
        ]
      },
      {
        "id": "who-uses-it",
        "title": "Para que equipos sirve",
        "paragraphs": [
          "dotblue encaja con equipos de producto que lanzan asistentes internos, equipos de implementacion que entregan entornos para clientes y equipos de plataforma que estandarizan el despliegue de asistentes IA entre distintas organizaciones.",
          "El mejor primer caso de uso suele ser un asistente enfocado con valor de negocio claro, como soporte al cliente, busqueda de conocimiento, copiloto comercial o asistencia a operaciones internas."
        ]
      },
      {
        "id": "first-success",
        "title": "Como se ve un primer exito",
        "steps": [
          {
            "title": "Abrir el sitio del producto",
            "desc": "Confirma que la pagina principal localizada y la documentacion estan disponibles bajo la misma estrategia de URL publica."
          },
          {
            "title": "Entrar por Casdoor",
            "desc": "Verifica el flujo de acceso con marca, el callback y la creacion de sesion hacia el Dashboard."
          },
          {
            "title": "Configurar un modelo",
            "desc": "Guarda al menos un modelo LLM de plataforma o empresa para que los asistentes puedan responder."
          },
          {
            "title": "Crear un asistente",
            "desc": "Define un asistente con alcance acotado, un system prompt claro y una expectativa de salida predecible."
          },
          {
            "title": "Abrir Chat y enviar un mensaje",
            "desc": "Confirma de extremo a extremo el flujo conversacional para el usuario y el comportamiento del runtime."
          }
        ]
      }
    ]
  },
  "quick-start": {
    "sectionSlug": "getting-started",
    "slug": "quick-start",
    "title": "Inicio rapido",
    "summary": "La ruta mas corta para levantar la stack local, completar el primer acceso y validar el primer asistente.",
    "seoTitle": "Inicio rapido | dotblue Docs",
    "seoDescription": "Sigue la guia practica de dotblue para un arranque local con Compose, generacion coherente de configuracion, primer acceso y validacion del primer asistente.",
    "readingTime": "8 min de lectura",
    "sections": [
      {
        "id": "before-you-run",
        "title": "Antes de ejecutar la stack",
        "bullets": [
          "Prepara Docker y Docker Compose en el entorno que realmente usaras para el arranque local.",
          "Decide las URL publicas visibles en el navegador antes de generar la configuracion, especialmente si accedes por IP del host o por direcciones expuestas desde WSL.",
          "Prepara una cuenta administradora util y al menos una API Key LLM valida para que la primera prueba de extremo a extremo llegue a un modelo real."
        ],
        "code": {
          "language": "bash",
          "value": "CASDOOR_PUBLIC_URL=https://auth.example.com\nDOTBLUE_PUBLIC_URL=https://app.example.com\nDOTBLUE_BACKEND_PUBLIC_URL=https://api.example.com\n\nDOTBLUE_ADMIN_USERNAME=admin\nDOTBLUE_ADMIN_EMAIL=admin@example.com\nDOTBLUE_ADMIN_PASSWORD=replace-with-a-strong-password\n\nDOTBLUE_LLM_PROVIDER_TYPE=openai\nDOTBLUE_LLM_API_BASE=https://api.openai.com/v1\nDOTBLUE_LLM_API_KEY=replace-with-provider-key\nDOTBLUE_LLM_MODEL=gpt-4.1-mini"
        }
      },
      {
        "id": "compose-path",
        "title": "Levantar la stack con Compose",
        "paragraphs": [
          "El inicio rapido local se basa en configuracion generada mas un solo comando Compose. Lo importante es que, despues de generar la configuracion, Casdoor, backend y web usen la misma estrategia de URL publica."
        ],
        "code": {
          "language": "bash",
          "value": "cd deploy/compose\ncp .env.example .env\n./prepare-config.sh\ndocker compose up -d --build"
        }
      },
      {
        "id": "windows-path",
        "title": "Ruta para Windows",
        "paragraphs": [
          "Si tu flujo local es principalmente Windows, usa el script PowerShell de preparacion, pero manten las URL publicas generadas totalmente alineadas con la direccion que abriras en el navegador."
        ],
        "code": {
          "language": "powershell",
          "value": "cd deploy\\compose\ncopy .env.example .env\n.\\prepare-config.ps1\ndocker compose up -d --build"
        }
      },
      {
        "id": "first-validation",
        "title": "Validar la primera ejecucion correcta",
        "steps": [
          {
            "title": "Abrir `/es` o tu inicio localizado",
            "desc": "Confirma que la pagina principal del producto carga mediante la direccion para navegador que acabas de configurar."
          },
          {
            "title": "Abrir el flujo de acceso",
            "desc": "Verifica que Casdoor sea accesible y que los recursos de marca carguen con la misma estrategia de dominio publico."
          },
          {
            "title": "Completar el inicio de sesion",
            "desc": "Confirma el regreso al Dashboard sin desajustes de host ni redirecciones incorrectas."
          },
          {
            "title": "Crear el primer asistente",
            "desc": "Si no aparecen opciones de modelo, vuelve primero y guarda la configuracion del modelo en la plataforma."
          }
        ]
      }
    ]
  },
  "login-and-authentication": {
    "sectionSlug": "getting-started",
    "slug": "login-and-authentication",
    "title": "Inicio de sesion y autenticacion",
    "summary": "Explica como funciona hoy el acceso local, por que el registro se simplifica por defecto y como ampliarlo de forma segura.",
    "seoTitle": "Inicio de sesion y autenticacion | dotblue Docs",
    "seoDescription": "Comprende el flujo de acceso de dotblue con Casdoor, la ruta minima de registro para uso local y donde configurar opciones avanzadas de inicio de sesion y verificacion.",
    "readingTime": "7 min de lectura",
    "sections": [
      {
        "id": "default-flow",
        "title": "Flujo local de autenticacion por defecto",
        "paragraphs": [
          "La configuracion local actual mantiene el registro de forma intencional en su forma minima. El alta se centra en Username, Display name, Password y Confirm password para que los equipos puedan levantar la stack sin depender de SMTP, SMS ni verificaciones especificas de cada proveedor.",
          "Esto simplifica la validacion local: una sola stack, una sola ruta de acceso, una sola ruta de callback y una sola estrategia de dominio publico orientada al navegador."
        ]
      },
      {
        "id": "why-simplified",
        "title": "Por que se simplifica el registro local",
        "bullets": [
          "La verificacion por correo requiere entrega SMTP, configuracion del remitente, plantillas y comprobaciones de alcance.",
          "La verificacion por telefono requiere proveedores SMS, plantillas, cuotas y manejo de fallos.",
          "Los equipos que primero validan el flujo del producto suelen necesitar mas un acceso fiable que un despliegue avanzado de identidad desde el primer dia."
        ]
      },
      {
        "id": "advanced-options",
        "title": "Opciones avanzadas de acceso y registro",
        "note": "Trata la ampliacion avanzada de identidad como una tarea de autenticacion de nivel produccion, no como el comportamiento local por defecto.",
        "bullets": [
          "Activa la verificacion por correo solo cuando SMTP este configurado y probado.",
          "Activa la verificacion por telefono solo cuando la entrega SMS forme parte del plan real de despliegue.",
          "El inicio de sesion social, WebAuthn, LDAP o el SSO empresarial deben validarse por etapas dentro de un despliegue controlado."
        ],
        "links": [
          {
            "label": "Casdoor Sign-up Items",
            "url": "https://casdoor.ai/docs/application/signup-items-table",
            "description": "Configura los campos de registro y los requisitos de verificacion."
          },
          {
            "label": "Casdoor Sign-in Methods",
            "url": "https://casdoor.ai/docs/application/signin-methods",
            "description": "Elige Password, verification code, WebAuthn, LDAP y otros metodos de acceso."
          },
          {
            "label": "Casdoor Application Config",
            "url": "https://casdoor.ai/docs/application/config",
            "description": "Revisa las URL de redireccion, los tiempos de reenvio y el comportamiento de autenticacion a nivel de aplicacion."
          },
          {
            "label": "Casdoor Email Provider",
            "url": "https://casdoor.ai/docs/provider/email/overview",
            "description": "Configura SMTP para que la verificacion y el restablecimiento de contrasena puedan enviar mensajes de verdad."
          }
        ]
      }
    ]
  }
};
