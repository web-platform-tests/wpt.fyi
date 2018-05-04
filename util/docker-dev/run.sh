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

set -e

DOCKER_INSTANCE_PID=""
if [ "${INSPECT_STATUS}" != "0" ]; then
  info "Docker instance wptd-dev-instance not found. Starting it..."
  docker run -t -d --entrypoint /bin/bash \
      -v "${WPTD_PATH}":/home/jenkins/wpt.fyi \
      -u $(id -u $USER):$(id -g $USER) \
      -p "${WPTD_HOST_WEB_PORT}:8080" \
      -p "${WPTD_HOST_ADMIN_WEB_PORT}:8000" \
      -p "${WPTD_HOST_API_WEB_PORT}:9999" \
      --name wptd-dev-instance wptd-dev
  DOCKER_INSTANCE_PID="${!}"
else
  info "Found existing docker instance wptd-dev-instance"
fi

info "Ensuring current users has root..."
wptd_chown "/home/jenkins"

function stop() {
  warn "run.sh: Recieved interrupt. Exiting..."
  info "Stopping wptd-dev-instance..."
  wptd_stop
  info "Removing wptd-dev-instance..."
  wptd_rm
  exit 0
}

info "Instance wptd-dev-instance started."

trap stop INT

while true; do
    info "Hit Ctrl+C to end"
    read input
    [[ $input == finish ]] && break
    bash -c "$input"
done
