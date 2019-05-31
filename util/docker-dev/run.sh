#!/bin/bash

# Start Docker-based development server as `wptd-dev-instance` in the
# foreground.

DOCKER_DIR=$(dirname $0)
source "${DOCKER_DIR}/../commands.sh"
source "${DOCKER_DIR}/../logging.sh"
source "${DOCKER_DIR}/../path.sh"
WPT_PATH=${WPT_PATH:-$(absdir ${DOCKER_DIR}/../../..)}
WPTD_PATH="${WPT_PATH}/wpt.fyi"

DOCKER_INSTANCE="${DOCKER_INSTANCE:-wptd-dev-instance}"

WPTD_HOST_WEB_PORT=${WPTD_HOST_WEB_PORT:-"8080"}
WPTD_HOST_ADMIN_WEB_PORT=${WPTD_HOST_ADMIN_WEB_PORT:-"8000"}
WPTD_HOST_API_WEB_PORT=${WPTD_HOST_API_WEB_PORT:-"9999"}

function usage() {
  USAGE="USAGE: $(basename ${0}) [-q] [-a] [-d]
    -d  daemon mode: Run in the background rather than blocking then cleaning up
    -q  quiet mode: Assume default for all prompts"
  >&2 echo "${USAGE}"
}

PR=""
function confirm_preserve_remove() {
  if confirm "${1}. Remove?"; then
    PR="r"
  else
    PR="p"
  fi
}

DAEMON="false"
QUIET="false"
while getopts ':dhaq' FLAG; do
  case "${FLAG}" in
    d)
      DAEMON="true" ;;
    q)
      QUIET="true"
      PR="r" ;;
    h|*) usage && exit 0 ;;
  esac
done

info "Creating docker instance for dev server. Instance name: wptd-dev-instance"
docker inspect "${DOCKER_INSTANCE}" > /dev/null 2>&1
INSPECT_STATUS="${?}"
docker inspect "${DOCKER_INSTANCE}" | grep '"Running": true' | read
RUNNING_STATUS="${?}"

function stop() {
  info "Stopping ${DOCKER_INSTANCE}..."
  wptd_stop
  info ""${DOCKER_INSTANCE}" stopped."
  if [[ "${PR}" == "" ]]; then
    confirm_preserve_remove "Docker instance ${DOCKER_INSTANCE} still exists"
  fi
  if [[ "${PR}" == "r" ]]; then
    info "Removing ${DOCKER_INSTANCE}..."
    wptd_rm
    info "${DOCKER_INSTANCE} removed."
  fi
}

function quit() {
  warn "run.sh: Recieved interrupt. Exiting..."
  stop
  exit 0
}

if [ "${INSPECT_STATUS}" == "0" ]; then
  if [[ "${PR}" == "" ]]; then
    confirm_preserve_remove "Found existing docker instance ${DOCKER_INSTANCE}"
  fi
  if [[ "${PR}" == "r" ]]; then
    stop
  fi
fi

set -e

# Create a docker instance:
#
# -t                                     Give the container a TTY
# -v "${WPTD_PATH}":/wpt.fyi             Mount the repository
# -u $(id -u $USER)                      Run as current user
# --cap-add=SYS_ADMIN                    Allow Chrome to use sandbox:
#   https://github.com/GoogleChrome/puppeteer/blob/master/docs/troubleshooting.md
# -p "${WPTD_HOST_WEB_PORT}:8080"        Expose web server port
# --name "${DOCKER_INSTANCE}"            Name the instance
# wptd-dev                               Identify image to use

VOLUMES="-v ${WPTD_PATH}:/home/user/wpt.fyi"

if [[ "${INSPECT_STATUS}" != 0 ]] || [[ "${PR}" == "r" ]]; then
  info "Starting docker instance ${DOCKER_INSTANCE}..."
  docker run -t -d --entrypoint /bin/bash \
      ${VOLUMES} \
      -u $(id -u $USER) \
      --cap-add=SYS_ADMIN \
      -p "${WPTD_HOST_WEB_PORT}:8080" \
      -p "${WPTD_HOST_ADMIN_WEB_PORT}:8000" \
      -p "${WPTD_HOST_API_WEB_PORT}:9999" \
      --name "${DOCKER_INSTANCE}" wptd-dev
  info "Setting up local user"
  wptd_useradd

  info "Ensuring the home directory is owned by the user..."
  wptd_chown "/home/user"

  info "Instance ${DOCKER_INSTANCE} started."
elif [[ "${RUNNING_STATUS}" != "0" ]]; then
  info "Restarting docker instance ${DOCKER_INSTANCE}..."
  docker start "${DOCKER_INSTANCE}"
  info "Instance ${DOCKER_INSTANCE} restarted."
else
  info "Docker instance ${DOCKER_INSTANCE} already running."
  exit 0
fi

info "Updating system/packages..."
wptd_exec make sys_deps

if [[ "${DAEMON}" == "true" ]]; then
  exit 0
fi

trap quit INT

while true; do
    info "Hit Ctrl+C to end"
    read input
    [[ $input == finish ]] && break
    bash -c "$input"
done
