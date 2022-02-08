#!/bin/bash

# Helper script for using a standardized version flag when deploying.

REPO_DIR="$(git rev-parse --show-toplevel)"
source "${REPO_DIR}/util/logging.sh"
WPTD_PATH=${WPTD_PATH:-"${REPO_DIR}"}

usage() {
  USAGE="Usage: deploy.sh [-p] [-r] [-q] [-b] [-h] [app path]
    -p : Promote (i.e. pass --promote flag to deploy)
    -r : Release (use the git-hash as the version)
    -q : Quiet (no user prompts, debugging off)
    -b : Branch name - defaults to current Git branch
    -h : Show (this) help information
    app path: wpt.fyi relative path for the app, e.g. \"webapp\""
  echo "${USAGE}"
}

# Take the last argument.
APP_PATH=${@: -1}
# Trim the trailing slash (if any).
APP_PATH=${APP_PATH%/}

while getopts ':b:prhq' flag; do
  case "${flag}" in
    r) RELEASE='true' ;;
    b) BRANCH_NAME="${OPTARG}" ;;
    p) PROMOTE='true' ;;
    q) QUIET='true' ;;
    :) echo "Option -$OPTARG requires an argument." && exit 1;;
    h|*) usage && exit 1;;
  esac
done

if [[ "${APP_PATH}" == ""  ]]; then fatal "app path not specified."; fi
case "${APP_PATH}" in
  "webapp/web" | \
  "webapp/web/app.staging.yaml" | \
  "results-processor" | \
  "api/query/cache/service" | \
  "api/query/cache/service/app.staging.yaml")
  ;;
*)
  fatal "Unrecognized app path \"${APP_PATH}\"."
  ;;
esac

# Ensure dependencies are installed.
if [[ -z "${QUIET}" ]]; then info "Installing dependencies..."; fi
cd "${WPTD_PATH}"
if [[ "${APP_PATH}" == "webapp" ]]; then
  make deployment_state || fatal "Error installing deps"
fi

format_branch_name() {
  local BRANCH="$1"
  # Normalize to lower-case, and replace all non-alphanumeric characters
  # with '-'.
  BRANCH="$(echo -n $BRANCH | tr [:upper:] [:lower:] | tr -c [:alnum:] -)"

  # Limit version names to 22 characters, to leave enough space for the HTTPS
  # domain name (which could have a very long suffix like
  # "-dot-searchcache-wptdashboard-staging"). Domain name parts can be no
  # longer than 63 characters in total.
  BRANCH="$(echo -n $BRANCH | cut -c 1-22)"

  # GCP requires that the branch name start and end in a letter or digit.
  while [[ "$BRANCH" == -* ]]; do
    BRANCH="${BRANCH:1}"
  done
  while [[ "$BRANCH" == *- ]]; do
    BRANCH="${BRANCH::-1}"
  done

  echo $BRANCH
}

# Create a name for this version
VERSION_BRANCH_NAME="$(format_branch_name "${BRANCH_NAME:-"$(git rev-parse --abbrev-ref HEAD)"}")"
USER="$(git remote -v get-url origin | sed -E 's#(https?:\/\/|git@)github.com(\/|:)##' | sed 's#/.*$##')-"
if [[ "${USER}" == "web-platform-tests-" ]]; then USER=""; fi

VERSION="${USER}${VERSION_BRANCH_NAME}"
if [[ -n ${RELEASE} ]]
then
  # Use SHA for releases.
  # Add a prefix to prevent gcloud from interpreting the version name as a number.
  VERSION="rev-$(git rev-parse --short HEAD)"
fi

PROMOTE_FLAG="--no-promote"
if [[ -n ${PROMOTE} ]]
then
  PROMOTE_FLAG="--promote"
  if [[ -z "${QUIET}" ]]; then debug "Producing production configuration..."; fi
  if [[ "${USER}" != "" ]]
  then
    if [[ -z "${QUIET}" ]]
    then
      confirm "Are you sure you want to be deploying a non-web-platform-tests repo (${USER})?"
      if [ "${?}" != "0" ]; then fatal "User cancelled the deploy"; fi
    fi
  fi
fi

if [[ -n "${QUIET}" ]]
then
    QUIET_FLAG="-q"
else
    QUIET_FLAG=""
fi
COMMAND="gcloud app deploy ${PROMOTE_FLAG} ${QUIET_FLAG} --version=${VERSION} ${APP_PATH}"

if [[ -z "${QUIET}" ]]; then info "Executing deploy command:\n${COMMAND}"; fi

set -e

${COMMAND} || fatal "Deploy returned non-zero exit code $?"

exit 0
