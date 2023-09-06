#!/bin/bash

# Helper script for deploying to production.
# Needs the following packages to be installed: google-cloud-cli, gh, jq

set -e

usage() {
  USAGE="Usage: deploy-production.sh [-f]
    -f : Always deploy (even if checks have failed)"
  echo "${USAGE}"
}

while getopts ':fh' flag; do
  case "${flag}" in
    f) FORCE_DEPLOY='true' ;;
    h|*) usage && exit 0;;
  esac
done

GH_OWNER="web-platform-tests"
GH_REPO="wpt.fyi"

# Find changes to deploy.
LAST_DEPLOYED_SHA=$(gcloud app --project=wptdashboard versions list --hide-no-traffic --filter='service=default' --format=yaml | grep id | head -1 | cut -d' ' -f2 | sed 's/rev-//')
CHANGELIST=$(git log $LAST_DEPLOYED_SHA..HEAD --oneline)
if [[ "${CHANGELIST}" == ""  ]];
then
    echo "No new changes to deploy."
    exit 0
fi
CHANGE_COUNT=$(echo "$CHANGELIST"|wc -l)
echo -e "There are $CHANGE_COUNT changes to deploy:\n$CHANGELIST"

# Verfiy that all commit checks passed.
FAILED_CHECKS=$(gh api /repos/"$GH_OWNER"/"$GH_REPO"/commits/HEAD/check-runs | jq -r '.check_runs | map(select(.conclusion == "failure"))')
FAILURES=$(echo "$FAILED_CHECKS" | jq -r 'length')
if [[ "${FAILURES}" != "0"  ]];
then
    echo -e "\n$FAILURES checks failed for the latest commit:"
    echo "$FAILED_CHECKS" | jq -r '.[] | .name + ": " + .html_url'
    if [[ "${FORCE_DEPLOY}" != "true" ]];
    then
        echo -e "\nVisit the link(s) above and if failed checks should not block deployment, run the script again with -f"
        exit 1
    fi
fi

# File a deployment bug.
NEW_SHA=$(git rev-parse --short HEAD)
PROD_LABEL="prod"
RELEASE_LABEL="release"
LAST_DEPLOYMENT_ISSUE=$(gh issue list --state closed --label "$PROD_LABEL" --label "$RELEASE_LABEL" --limit 1 --json number --jq '.[] | .number')
BUG_TITLE="Deploy $NEW_SHA to production"
BUG_BODY=$(cat << EOF
Previous deployment was #$LAST_DEPLOYMENT_ISSUE ($LAST_DEPLOYED_SHA)

Changelist $LAST_DEPLOYED_SHA...$NEW_SHA

Changes:
$CHANGELIST

This push is happening as part of the regular weekly push.

Pushing all three services - webapp, processor, and searchcache.
EOF
)

gh issue create --title "$BUG_TITLE" --body "$BUG_BODY" --label "$PROD_LABEL" --label "$RELEASE_LABEL"
if [[ $? != 0 ]];
then
    echo "GitHub issue creation failed"
    exit 2
fi

# TODO(past): automate the remaining steps
echo "Until the process if fully automated, continue with the manual deployment process starting at step 4."
