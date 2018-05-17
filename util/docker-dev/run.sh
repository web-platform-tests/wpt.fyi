#!/bin/bash

# Start Docker-based development server as `wptd-dev-instance` in the
# foreground.

DOCKER_DIR=$(dirname $0)
source "${DOCKER_DIR}/../commands.sh"
source "${DOCKER_DIR}/../logging.sh"
source "${DOCKER_DIR}/../path.sh"
WPTD_PATH=${WPTD_PATH:-$(absdir ${DOCKER_DIR}/../..)}

WPTD_HOST_WEB_PORT=${WPTD_HOST_WEB_PORT:-"8080"}
WPTD_HOST_ADMIN_WEB_PORT=${WPTD_HOST_ADMIN_WEB_PORT:-"8000"}
WPTD_HOST_API_WEB_PORT=${WPTD_HOST_API_WEB_PORT:-"9999"}

PR=""

function usage() {
  USAGE="USAGE: $(basename ${0}) [-q]
    -q  quiet mode: Assume default for all prompts"
  >&2 echo "${USAGE}"
}

QUIET="false"
while getopts ':hq' FLAG; do
  case "${FLAG}" in
    q)
      QUIET="true" ;;
    h|*) usage && exit 0 ;;
  esac
done

# Create a docker instance:
#
# --rm                                      Auto-remove when stopped
# -it                                       Interactive mode (Ctrl+c will halt
#                                           instance)
# -v "${WPTD_PATH}":/wpt.fyi                Mount the repository
# -u $(id -u $USER):$(id -g $USER)          Run as current user and group
# -p "${WPTD_HOST_WEB_PORT}:8080"           Expose web server port
# --name wptd-dev-instance                  Name the instance
# wptd-dev                                  Identify image to use
# /wpt.fyi/util/docker/inner/watch.sh       Identify code to execute

info "Creating docker instance for dev server. Instance name: wptd-dev-instance"
docker inspect wptd-dev-instance > /dev/null 2>&1
INSPECT_STATUS="${?}"
docker inspect wptd-dev-instance | grep '"Running": true' | read
RUNNING_STATUS="${?}"

function stop() {
  warn "run.sh: Recieved interrupt. Exiting..."
  info "Stopping wptd-dev-instance..."
  wptd_stop
  info "wptd-dev-instance stopped."
  if [[ "${QUIET}" != "true" ]]; then
    read -p "Docker instance wpt-dev-instance: (p)reserve or (r)emove (P/r): " PR
  fi
  if [[ "${PR}" == "r" ]] || [[ "${PR}" == "R" ]]; then
    info "Removing wptd-dev-instance..."
    wptd_rm
    info "wptd-dev-instance removed."
  fi
}

function quit() {
  stop
  exit 0
}

if [ "${INSPECT_STATUS}" == "0" ]; then
  info "Found existing docker instance wptd-dev-instance."
  if [[ "${QUIET}" != "true" ]]; then
    read -p "Docker instance wpt-dev-instance: (p)reserve or (r)emove (P/r): " PR
  fi
  if [[ "${PR}" == "r" ]] || [[ "${PR}" == "R" ]]; then
    stop
  fi
fi

set -e



if [[ "${INSPECT_STATUS}" != 0 ]] || [[ "${PR}" == "r" ]] || [[ "${PR}" == "R" ]]; then
  info "Starting docker instance wptd-dev-instance..."
  docker run -t -d --entrypoint /bin/bash \
      -v "${WPTD_PATH}":/home/user/wpt.fyi \
      -u $(id -u $USER):$(id -g $USER) \
      -p "${WPTD_HOST_WEB_PORT}:8080" \
      -p "${WPTD_HOST_ADMIN_WEB_PORT}:8000" \
      -p "${WPTD_HOST_API_WEB_PORT}:9999" \
      --name wptd-dev-instance wptd-dev
  info "Setting up local user"
  wptd_useradd

  info "Ensuring the home directory is owned by the user..."
  wptd_chown "/home/user"

  info "Instance wptd-dev-instance started."
elif [[ "${RUNNING_STATUS}" != "0" ]]; then
  info "Restarting docker instance wptd-dev-instance..."
  docker start wptd-dev-instance
  info "Instance wptd-dev-instance restarted."
else
  info "Docker instance wptd-dev-instance already running."
  exit 0
fi



trap quit INT

while true; do
    info "Hit Ctrl+C to end"
    read input
    [[ $input == finish ]] && break
    bash -c "$input"
done
