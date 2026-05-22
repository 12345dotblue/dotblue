#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/.env"
GENERATED_CONFIG_FILE="${SCRIPT_DIR}/.generated/dotblue/config.yaml"

fail() {
  echo "error: $*" >&2
  exit 1
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

print_section() {
  echo
  echo "== $1 =="
}

require_compose_service() {
  local service="$1"
  if ! docker compose ps --services | grep -qx "${service}"; then
    fail "compose service not found: ${service}"
  fi
}

container_id_for_service() {
  docker compose ps -q "$1"
}

check_equals() {
  local label="$1"
  local expected="$2"
  local actual="$3"
  if [[ "${expected}" == "${actual}" ]]; then
    echo "[ok] ${label}: ${actual}"
  else
    echo "[warn] ${label}: expected '${expected}', got '${actual}'"
  fi
}

check_non_empty() {
  local label="$1"
  local actual="$2"
  if [[ -n "${actual}" ]]; then
    echo "[ok] ${label}: ${actual}"
  else
    echo "[warn] ${label}: empty"
  fi
}

read_generated_engine_value() {
  local key="$1"
  [[ -f "${GENERATED_CONFIG_FILE}" ]] || return 0
  awk -F': ' -v target="${key}" '
    $1 ~ /^[[:space:]]+[A-Za-z]/ {
      gsub(/^[[:space:]]+/, "", $1)
      if ($1 == target) {
        value = $2
        gsub(/^"/, "", value)
        gsub(/"$/, "", value)
        print value
        exit
      }
    }
  ' "${GENERATED_CONFIG_FILE}"
}

[[ -f "${ENV_FILE}" ]] || fail "missing ${ENV_FILE}. Copy .env.example to .env first."

cd "${SCRIPT_DIR}"
load_env_file

require_compose_service backend

backend_cid="$(container_id_for_service backend)"
[[ -n "${backend_cid}" ]] || fail "backend container is not created"

host_socket_gid=""
if [[ -S /var/run/docker.sock ]]; then
  host_socket_gid="$(stat -c '%g' /var/run/docker.sock)"
fi

backend_group_add="$(docker inspect "${backend_cid}" --format '{{join .HostConfig.GroupAdd ","}}')"
backend_user="$(docker inspect "${backend_cid}" --format '{{.Config.User}}')"
backend_status="$(docker inspect "${backend_cid}" --format '{{.State.Status}}')"
backend_networks="$(docker inspect "${backend_cid}" --format '{{range $name, $_ := .NetworkSettings.Networks}}{{$name}} {{end}}')"
backend_socket_meta="$(docker exec "${backend_cid}" sh -lc 'if [ -S /var/run/docker.sock ]; then stat -c "%g %a %U %G" /var/run/docker.sock; fi' 2>/dev/null || true)"
backend_identity="$(docker exec "${backend_cid}" sh -lc 'id' 2>/dev/null || true)"
effective_docker_network="$(read_generated_engine_value "dockerNetwork")"

print_section "Config"
check_non_empty "runtime mode" "${DOTBLUE_ENGINE_RUNTIME_MODE:-}"
check_non_empty "endpoint mode" "${DOTBLUE_ENGINE_ENDPOINT_MODE:-}"
check_non_empty "docker endpoint" "${DOTBLUE_ENGINE_DOCKER_ENDPOINT:-}"
check_non_empty "docker network (effective)" "${effective_docker_network:-${DOTBLUE_ENGINE_DOCKER_NETWORK:-}}"
check_non_empty "host data path" "${DOTBLUE_ENGINE_HOST_DATA_PATH:-}"
check_non_empty "mount data path" "${DOTBLUE_ENGINE_MOUNT_DATA_PATH:-}"

print_section "Docker Socket"
if [[ -n "${host_socket_gid}" ]]; then
  check_equals "host docker.sock gid vs env" "${host_socket_gid}" "${DOTBLUE_ENGINE_DOCKER_SOCKET_GID:-}"
else
  echo "[warn] host docker.sock gid: unavailable from current shell"
fi
check_non_empty "backend group_add" "${backend_group_add}"
check_non_empty "backend user" "${backend_user}"
check_non_empty "backend state" "${backend_status}"
check_non_empty "backend networks" "${backend_networks}"
check_non_empty "backend socket stat" "${backend_socket_meta}"
check_non_empty "backend id" "${backend_identity}"

if [[ -n "${DOTBLUE_ENGINE_DOCKER_SOCKET_GID:-}" && -n "${backend_group_add}" ]]; then
  if [[ ",${backend_group_add}," == *",${DOTBLUE_ENGINE_DOCKER_SOCKET_GID},"* ]]; then
    echo "[ok] backend group_add contains docker socket gid"
  else
    echo "[warn] backend group_add missing docker socket gid ${DOTBLUE_ENGINE_DOCKER_SOCKET_GID}"
  fi
fi

print_section "Compose"
docker compose ps

print_section "Hermes Containers"
hermes_lines="$(docker ps --format '{{.Names}}\t{{.Status}}\t{{.Ports}}' --filter 'name=^hermes_' || true)"
if [[ -n "${hermes_lines}" ]]; then
  printf '%s\n' "${hermes_lines}"
else
  echo "(none running)"
fi
