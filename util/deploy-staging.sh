#!/bin/bash

# Helper script for posting a GitHub comment pointing to the deployed environment,
# from GitHub Actions. Also see deploy.sh

usage() {
  USAGE="Usage: deploy-staging.sh [-f] [app path]
    -f : Always deploy (even if no changes detected)
    app path: wpt.fyi relative path for the app, e.g. \"webapp\""
  echo "${USAGE}"
}

APP_PATH=${@: -1}
while getopts ':fhq' flag; do
  case "${flag}" in
    f) FORCE_PUSH='true' ;;
    h|*) usage && exit 0;;
  esac
done

if [[ "${APP_PATH}" == ""  ]]; then fatal "app path not specified."; fi

APP_DEPS="${APP_PATH}"
if [[ "${APP_PATH}" == webapp/web* ]]; then APP_DEPS="webapp|api|shared"; fi
# Be more conservative: only deploy searchcache when it or shared are modified.
if [[ "${APP_PATH}" == api/query/cache/service* ]]; then APP_DEPS="api/query|shared"; fi
if [[ "${APP_PATH}" == "results-processor/app.staging.yaml" ]]; then APP_DEPS="results-processor"; fi
APP_DEPS_REGEX="^(${APP_DEPS})/"

EXCLUSIONS="_test.go$|webapp/components/test/"

UTIL_DIR="$(dirname "${BASH_SOURCE[0]}")"
source "${UTIL_DIR}/logging.sh"

# Skip if nothing under $APP_PATH was modified.
if [ "${FORCE_PUSH}" != "true" ];
then
  git diff --name-only HEAD^..HEAD | egrep -v "${EXCLUSIONS}" | egrep "${APP_DEPS_REGEX}" || {
    info "No changes detected under ${APP_DEPS}. Skipping deploying ${APP_PATH}."
    exit 0
  }
fi

debug "Copying output to ${TEMP_FILE:=$(mktemp)}"
# NOTE: Most gcloud output is stderr, so need to redirect it to stdout.
docker exec -t -u $(id -u $USER):$(id -g $USER) "${DOCKER_INSTANCE}" \
    make deploy_staging \
        PROJECT=wptdashboard-staging \
        APP_PATH="${APP_PATH}" \
        BRANCH_NAME="${GITHUB_HEAD_REF:-$GITHUB_REF}" 2>&1 \
            | tee ${TEMP_FILE}
if [ "${EXIT_CODE:=${PIPESTATUS[0]}}" != "0" ]; then exit ${EXIT_CODE}; fi
DEPLOYED_URL=$(tr -d "\r" < ${TEMP_FILE} | sed -ne 's/^Deployed service.*to \[\(.*\)\]$/\1/p')

# TODO(kyle): Fix deploy-comment.sh; rewrite to GitHub Actions equivalent.
# Add a GitHub comment to the PR (if there is a PR).
#if [[ -n "${GITHUB_HEAD_REF}" ]];
#then
#  ${UTIL_DIR}/deploy-comment.sh -e "${APP_PATH}" "${DEPLOYED_URL}";
#fi
