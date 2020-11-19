#!/usr/bin/env bash

DOCKER_DIR=$(dirname $0)
source "${DOCKER_DIR}/../commands.sh"

# Run util/populate_dev_data.go (via make) in the docker environment.
wptd_exec make dev_data FLAGS=\"-project=wptdashboard-local -remote_host=staging.wpt.fyi -datastore_host=localhost:8001\"
