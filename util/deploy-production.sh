#!/bin/bash

# Helper script for deploying to production.
# Needs the following packages to be installed: google-cloud-cli, gh, jq

#set -x #echo on for debugging purposes
set -e

usage() {
  USAGE="Usage: deploy-production.sh [-f]
    -f : Always deploy (even if checks have failed)"
  echo "${USAGE}"
}

# Deletes the service passed as a parameter.
delete_oldest_version() {
  OLDEST_REV=$(gcloud app --project=wptdashboard versions list --sort-by=last_deployed_time --filter="service=$1" --limit=1 --format=json | jq -r '.[] | .id')
  echo "Deleting $1 service version $OLDEST_REV"
  if confirm "Delete $1 service version $OLDEST_REV?"; then
    gcloud app versions delete --service=$SERVICE $OLDEST_REV
  else
    echo "Skipping $1 service version $OLDEST_REV"
  fi
}

while getopts ':fh' flag; do
  case "${flag}" in
    f) FORCE_DEPLOY='true' ;;
    h|*) usage && exit 0;;
  esac
done

GH_OWNER="web-platform-tests"
GH_REPO="wpt.fyi"
UTIL_DIR=$(dirname $0)
source "${UTIL_DIR}/logging.sh"
source "${UTIL_DIR}/commands.sh"

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

# Confirm there are 3 versions for each service and delete the oldest version.
SERVICES="default processor searchcache"
for SERVICE in $SERVICES
do
  VERSIONS=$(gcloud app --project=wptdashboard versions list --filter="service=$SERVICE" --format=list | wc -l)
  if [[ "${VERSIONS}" -eq "3"  ]];
  then
    echo "Found 3 versions for service $SERVICE, will delete the oldest"
    delete_oldest_version $SERVICE
  elif [[ "${VERSIONS}" -lt "3"  ]];
  then
    echo -e "\n$VERSIONS versions found for service $SERVICE"
  else
    echo -e "\n$VERSIONS versions found for service $SERVICE!"
    exit 3
  fi
done

# Start a docker instance.
${UTIL_DIR}/docker-dev/run.sh -d
# Login to gcloud if not already logged in.
wptd_exec_it gcloud auth login
# Deploy the services.
wptd_exec_it make deploy_production PROJECT=wptdashboard APP_PATH=webapp/web
wptd_exec_it make deploy_production PROJECT=wptdashboard APP_PATH=results-processor
wptd_exec_it make deploy_production PROJECT=wptdashboard APP_PATH=api/query/cache/service
cd webapp/web
gcloud app deploy --project=wptdashboard index.yaml queue.yaml dispatch.yaml
cd ../..

# Stop docker.
${UTIL_DIR}/docker-dev/run.sh -s

# Confirm that everything works as expected and redirect traffic.
VERSION_URL=$(gcloud app --project=wptdashboard versions list --sort-by=~last_deployed_time --filter='service=default' --limit=1 --format=json | jq -r '.[] | .version.versionUrl')
LATEST_VERSION=$(gcloud app --project=wptdashboard versions list --sort-by=~last_deployed_time --filter='service=default' --limit=1 --format=json | jq -r '.[] | .id')
MESSAGE="Visit $VERSION_URL to confirm that everything works (page load, search, test expansion, show history). Redirect traffic now?"
if confirm "$MESSAGE"; then
  for SERVICE in $SERVICES
  do
    gcloud app services set-traffic $SERVICE --splits $LATEST_VERSION=1
  done
else
  echo "Don't forget to migrate traffic to the new version."
fi

# Update and close deployment bug.
LAST_DEPLOYMENT_ISSUE=$(gh issue list --state open --label "$PROD_LABEL" --label "$RELEASE_LABEL" --limit 1 --json number --jq '.[] | .number')
gh issue close "$LAST_DEPLOYMENT_ISSUE" -c "Deployment is now complete."
