#!/bin/bash

# Start the Google Cloud web development server in `wptd-dev-instance`
# (started using ./run.sh).

DOCKER_DIR=$(dirname $0)
source "${DOCKER_DIR}/../commands.sh"
source "${DOCKER_DIR}/../logging.sh"
source "${DOCKER_DIR}/../path.sh"
WPTD_PATH=${WPTD_PATH:-$(absdir ${DOCKER_DIR}/../..)}


info "Selecting gcloud project: wptdashboard"
wptd_exec "gcloud config set project wptdashboard"

info "Checking application default credentials"
wptd_exec "gcloud auth application-default print-access-token"

DOCKER_STATUS="${?}"
if [ "${DOCKER_STATUS}" != "0" ]; then
  warn "No credentials yet. Logging in..."
  wptd_exec_it "gcloud auth application-default login"

  DOCKER_STATUS="${?}"
  if [ "${DOCKER_STATUS}" != "0" ]; then
    error "Failed to get application default credentials"
    exit "${DOCKER_STATUS}"
  fi
fi
info "Application default credentials installed"
exit "${DOCKER_STATUS}"
