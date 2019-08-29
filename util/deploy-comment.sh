#!/bin/bash

# Helper script for posting a GitHub comment pointing to the deployed environment,
# from Travis CI. Also see deploy.sh

STAGING_URL="$1"

UTIL_DIR="$(dirname "${BASH_SOURCE[0]}")"
source "${UTIL_DIR}/logging.sh"

if [[ -z "${STAGING_URL}" ]];
then fatal "Deployed URL is required";
else debug "Deployed URL: ${STAGING_URL}";
fi
if [[ -z "${GITHUB_TOKEN}" ]];
then fatal "GitHub Token is required";
else debug "GitHub token detected.";
fi
if [[ -z "${TRAVIS_REPO_SLUG}" ]];
then fatal "Travis Repo slug (user/repo) is required";
else debug "Travis Repo slug: ${TRAVIS_REPO_SLUG}";
fi
if [[ -z "${TRAVIS_BRANCH}" ]];
then fatal "Travis branch is required";
else debug "Travis branch: ${TRAVIS_BRANCH}";
fi

set -e
set -o pipefail

info "Posting deployed enviroment to GitHub..."
POST_URL="https://api.github.com/repos/${TRAVIS_REPO_SLUG}/deployments"
debug "${POST_URL}"
POST_BODY="{
                \"ref\": \"${TRAVIS_BRANCH}\",
                \"task\": \"deploy\",
                \"auto_merge\": false,
                \"environment\": \"${APP_PATH}\",
                \"transient_environment\": true
            }"
debug "POST body: ${POST_BODY}"

debug "Copying output to ${TEMP_FILE:=$(mktemp)}"
curl -H "Authorization: token ${GITHUB_TOKEN}" \
     -H "Accept: application/vnd.github.ant-man-preview+json" \
     -X "POST" \
     -d "${POST_BODY}" \
     -s \
     "${POST_URL}" \
     | tee "${TEMP_FILE}"
if [[ "${EXIT_CODE:=${PIPESTATUS[0]}}" != "0" ]]; then exit ${EXIT_CODE}; fi

DEPLOYMENT_ID=$(jq .id ${TEMP_FILE})
if [[ "${EXIT_CODE}" == "0" ]]
then
    debug "Created deployment ${DEPLOYMENT_ID}"
fi

debug "Setting status to deployed"
POST_BODY="{
                \"state\": \"success\",
                \"environment_url\": \"${STAGING_URL}\",
                \"auto_inactive\": true
            }"
curl -H "Authorization: token ${GITHUB_TOKEN}" \
     -H "Accept: application/vnd.github.ant-man-preview+json" \
     -X "POST" \
     -d "${POST_BODY}" \
     -s \
     "${POST_URL}/${DEPLOYMENT_ID}/statuses"
