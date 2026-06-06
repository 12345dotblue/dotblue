#!/bin/sh
set -eu

escape_js() {
  printf '%s' "${1:-}" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

casdoor_server_url="${DOTBLUE_WEB_CASDOOR_SERVER_URL:-${VITE_CASDOOR_SERVER_URL:-}}"
casdoor_client_id="${DOTBLUE_WEB_CASDOOR_CLIENT_ID:-${VITE_CASDOOR_CLIENT_ID:-}}"
casdoor_org_name="${DOTBLUE_WEB_CASDOOR_ORG_NAME:-${VITE_CASDOOR_ORG_NAME:-}}"
casdoor_app_name="${DOTBLUE_WEB_CASDOOR_APP_NAME:-${VITE_CASDOOR_APP_NAME:-}}"
casdoor_redirect_url="${DOTBLUE_WEB_CASDOOR_REDIRECT_URL:-${VITE_CASDOOR_REDIRECT_URL:-}}"
backend_url="${DOTBLUE_WEB_BACKEND_URL:-${VITE_BACKEND_URL:-}}"

cat > /usr/share/nginx/html/runtime-config.js <<EOF
window.__DOTBLUE_CONFIG__ = {
  VITE_CASDOOR_SERVER_URL: "$(escape_js "$casdoor_server_url")",
  VITE_CASDOOR_CLIENT_ID: "$(escape_js "$casdoor_client_id")",
  VITE_CASDOOR_ORG_NAME: "$(escape_js "$casdoor_org_name")",
  VITE_CASDOOR_APP_NAME: "$(escape_js "$casdoor_app_name")",
  VITE_CASDOOR_REDIRECT_URL: "$(escape_js "$casdoor_redirect_url")",
  VITE_BACKEND_URL: "$(escape_js "$backend_url")"
};
EOF
