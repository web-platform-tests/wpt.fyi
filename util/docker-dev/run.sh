#!/bin/bash

# Start Docker-based development server as `wptd-dev-instance` in the
# foreground.

DOCKER_DIR=$(dirname $0)
source "${DOCKER_DIR}/../commands.sh"
source "${DOCKER_DIR}/../logging.sh"

function usage() {
  USAGE="USAGE: $(basename ${0}) [-q] [-d] [-s]
    -d  daemon mode: Run in the background rather than blocking then cleaning up
    -q  quiet mode: Assume default for all prompts
    -s  stop daemon: Stop a running daemon"
  >&2 echo "${USAGE}"
}

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
while getopts ':dhqs' FLAG; do
  case "${FLAG}" in
    d)
      DAEMON="true" ;;
    s)
      stop && exit 0 ;;
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

function quit() {
  warn "run.sh: Received interrupt. Exiting..."
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
#   https://github.com/GoogleChrome/puppeteer/blob/main/docs/troubleshooting.md
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
      -p "${WPTD_HOST_GCD_PORT}:8001" \
      -p "12345:12345" \
      --workdir "/home/user/wpt.fyi" \
      --name "${DOCKER_INSTANCE}" \
      ${DOCKER_IMAGE}
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

trap quit INT

info "Updating system/packages..."
wptd_exec make sys_deps

info "Installing dev dependencies..."
wptd_exec make dev_appserver_deps

if [[ "${DAEMON}" == "true" ]]; then
  exit 0
fi

info "Starting Cloud Datastore emulator. Port forwarded to host: ${WPTD_HOST_GCD_PORT}"
info "=== Hit Ctrl+C to end ==="
wptd_exec gcloud beta emulators datastore start \
  --project=wptdashboard-local \
  --consistency=1.0 \
  --host-port=localhost:8001 2> /dev/null
