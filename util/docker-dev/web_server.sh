#!/bin/bash

# Start the Google Cloud web development server in `wptd-dev-instance`
# (started using ./run.sh).

DOCKER_DIR=$(dirname $0)
source "${DOCKER_DIR}/../commands.sh"
source "${DOCKER_DIR}/../logging.sh"

set -e

info "Building web server..."
wptd_exec make go_build

DOCKER_STATUS="${?}"
if [ "${DOCKER_STATUS}" != "0" ]; then
  error "Failed to install web server code dependencies"
  exit "${DOCKER_STATUS}"
fi

info "Starting web server. Port forwarded to host: ${WPTD_HOST_WEB_PORT}"
wptd_exec "\$(gcloud beta emulators datastore env-init) && ./web"
