#!/bin/bash

# Start the Google Cloud web development server in `wptd-dev-instance`
# (started using ./run.sh).

usage() {
  USAGE="Usage: web_server.sh [-d]
    -d : Start a debugging session with Delve"
  echo "${USAGE}"
}

while getopts ':dh' flag; do
  case "${flag}" in
    d) DEBUG='-d' ;;
    h|*) usage && exit 0;;
  esac
done

DOCKER_DIR=$(dirname $0)
source "${DOCKER_DIR}/../commands.sh"
source "${DOCKER_DIR}/../logging.sh"

set -e

if [[ ${DEBUG} != "true" ]];
then
  wptd_exec make inotifywait
fi
info "Building web server..."
# Build the full go_build target to get node_modules.
wptd_exec make go_build

DOCKER_STATUS="${?}"
if [ "${DOCKER_STATUS}" != "0" ]; then
  error "Failed to install web server code dependencies"
  exit "${DOCKER_STATUS}"
fi

info "Starting web server. Port forwarded to host: ${WPTD_HOST_WEB_PORT}"
wptd_exec_it "\$(gcloud beta emulators datastore env-init) && util/server-watch.sh ${DEBUG}"
