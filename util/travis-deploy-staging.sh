#!/bin/bash

# Helper script for posting a GitHub comment pointing to the deployed environment,
# from Travis CI. Also see deploy.sh

APP_PATH=$1

UTIL_DIR="$(dirname $0)"
source "${UTIL_DIR}/logging.sh"

if [ "${TRAVIS_SECURE_ENV_VARS}" == "false" ]; then
  info "Travis secrets unavaible. Skipping ${APP_PATH} deployment."
  exit 0
fi

if [ -z "${TRAVIS_PULL_REQUEST_BRANCH}"]; then
  info "Not on a PR. Skipping ${APP_PATH} deployment."
  exit 0
fi

# Skip if webapp isn't modified.
git diff --name-only FETCH_HEAD...${TRAVIS_PULL_REQUEST_BRANCH} | grep ^$APP_PATH/ || {
  info "No changes detected under ${APP_PATH}. Skipping deployment."
  exit 0
}

echo "Copying output to ${TEMP_FILE:=$(mktemp)}"
# NOTE: Most gcloud output is stderr, so need to redirect it to stdout.
docker exec -t -u $(id -u $USER):$(id -g $USER) "${DOCKER_INSTANCE}" make deploy_staging APP_PATH=${APP_PATH} BRANCH_NAME=${TRAVIS_PULL_REQUEST_BRANCH} 2>&1 | tee ${TEMP_FILE}
if [ "${EXIT_CODE:=${PIPESTATUS[0]}}" != "0" ]; then exit ${EXIT_CODE}; fi
DEPLOYED_URL="$(grep -Po 'Deployed to \K[^\s]+' ${TEMP_FILE} | tr -d '\n')"
${UTIL_DIR}/deploy-comment.sh "${DEPLOYED_URL}"
