#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/.env"

required_services=(postgres redis casdoor dotblue web)
max_attempts=20
retry_interval_seconds=2

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

require_running_service() {
  local service="$1"
  if ! docker compose ps --services --status running | grep -qx "${service}"; then
    echo "service not running: ${service}" >&2
    return 1
  fi
}

retry_until_success() {
  local name="$1"
  shift

  local attempt=1
  while (( attempt <= max_attempts )); do
    if "$@"; then
      return 0
    fi
    if (( attempt == max_attempts )); then
      echo "${name} check failed after ${max_attempts} attempts" >&2
      return 1
    fi
    sleep "${retry_interval_seconds}"
    ((attempt++))
  done
}

check_http_once() {
  local url="$1"
  curl -fsS "${url}" >/dev/null
}

check_http() {
  local name="$1"
  local url="$2"
  echo "checking ${name}: ${url}"
  retry_until_success "${name}" check_http_once "${url}"
}

check_json_contains_once() {
  local url="$1"
  local needle="$2"
  local body
  body="$(curl -fsS "${url}")"
  if [[ "${body}" != *"${needle}"* ]]; then
    echo "response missing expected content: ${needle}" >&2
    echo "${body}" >&2
    return 1
  fi
}

check_json_contains() {
  local name="$1"
  local url="$2"
  local needle="$3"
  echo "checking ${name}: ${url}"
  retry_until_success "${name}" check_json_contains_once "${url}" "${needle}"
}

[[ -f "${ENV_FILE}" ]] || {
  echo "missing ${ENV_FILE}. Copy .env.example to .env first." >&2
  exit 1
}

cd "${SCRIPT_DIR}"
load_env_file

for service in "${required_services[@]}"; do
  require_running_service "${service}"
done

check_http "casdoor" "http://127.0.0.1:${CASDOOR_PORT}/"
check_json_contains "dotblue setup status" "http://127.0.0.1:${DOTBLUE_BACKEND_PORT}/api/setup/status" '"initialized":true'
check_http "web" "http://127.0.0.1:${DOTBLUE_WEB_PORT}/"

echo "smoke test passed"
