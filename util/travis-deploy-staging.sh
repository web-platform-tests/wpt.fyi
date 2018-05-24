#!/bin/bash

# Helper script for posting a GitHub comment pointing to the deployed environment,
# from Travis CI. Also see deploy.sh

APP_PATH="$@"

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

UTIL_DIR="$(dirname "${BASH_SOURCE[0]}")"
source "${UTIL_DIR}/logging.sh"

if [ "${TRAVIS_SECURE_ENV_VARS}" == "false" ]; then
  info "Travis secrets unavaible. Skipping ${APP_PATH} deployment."
  exit 0
fi

# Skip if nothing under $APP_PATH was modified.
if [ "${FORCE_PUSH}" != "true" ];
then
  git diff --name-only ${TRAVIS_BRANCH}..HEAD | grep "^${APP_PATH}/" || {
    info "No changes detected under ${APP_PATH}. Skipping deployment."
    exit 0
  }
fi

echo "Copying output to ${TEMP_FILE:=$(mktemp)}"
# NOTE: Most gcloud output is stderr, so need to redirect it to stdout.
docker exec -t -u $(id -u $USER):$(id -g $USER) "${DOCKER_INSTANCE}" \
    make deploy_staging \
        APP_PATH=${APP_PATH} \
        BRANCH_NAME=${TRAVIS_PULL_REQUEST_BRANCH:-TRAVIS_BRANCH} 2>&1 \
            | tee ${TEMP_FILE}
if [ "${EXIT_CODE:=${PIPESTATUS[0]}}" != "0" ]; then exit ${EXIT_CODE}; fi
DEPLOYED_URL="$(grep -Po 'Deployed to \K[^\s]+' ${TEMP_FILE} | tr -d '\n')"
${UTIL_DIR}/deploy-comment.sh "${DEPLOYED_URL}"
