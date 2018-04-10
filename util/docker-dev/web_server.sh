#!/bin/bash

# Start the Google Cloud web development server in `wptd-dev-instance`
# (started using ./run.sh).

DOCKER_DIR=$(dirname $0)
source "${DOCKER_DIR}/../commands.sh"
source "${DOCKER_DIR}/../logging.sh"
source "${DOCKER_DIR}/../path.sh"
WPTD_PATH=${WPTD_PATH:-$(absdir ${DOCKER_DIR}/../..)}

WPTD_CONTAINER_HOST=0.0.0.0

info "Installing web server code dependencies"
wptd_exec "make build"

DOCKER_STATUS="${?}"
if [ "${DOCKER_STATUS}" != "0" ]; then
  error "Failed to install web server code dependencies"
  exit "${DOCKER_STATUS}"
fi

info "Starting web server. Port forwarded from wptd-dev-instance: 8080"
wptd_exec_it "dev_appserver.py --host $WPTD_CONTAINER_HOST --port=8080 --admin_host=$WPTD_CONTAINER_HOST --admin_port=8000 --api_host=$WPTD_CONTAINER_HOST --api_port=9999 -A=wptdashboard webapp"
