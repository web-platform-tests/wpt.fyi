#!/usr/bin/env bash

DOCKER_DIR=$(dirname $0)
source "${DOCKER_DIR}/../commands.sh"

# Run util/populate_dev_data.go (via make) in the docker environment.
wptd_exec "\$(gcloud beta emulators datastore env-init) && make dev_data FLAGS=\"-remote_host=staging.wpt.fyi\""
