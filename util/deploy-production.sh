#!/bin/bash

# Helper script for deploying to production.
# Needs the following packages to be installed: google-cloud-cli, gh, jq

#set -x #echo on for debugging purposes
set -e

# Pull the latest version of the docker image
docker pull webplatformtests/wpt.fyi:latest 

usage() {
  USAGE="Usage: deploy-production.sh [-f] [-b] [-q]
    -b : Skip GitHub issue creation
    -f : Always deploy (even if checks have failed)
    -q : Disable all interactive prompts and debugging output when running gcloud deploy commands"
  echo "${USAGE}"
}

while getopts ':bfqh' flag; do
  case "${flag}" in
    b) SKIP_ISSUE_CREATION='true' ;;
    f) FORCE_DEPLOY='true' ;;
    q) QUIET='true' ;;
    h|*) usage && exit 0;;
  esac
done

GH_OWNER="web-platform-tests"
GH_REPO="wpt.fyi"
PROD_LABEL="prod"
RELEASE_LABEL="release"
UTIL_DIR=$(dirname $0)
source "${UTIL_DIR}/logging.sh"
source "${UTIL_DIR}/commands.sh"

if [[ ${SKIP_ISSUE_CREATION} != "true" ]];
then
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
  MAIN_SHA=$(git rev-parse main)
  FAILED_CHECKS=$(gh api /repos/"$GH_OWNER"/"$GH_REPO"/commits/$MAIN_SHA/check-runs | jq -r '.check_runs | map(select(.conclusion == "failure" and .name != "Dependabot"))')
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
fi

# Confirm there are no more than two versions for each service to make sure
# there's room for the ones we're about to push. If there are more than two
# versions available, something didn't go as planned in the previous
# deployment. If so, delete old versions manually in the cloud console.
SERVICES="default processor searchcache"
for SERVICE in $SERVICES
do
  VERSIONS=$(gcloud app --project=wptdashboard versions list --filter="service=$SERVICE" --format=list | wc -l)
  if ((${VERSIONS} > 2));
  then
    echo -e "Found more than 2 versions ($VERSIONS) for service $SERVICE.\nPlease make sure there are no more than 2 versions of each service and try\nagain."

    exit 3
  fi

  echo "Found $VERSIONS versions for service $SERVICE. Good to proceed."
done

# Start a docker instance.
${UTIL_DIR}/docker-dev/run.sh -d
# Login to gcloud if not already logged in.
wptd_exec_it gcloud auth login
# Deploy the services.
wptd_exec_it make deploy_production PROJECT=wptdashboard APP_PATH=webapp/web ${QUIET:+QUIET=true}
wptd_exec_it make deploy_production PROJECT=wptdashboard APP_PATH=results-processor ${QUIET:+QUIET=true}
wptd_exec_it make deploy_production PROJECT=wptdashboard APP_PATH=api/query/cache/service ${QUIET:+QUIET=true}
cd webapp/web
gcloud app deploy ${QUIET:+--quiet} --project=wptdashboard index.yaml queue.yaml dispatch.yaml
cd ../..

# Stop docker.
${UTIL_DIR}/docker-dev/run.sh -s

# Confirm that everything works as expected and redirect traffic.
VERSION_URL=$(gcloud app --project=wptdashboard versions list --sort-by=~last_deployed_time --filter='service=default' --limit=1 --format=json | jq -r '.[] | .version.versionUrl')
LATEST_VERSION=$(gcloud app --project=wptdashboard versions list --sort-by=~last_deployed_time --filter='service=default' --limit=1 --format=json | jq -r '.[] | .id')
MESSAGE="Visit $VERSION_URL to confirm that everything works (page load, search, test expansion, show history). Wait 15 minutes before redirecting traffic (https://cloud.google.com/appengine/docs/flexible/known-issues). Redirect traffic now?"
if confirm "$MESSAGE"; then
  for SERVICE in $SERVICES
  do
    gcloud app --project=wptdashboard services set-traffic $SERVICE --splits $LATEST_VERSION=1
  done
else
  echo "Don't forget to migrate traffic to the new version."
fi

# Update and close deployment bug.
LAST_DEPLOYMENT_ISSUE=$(gh issue list --state open --label "$PROD_LABEL" --label "$RELEASE_LABEL" --limit 1 --json number --jq '.[] | .number')
gh issue close "$LAST_DEPLOYMENT_ISSUE" -c "Deployment is now complete."

# Check if there are more more than two versions of the default service left
# after we're done with this deplyment to make sure there's room for the next
# deployment. If there are, ask to delete the oldest default service version,
# and also delete the same version from the other services which will also exist
# if all went well during the deployment. This check isn't fail safe, but
# combined with the check we do before doing any deployments earlier in this
# script, this should leave us in a good state.

VERSIONS=$(gcloud app --project=wptdashboard versions list --filter="service=default" --format=list | wc -l)

if (($VERSIONS == 3)); then
  echo -e "Please ensure the deployment was successful. If so, we can go ahead and\ndelete the oldest version of all services if necessary, leaving the one just\ndeployed and the one running before this deployment. This will ensure we leave\nroom for the next deployment.\n"

  read -p "Delete oldest version of all services to leave room for the next deplyment? (y/n): " DELETE

  if [[ $DELETE == "y" ]]; then
    echo "Found $VERSIONS for the default service, deleting the oldest version of all services."

    OLDEST_REV=$(gcloud app --project=wptdashboard versions list --sort-by=last_deployed_time --filter="service=default" --limit=1 --format=json | jq -r '.[] | .id')
    for SERVICE in $SERVICES; do
      echo "Deleting $SERVICE service version $OLDEST_REV"
      gcloud app --project=wptdashboard versions delete --service=$SERVICE --quiet $OLDEST_REV
    done
  fi
elif (($VERSIONS > 3)); then
  echo -e "\nUnexpectedly found $VERSIONS versions for the default service.\nPlease delete old versions for all services manually until there are no more than two left."
fi
