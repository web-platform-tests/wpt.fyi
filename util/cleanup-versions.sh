#!/bin/bash

# Helper script to remove stale versions (without upstream branches) from the
# staging project, and to stop previews that haven't been deployed recently.
# Runs after every push to main (deploy.yml) and daily (cleanup-staging.yml).

set -e

UTIL_DIR="$(dirname "${BASH_SOURCE[0]}")"
source "${UTIL_DIR}/logging.sh"

# A safety cutoff. Only versions last deployed more than 1 day ago may be
# deleted.
CUTOFF="-P1D"

# This is a constant instead of an argument because production versions should
# be deleted carefully and manually.
PROJECT_ARG="--project=wptdashboard-staging"

function cleanup() {
  info "Cleaning stale versions of $1..."

  local SERVICE_ARG="-s $1"
  local FILTER_ARG="traffic_split=0.0 last_deployed_time.datetime<$CUTOFF"
  local versions_to_delete=()

  for version in $( gcloud app versions list $PROJECT_ARG $SERVICE_ARG --filter="$FILTER_ARG" --format="value(id)" ); do
    if ! git ls-remote --exit-code origin $version; then
      debug "'$version' is not a branch in upstream and will be deleted."
      versions_to_delete+=($version)
    fi
  done

  if [[ ${#versions_to_delete[*]} == 0 ]]; then
    debug "Nothing to do"
    return 0
  fi

  gcloud app versions delete --quiet $PROJECT_ARG $SERVICE_ARG ${versions_to_delete[*]}
}

# Previews whose branch still exists keep running, and App Engine bills a
# running version even when it receives no traffic. Stop (but don't delete)
# the ones that haven't been deployed for IDLE_CUTOFF. The next push to the
# branch deploys the preview again, which starts it; to start one without a
# push, run:
#   gcloud app versions start --project=wptdashboard-staging -s SERVICE VERSION
IDLE_CUTOFF="-P3D"

function stop_idle() {
  info "Stopping idle versions of $1..."

  local filter="traffic_split=0.0 version.servingStatus=SERVING last_deployed_time.datetime<$IDLE_CUTOFF"
  local versions_to_stop
  versions_to_stop=$(gcloud app versions list "$PROJECT_ARG" -s "$1" --filter="$filter" --format="value(id)")

  if [[ -z "$versions_to_stop" ]]; then
    debug "Nothing to do"
    return 0
  fi

  debug "Not deployed since before $IDLE_CUTOFF, stopping: ${versions_to_stop//$'\n'/ }"
  # shellcheck disable=SC2086 # One version ID per word.
  gcloud app versions stop --quiet "$PROJECT_ARG" -s "$1" $versions_to_stop
}


# Sanity check (script will exit if origin does not exist because of set -e.)
REMOTE_URL=$(git remote get-url origin)
if [[ $REMOTE_URL != *web-platform-tests/wpt.fyi* ]]; then
  fatal "origin isn't web-platform-tests/wpt.fyi" 1
fi

cleanup "default"
cleanup "processor"
cleanup "searchcache"

stop_idle "default"
stop_idle "processor"
stop_idle "searchcache"
