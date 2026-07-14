#!/bin/bash

CI="${CI:-false}"
CLOUD_BUILD="${CLOUD_BUILD:-false}"
QUIET="${QUIET:-false}"

DOCKER_IMAGE=${DOCKER_IMAGE:-"webplatformtests/wpt.fyi:latest"}
DOCKER_INSTANCE=${DOCKER_INSTANCE:-"wptd-dev-instance"}
WPTD_HOST_WEB_PORT=${WPTD_HOST_WEB_PORT:-"8080"}
WPTD_HOST_GCD_PORT=${WPTD_HOST_GCD_PORT:-"8001"}
WPTD_PATH="$(git rev-parse --show-toplevel)"

function wptd_chown() {
  docker exec -u 0:0 "${DOCKER_INSTANCE}" chown -R $(id -u $USER):$(id -g $USER) $1
}
function wptd_useradd() {
  # Allow the exit code of groupadd to be 4 (GID not unique) or 9 (group name not unique).
  docker exec -u 0:0 "${DOCKER_INSTANCE}" groupadd -g $(id -g $USER) user 2>/dev/null || true
  # Add user to audio & video groups to ensure Chrome can use sandbox. Allow 4 (UID not unique) or 9.
  docker exec -u 0:0 "${DOCKER_INSTANCE}" useradd -u $(id -u $USER) -g $(id -g $USER) -G audio,video user 2>/dev/null || true
  docker exec -u 0:0 "${DOCKER_INSTANCE}" sh -c 'echo "user ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers' 2>/dev/null || true
}

function _get_auth_args() {
  local AUTH_ARGS=()
  local TOKEN=""
  if [[ "${CI}" == "true" || "${CLOUD_BUILD}" == "true" ]]; then
    TOKEN="$(gcloud auth print-access-token 2>/dev/null || true)"
  elif [[ -n "${CLOUDSDK_AUTH_ACCESS_TOKEN:-}" ]]; then
    TOKEN="${CLOUDSDK_AUTH_ACCESS_TOKEN}"
  fi
  if [[ -n "${TOKEN}" ]]; then
    AUTH_ARGS=("-e" "CLOUDSDK_AUTH_ACCESS_TOKEN=${TOKEN}")
  fi
  echo "${AUTH_ARGS[@]}"
}

function wptd_exec() {
  local AUTH_ARGS=($(_get_auth_args))
  docker exec -u $(id -u $USER) "${AUTH_ARGS[@]}" "${DOCKER_INSTANCE}" sh -c "$*"
}
function wptd_exec_it() {
  local AUTH_ARGS=($(_get_auth_args))
  docker exec -it -u $(id -u $USER) "${AUTH_ARGS[@]}" "${DOCKER_INSTANCE}" sh -c "$*"
}
# function wptd_run() {}
function wptd_stop() {
  docker stop "${DOCKER_INSTANCE}"
}
function wptd_rm() {
  docker rm "${DOCKER_INSTANCE}"
}
