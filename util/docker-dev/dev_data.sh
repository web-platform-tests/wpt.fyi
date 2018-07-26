#!/usr/bin/env bash

DOCKER_DIR=$(dirname $0)
source "${DOCKER_DIR}/../commands.sh"

# Run util/populate_dev_data.go (via make) in the docker environment.
wptd_exec make dev_data FLAGS=\"$@\"
