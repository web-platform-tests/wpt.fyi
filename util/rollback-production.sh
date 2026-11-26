#!/bin/bash
# Copyright 2026 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

# Moves production traffic for all services back to an earlier version.
#
# After a release serves steadily, deploy-production.sh stops the previous
# processor and searchcache versions (the previous webapp version keeps
# running). This script starts every stopped version at the same time, since
# each start takes a few minutes, and moves each service's traffic as soon as
# its version is running, webapp first. If the webapp version is still
# running, the site rolls back right away. Results that arrive while the
# processor starts wait in the queue, and a searchcache that was just started
# reloads its runs in the background, like after any deploy.
#
# Usage: util/rollback-production.sh <version>
#   version   the version to roll back to, e.g. rev-9f9e929a. List them with:
#             gcloud app versions list --project=wptdashboard

set -euo pipefail

UTIL_DIR="$(dirname "${BASH_SOURCE[0]}")"
source "${UTIL_DIR}/logging.sh"

PROJECT="wptdashboard"
SERVICES="default processor searchcache"
VERSION="${1:-}"

if [[ -z "${VERSION}" || "${VERSION}" == "-h" ]]; then
  echo "Usage: util/rollback-production.sh <version>"
  echo "Available versions:"
  gcloud app versions list --project="${PROJECT}" \
    --format='table(service,id,version.servingStatus,traffic_split)'
  exit 1
fi

# Check that the version exists for every service before changing anything.
serving_status() {
  gcloud app versions describe "${VERSION}" --project="${PROJECT}" \
    --service="$1" --format='value(servingStatus)' 2>/dev/null
}
for SERVICE in ${SERVICES}; do
  serving_status "${SERVICE}" > /dev/null \
    || fatal "Version ${VERSION} of service ${SERVICE} does not exist."
done

# Start the stopped versions in the background, all at once. Each start writes
# its output and process ID to files named after the service.
START_DIR="$(mktemp -d)"
trap 'rm -rf "${START_DIR}"' EXIT
for SERVICE in ${SERVICES}; do
  if [[ "$(serving_status "${SERVICE}")" == "STOPPED" ]]; then
    info "Starting ${SERVICE} version ${VERSION} (this takes a few minutes)..."
    gcloud app versions start "${VERSION}" \
      --project="${PROJECT}" --service="${SERVICE}" --quiet \
      > "${START_DIR}/${SERVICE}.log" 2>&1 &
    echo $! > "${START_DIR}/${SERVICE}.pid"
  fi
done

for SERVICE in ${SERVICES}; do
  if [[ -f "${START_DIR}/${SERVICE}.pid" ]]; then
    if ! wait "$(cat "${START_DIR}/${SERVICE}.pid")"; then
      cat "${START_DIR}/${SERVICE}.log"
      fatal "Could not start ${SERVICE} version ${VERSION}; its traffic was not moved."
    fi
    info "Started ${SERVICE} version ${VERSION}."
  fi
  info "Moving all ${SERVICE} traffic to ${VERSION}..."
  gcloud app services set-traffic "${SERVICE}" \
    --project="${PROJECT}" --splits="${VERSION}=1" --quiet
done

info "All services now serve ${VERSION}."
