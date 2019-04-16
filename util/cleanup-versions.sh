#!/bin/bash

# Helper script to remove stale versions (without upstream branches) from the
# staging project semi-automatically (users need to confirm before deleting).

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

  # Ensure remote branches are fetched.
  git fetch origin

  for version in $( gcloud app versions list $PROJECT_ARG $SERVICE_ARG --filter="$FILTER_ARG" --format="value(id)" ); do
    if ! git show-ref --quiet --verify refs/remotes/origin/$version; then
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


# Sanity check (script will exit if origin does not exist because of set -e.)
REMOTE_URL=$(git remote get-url origin)
if [[ $REMOTE_URL != *web-platform-tests/wpt.fyi* ]]; then
  fatal "origin isn't web-platform-tests/wpt.fyi" 1
fi

cleanup "default"
cleanup "processor"
cleanup "announcer"
cleanup "searchcache"
