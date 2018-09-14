#!/bin/bash

# Helper script for posting a GitHub comment pointing to the deployed environment,
# from Travis CI. Also see deploy.sh

usage() {
  USAGE="Usage: travis-staging-deploy.sh [-f] [app path]
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
if [[ "${APP_PATH}" == "webapp" ]]; then APP_DEPS="${APP_DEPS}|api|shared"; fi
if [[ "${APP_PATH}" == "revisions/service" ]]; then APP_DEPS="${APP_DEPS}|revisions|shared"; fi
APP_DEPS_REGEX="^(${APP_DEPS})/"

EXCLUSIONS="_test.go$$"
if [[ "${APP_PATH}" == "webapp" ]]; then EXCLUSIONS="${EXCLUSIONS}|webapp/components/test/"

UTIL_DIR="$(dirname "${BASH_SOURCE[0]}")"
source "${UTIL_DIR}/logging.sh"

if [ "${TRAVIS_SECURE_ENV_VARS}" != "true" ]; then
  info "Travis secrets unavaible. Skipping ${APP_PATH} deployment."
  exit 0
fi

# Skip if nothing under $APP_PATH was modified.
if [ "${FORCE_PUSH}" != "true" ];
then
  git diff --name-only ${TRAVIS_BRANCH}..HEAD | egrep -v "${EXCLUSIONS}" | egrep "${APP_DEPS_REGEX}" || {
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
        BRANCH_NAME="${TRAVIS_PULL_REQUEST_BRANCH:-$TRAVIS_BRANCH}" 2>&1 \
            | tee ${TEMP_FILE}
if [ "${EXIT_CODE:=${PIPESTATUS[0]}}" != "0" ]; then exit ${EXIT_CODE}; fi
DEPLOYED_URL=$(tr -d "\r" < ${TEMP_FILE} | sed -ne 's/^Deployed service.*to \[\(.*\)\]$/\1/p')

# Add a GitHub comment to the PR (if there is a PR).
if [[ -n "${TRAVIS_PULL_REQUEST_BRANCH}" ]];
then
  ${UTIL_DIR}/deploy-comment.sh "${DEPLOYED_URL}";
fi
