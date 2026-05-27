#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/.env"
GENERATED_DIR="${SCRIPT_DIR}/.generated"
CASDOOR_DIR="${GENERATED_DIR}/casdoor"
DOTBLUE_DIR="${GENERATED_DIR}/dotblue"

GENERATED_START="# >>> prepare-config generated >>>"
GENERATED_END="# <<< prepare-config generated <<<"

fail() {
  echo "Error: $*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

load_env_file() {
  while IFS= read -r line || [[ -n "${line}" ]]; do
    line="${line%$'\r'}"
    [[ -z "${line}" ]] && continue
    [[ "${line}" =~ ^[[:space:]]*# ]] && continue
    local key="${line%%=*}"
    local value="${line#*=}"
    export "${key}=${value}"
  done < "${ENV_FILE}"
}

require_env() {
  local key="$1"
  if [[ -z "${!key:-}" ]]; then
    fail "missing required value in .env: ${key}"
  fi
}

read_env_value() {
  local key="$1"
  if [[ -f "${ENV_FILE}" ]]; then
    awk -F= -v target="${key}" '$1 == target { value = substr($0, index($0, "=") + 1) } END { print value }' "${ENV_FILE}"
  fi
}

json_escape_string() {
  printf '%s' "$1" | sed -e 's/\\/\\\\/g' -e 's/"/\\"/g'
}

json_escape_file() {
  awk 'BEGIN { first = 1 }
    {
      gsub(/\r/, "");
      gsub(/\\/, "\\\\");
      gsub(/"/, "\\\"");
      if (!first) {
        printf "\\n";
      }
      printf "%s", $0;
      first = 0;
    }' "$1"
}

indent_file() {
  local prefix="$1"
  local file="$2"
  sed "s/^/${prefix}/" "$file"
}

resolve_host_path() {
  local value="$1"
  if [[ "${value}" = /* ]]; then
    printf '%s' "${value}"
    return
  fi
  printf '%s/%s' "${SCRIPT_DIR}" "${value#./}"
}

resolve_docker_socket_gid() {
  local override="${1:-}"
  local endpoint="${2:-}"
  if [[ -n "${override}" ]]; then
    printf '%s' "${override}"
    return
  fi
  if [[ "${endpoint}" != "unix:///var/run/docker.sock" ]]; then
    printf ''
    return
  fi
  [[ -e /var/run/docker.sock ]] || fail "cannot detect docker socket gid: /var/run/docker.sock is not available. Set DOTBLUE_ENGINE_DOCKER_SOCKET_GID manually in .env."
  stat -c '%g' /var/run/docker.sock
}

update_env_block() {
  local tmp
  tmp="$(mktemp)"
  if [[ -f "${ENV_FILE}" ]]; then
    awk -v start="${GENERATED_START}" -v end="${GENERATED_END}" '
      {
        line = $0
        sub(/\r$/, "", line)
      }
      line == start { skip = 1; next }
      line == end { skip = 0; next }
      !skip {
        sub(/\r$/, "", $0)
        print
      }
    ' "${ENV_FILE}" > "${tmp}"
  fi

  {
    cat "${tmp}"
    if [[ -s "${tmp}" ]] && [[ "$(tail -c1 "${tmp}" | od -An -t x1 | tr -d ' \n')" != "0a" ]]; then
      printf '\n'
    fi
    echo "${GENERATED_START}"
    echo "DOTBLUE_CASDOOR_CLIENT_ID=${DOTBLUE_CASDOOR_CLIENT_ID}"
    echo "DOTBLUE_CASDOOR_CLIENT_SECRET=${DOTBLUE_CASDOOR_CLIENT_SECRET}"
    echo "DOTBLUE_CASDOOR_CERT_NAME=${DOTBLUE_CASDOOR_CERT_NAME}"
    echo "DOTBLUE_ENGINE_DOCKER_SOCKET_GID=${DOTBLUE_ENGINE_DOCKER_SOCKET_GID}"
    echo "${GENERATED_END}"
  } > "${ENV_FILE}.tmp"

  mv "${ENV_FILE}.tmp" "${ENV_FILE}"
  rm -f "${tmp}"
}

generate_cert() {
  local cert_file="$1"
  local key_file="$2"

  if [[ -f "${cert_file}" && -f "${key_file}" ]]; then
    return
  fi

  openssl req \
    -x509 \
    -nodes \
    -newkey rsa:2048 \
    -keyout "${key_file}" \
    -out "${cert_file}" \
    -days 3650 \
    -subj "/CN=${DOTBLUE_CASDOOR_CERT_NAME}" >/dev/null 2>&1
}

require_cmd openssl

[[ -f "${ENV_FILE}" ]] || fail "missing ${ENV_FILE}. Copy .env.example to .env first."

load_env_file

require_env CASDOOR_PUBLIC_URL
require_env CASDOOR_INTERNAL_URL
require_env CASDOOR_ORG_NAME
require_env CASDOOR_APP_NAME
require_env CASDOOR_DB_NAME
require_env CASDOOR_DB_USER
require_env CASDOOR_DB_PASSWORD
require_env DOTBLUE_PUBLIC_URL
require_env DOTBLUE_BACKEND_PUBLIC_URL
require_env DOTBLUE_DB_NAME
require_env DOTBLUE_DB_USER
require_env DOTBLUE_DB_PASSWORD
require_env DOTBLUE_ADMIN_USERNAME
require_env DOTBLUE_ADMIN_DISPLAY_NAME
require_env DOTBLUE_ADMIN_EMAIL
require_env DOTBLUE_ADMIN_PASSWORD
require_env DOTBLUE_BRAND_NAME
require_env DOTBLUE_THEME_PRIMARY

COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-dotblue}"
DOTBLUE_ENGINE_HOST_DATA_PATH="${DOTBLUE_ENGINE_HOST_DATA_PATH:-./.runtime/agents-host}"
DOTBLUE_ENGINE_MOUNT_DATA_PATH="${DOTBLUE_ENGINE_MOUNT_DATA_PATH:-/runtime-data}"
DOTBLUE_ENGINE_RUNTIME_MODE="${DOTBLUE_ENGINE_RUNTIME_MODE:-container}"
DOTBLUE_ENGINE_ENDPOINT_MODE="${DOTBLUE_ENGINE_ENDPOINT_MODE:-docker_dns}"
DOTBLUE_ENGINE_DOCKER_ENDPOINT="${DOTBLUE_ENGINE_DOCKER_ENDPOINT:-unix:///var/run/docker.sock}"
DOTBLUE_ENGINE_DOCKER_NETWORK="${DOTBLUE_ENGINE_DOCKER_NETWORK:-${COMPOSE_PROJECT_NAME}_default}"
DOTBLUE_ENGINE_DOCKER_SOCKET_GID="$(resolve_docker_socket_gid "${DOTBLUE_ENGINE_DOCKER_SOCKET_GID:-}" "${DOTBLUE_ENGINE_DOCKER_ENDPOINT}")"
DOTBLUE_ENGINE_HOST_DATA_PATH_ABS="$(resolve_host_path "${DOTBLUE_ENGINE_HOST_DATA_PATH}")"

DOTBLUE_CASDOOR_CLIENT_ID="${DOTBLUE_CASDOOR_CLIENT_ID:-$(read_env_value DOTBLUE_CASDOOR_CLIENT_ID)}"
DOTBLUE_CASDOOR_CLIENT_SECRET="${DOTBLUE_CASDOOR_CLIENT_SECRET:-$(read_env_value DOTBLUE_CASDOOR_CLIENT_SECRET)}"
DOTBLUE_CASDOOR_CERT_NAME="${DOTBLUE_CASDOOR_CERT_NAME:-$(read_env_value DOTBLUE_CASDOOR_CERT_NAME)}"

DOTBLUE_CASDOOR_CLIENT_ID="${DOTBLUE_CASDOOR_CLIENT_ID:-$(openssl rand -hex 10)}"
DOTBLUE_CASDOOR_CLIENT_SECRET="${DOTBLUE_CASDOOR_CLIENT_SECRET:-$(openssl rand -hex 20)}"
DOTBLUE_CASDOOR_CERT_NAME="${DOTBLUE_CASDOOR_CERT_NAME:-dotblue-jwt-$(openssl rand -hex 4)}"

mkdir -p "${CASDOOR_DIR}/logs" "${CASDOOR_DIR}/certs" "${DOTBLUE_DIR}"
mkdir -p "${DOTBLUE_ENGINE_HOST_DATA_PATH_ABS}"
chmod 0777 "${DOTBLUE_ENGINE_HOST_DATA_PATH_ABS}" 2>/dev/null || true

CERT_FILE="${CASDOOR_DIR}/certs/${DOTBLUE_CASDOOR_CERT_NAME}.pem"
KEY_FILE="${CASDOOR_DIR}/certs/${DOTBLUE_CASDOOR_CERT_NAME}-key.pem"
generate_cert "${CERT_FILE}" "${KEY_FILE}"

CERT_PEM_JSON="$(json_escape_file "${CERT_FILE}")"
KEY_PEM_JSON="$(json_escape_file "${KEY_FILE}")"
CERT_PEM_BLOCK="$(indent_file "    " "${CERT_FILE}")"

update_env_block

cat > "${CASDOOR_DIR}/app.conf" <<EOF
appname = casdoor
httpport = 8000
runmode = prod
copyrequestbody = true
driverName = postgres
dataSourceName = "user=${CASDOOR_DB_USER} password=${CASDOOR_DB_PASSWORD} host=postgres port=5432 sslmode=disable dbname=${CASDOOR_DB_NAME}"
dbName = ${CASDOOR_DB_NAME}
tableNamePrefix =
showSql = false
redisEndpoint =
defaultStorageProvider =
isCloudIntranet = false
authState = "casdoor"
socks5Proxy =
verificationCodeTimeout = 10
initScore = 0
logPostOnly = true
isUsernameLowered = false
origin = ${CASDOOR_PUBLIC_URL}
originFrontend = ${CASDOOR_PUBLIC_URL}
staticBaseUrl = "https://cdn.casbin.org"
isDemoMode = false
batchSize = 100
showGithubCorner = false
forceLanguage = ""
defaultLanguage = "en"
aiAssistantUrl = "https://ai.casbin.com"
defaultApplication = "${CASDOOR_APP_NAME}"
maxItemsForFlatMenu = 7
enableErrorMask = false
enableGzip = true
inactiveTimeoutMinutes =
ldapServerPort = 389
ldapsCertId = ""
ldapsServerPort = 636
radiusServerPort = 1812
radiusDefaultOrganization = "${CASDOOR_ORG_NAME}"
radiusSecret = "secret"
quota = {"organization": -1, "user": -1, "application": -1, "provider": -1}
logConfig = {"adapter":"file", "filename": "logs/casdoor.log", "maxdays":99999, "perm":"0770"}
initDataNewOnly = false
initDataFile = "/init_data.json"
EOF

cat > "${CASDOOR_DIR}/init_data.json" <<EOF
{
  "organizations": [
    {
      "owner": "admin",
      "name": "$(json_escape_string "${CASDOOR_ORG_NAME}")",
      "displayName": "$(json_escape_string "${DOTBLUE_BRAND_NAME}")",
      "websiteUrl": "$(json_escape_string "${DOTBLUE_PUBLIC_URL}")",
      "favicon": "$(json_escape_string "${DOTBLUE_PUBLIC_URL}/brand/dotblue-favicon.svg")",
      "defaultApplication": "$(json_escape_string "${CASDOOR_APP_NAME}")",
      "passwordType": "bcrypt",
      "passwordOptions": ["AtLeast6"],
      "countryCodes": ["US", "CN", "DE", "JP", "SG"],
      "languages": ["en", "zh"],
      "isProfilePublic": true,
      "disableSignin": false
    }
  ],
  "applications": [
    {
      "owner": "admin",
      "name": "$(json_escape_string "${CASDOOR_APP_NAME}")",
      "displayName": "$(json_escape_string "${DOTBLUE_BRAND_NAME}")",
      "logo": "$(json_escape_string "${DOTBLUE_PUBLIC_URL}/brand/dotblue-logo.png")",
      "favicon": "$(json_escape_string "${DOTBLUE_PUBLIC_URL}/brand/dotblue-favicon.svg")",
      "homepageUrl": "$(json_escape_string "${DOTBLUE_PUBLIC_URL}")",
      "organization": "$(json_escape_string "${CASDOOR_ORG_NAME}")",
      "cert": "$(json_escape_string "${DOTBLUE_CASDOOR_CERT_NAME}")",
      "defaultGroup": "admin",
      "enablePassword": true,
      "enableSignUp": true,
      "disableSignin": false,
      "enableSigninSession": true,
      "clientId": "${DOTBLUE_CASDOOR_CLIENT_ID}",
      "clientSecret": "${DOTBLUE_CASDOOR_CLIENT_SECRET}",
      "signinMethods": [
        {
          "name": "Password",
          "displayName": "Password",
          "rule": "All"
        }
      ],
      "signupItems": [
        {
          "name": "Username",
          "visible": true,
          "required": true,
          "prompted": false,
          "rule": "None"
        },
        {
          "name": "Display name",
          "visible": true,
          "required": true,
          "prompted": false,
          "rule": "None"
        },
        {
          "name": "Password",
          "visible": true,
          "required": true,
          "prompted": false,
          "rule": "None"
        },
        {
          "name": "Confirm password",
          "visible": true,
          "required": true,
          "prompted": false,
          "rule": "None"
        },
        {
          "name": "Email",
          "visible": true,
          "required": true,
          "prompted": false,
          "rule": "None"
        }
      ],
      "grantTypes": ["authorization_code", "password", "client_credentials", "refresh_token"],
      "redirectUris": ["$(json_escape_string "${DOTBLUE_PUBLIC_URL}/callback")"],
      "tokenFormat": "JWT",
      "tokenFields": [],
      "expireInHours": 168,
      "themeData": {
        "themeType": "default",
        "colorPrimary": "$(json_escape_string "${DOTBLUE_THEME_PRIMARY}")",
        "borderRadius": 12,
        "isCompact": false,
        "isEnabled": true
      },
      "formCss": ".login-panel{backdrop-filter:blur(18px);background:rgba(255,255,255,0.78);border-radius:24px;box-shadow:0 24px 64px rgba(15,52,96,0.12);}.ant-btn-primary{box-shadow:none;}",
      "formSideHtml": "<div style=\"padding:32px;color:#0f172a;\"><img src=\"$(json_escape_string "${DOTBLUE_PUBLIC_URL}/brand/dotblue-logo.png")\" alt=\"$(json_escape_string "${DOTBLUE_BRAND_NAME}")\" style=\"width:160px;max-width:100%;height:auto;display:block;margin-bottom:20px;\" /><h2 style=\"margin:0 0 12px;\">$(json_escape_string "${DOTBLUE_BRAND_NAME}")</h2><p style=\"margin:0;color:#475569;\">AI workspace platform for modern product, operations, and growth teams.</p></div>",
      "formBackgroundUrl": "$(json_escape_string "${DOTBLUE_PUBLIC_URL}/brand/dotblue-login-bg.png")"
    }
  ],
  "users": [
    {
      "owner": "$(json_escape_string "${CASDOOR_ORG_NAME}")",
      "name": "$(json_escape_string "${DOTBLUE_ADMIN_USERNAME}")",
      "type": "normal-user",
      "password": "$(json_escape_string "${DOTBLUE_ADMIN_PASSWORD}")",
      "displayName": "$(json_escape_string "${DOTBLUE_ADMIN_DISPLAY_NAME}")",
      "avatar": "",
      "email": "$(json_escape_string "${DOTBLUE_ADMIN_EMAIL}")",
      "phone": "",
      "countryCode": "",
      "address": [],
      "addresses": [],
      "affiliation": "",
      "tag": "",
      "score": 2000,
      "ranking": 1,
      "isAdmin": true,
      "isForbidden": false,
      "isDeleted": false,
      "signupApplication": "$(json_escape_string "${CASDOOR_APP_NAME}")",
      "createdIp": "",
      "groups": ["admin"]
    }
  ],
  "certs": [
    {
      "owner": "admin",
      "name": "$(json_escape_string "${DOTBLUE_CASDOOR_CERT_NAME}")",
      "displayName": "$(json_escape_string "${DOTBLUE_BRAND_NAME}") JWT",
      "scope": "JWT",
      "type": "x509",
      "cryptoAlgorithm": "RS256",
      "bitSize": 2048,
      "expireInYears": 10,
      "certificate": "${CERT_PEM_JSON}",
      "privateKey": "${KEY_PEM_JSON}"
    }
  ],
  "groups": [
    {
      "owner": "$(json_escape_string "${CASDOOR_ORG_NAME}")",
      "name": "admin",
      "displayName": "Platform Admins",
      "manager": "$(json_escape_string "${DOTBLUE_ADMIN_USERNAME}")",
      "contactEmail": "$(json_escape_string "${DOTBLUE_ADMIN_EMAIL}")",
      "type": "Virtual",
      "parent_id": "",
      "isTopGroup": true,
      "title": "",
      "key": "",
      "children": [],
      "isEnabled": true
    }
  ]
}
EOF

cat > "${DOTBLUE_DIR}/config.yaml" <<EOF
server:
  address: ":8000"
  openapiPath: "/api.json"
  swaggerPath: "/swagger"

database:
  default:
    link: "pgsql:${DOTBLUE_DB_USER}:${DOTBLUE_DB_PASSWORD}@tcp(postgres:5432)/${DOTBLUE_DB_NAME}"
    debug: true

casdoor:
  endpoint: "${CASDOOR_INTERNAL_URL}"
  clientId: "${DOTBLUE_CASDOOR_CLIENT_ID}"
  clientSecret: "${DOTBLUE_CASDOOR_CLIENT_SECRET}"
  jwtSecret: |
${CERT_PEM_BLOCK}
  organizationName: "${CASDOOR_ORG_NAME}"
  applicationName: "${CASDOOR_APP_NAME}"
  bootstrap:
    endpoint: ""
    clientId: ""
    clientSecret: ""
    jwtSecret: ""

setup:
  initDataPath: ""

logger:
  level: "all"
  stdout: true

debug:
  sse: true

im:
  asyncTurn: true

redis:
  address: "redis:6379"
  password: ""
  db: 0
  keyPrefix: "dot"

session:
  ownerTTL: "30s"
  fenceTTL: "2m"
  gateTTL: "2m"
  stateTTL: "2h"

worker:
  id: "compose-all-in-one"
  embedded: true
  metaTTL: "30s"
  heartbeatInterval: "10s"
  inboxTTL: "2h"
  claimBlock: "2s"

dataplane:
  requestStateRunningTTL: "30m"
  requestStateFinalTTL: "1h"
  streamMaxLen: 5000

engine:
  dataBasePath: "${DOTBLUE_ENGINE_HOST_DATA_PATH_ABS}"
  dataMountPath: "${DOTBLUE_ENGINE_MOUNT_DATA_PATH}"
  containerPort: 8642
  runtimeMode: "${DOTBLUE_ENGINE_RUNTIME_MODE}"
  endpointMode: "${DOTBLUE_ENGINE_ENDPOINT_MODE}"
  dockerEndpoint: "${DOTBLUE_ENGINE_DOCKER_ENDPOINT}"
  dockerNetwork: "${DOTBLUE_ENGINE_DOCKER_NETWORK}"
EOF

if [[ -n "${DOTBLUE_LLM_API_KEY:-}" ]]; then
  PROVIDER_SECTION=$(cat <<EOF
,
  "provider": {
    "type": "$(json_escape_string "${DOTBLUE_LLM_PROVIDER_TYPE}")",
    "apiBase": "$(json_escape_string "${DOTBLUE_LLM_API_BASE}")",
    "apiKey": "$(json_escape_string "${DOTBLUE_LLM_API_KEY}")",
    "model": "$(json_escape_string "${DOTBLUE_LLM_MODEL}")"
  }
EOF
)
else
  PROVIDER_SECTION=""
fi

cat > "${DOTBLUE_DIR}/init_data.json" <<EOF
{
  "version": 1,
  "syncCasdoor": false,
  "organization": {
    "name": "$(json_escape_string "${CASDOOR_ORG_NAME}")",
    "displayName": "$(json_escape_string "${DOTBLUE_BRAND_NAME}")"
  },
  "admin": {
    "username": "$(json_escape_string "${DOTBLUE_ADMIN_USERNAME}")",
    "displayName": "$(json_escape_string "${DOTBLUE_ADMIN_DISPLAY_NAME}")",
    "email": "$(json_escape_string "${DOTBLUE_ADMIN_EMAIL}")",
    "password": "$(json_escape_string "${DOTBLUE_ADMIN_PASSWORD}")"
  },
  "platform": {
    "dataBasePath": "$(json_escape_string "${DOTBLUE_ENGINE_HOST_DATA_PATH_ABS}")",
    "dataMountPath": "$(json_escape_string "${DOTBLUE_ENGINE_MOUNT_DATA_PATH}")",
    "containerPort": 8642,
    "runtimeMode": "$(json_escape_string "${DOTBLUE_ENGINE_RUNTIME_MODE}")",
    "endpointMode": "$(json_escape_string "${DOTBLUE_ENGINE_ENDPOINT_MODE}")",
    "dockerEndpoint": "$(json_escape_string "${DOTBLUE_ENGINE_DOCKER_ENDPOINT}")",
    "dockerNetwork": "$(json_escape_string "${DOTBLUE_ENGINE_DOCKER_NETWORK}")"
  }${PROVIDER_SECTION}
}
EOF

echo "Generated files:"
echo "  - ${CASDOOR_DIR}/app.conf"
echo "  - ${CASDOOR_DIR}/init_data.json"
echo "  - ${DOTBLUE_DIR}/config.yaml"
echo "  - ${DOTBLUE_DIR}/init_data.json"
echo
echo "Updated generated values in ${ENV_FILE}"
