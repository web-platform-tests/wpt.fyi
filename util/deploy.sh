#!/bin/bash

# Helper script for using a standardized version flag when deploying.

REPO_DIR="$(dirname "${BASH_SOURCE[0]}")/.."
source "${REPO_DIR}/util/logging.sh"
source "${REPO_DIR}/util/path.sh"
WPTD_PATH=${WPTD_PATH:-$(absdir ${REPO_DIR})}

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

APP_PATH=${@: -1}
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
  "webapp" | \
  "results-processor" | \
  "revisions/service" | \
  "api/query/cache/service" | \
  "api/query/cache/service/app.staging.yaml")
  ;;
*)
  fatal "Unrecognized app path \"${APP_PATH}\"."
  ;;
esac

# Ensure dependencies are installed.
if [[ -z "${QUIET}" ]]; then info "Installing dependencies..."; fi
cd ${WPTD_PATH}
if [[ "${APP_PATH}" == "webapp" ]]; then
  make webapp_deps || fatal "Error installing deps"
  make webapp_node_modules_prune || fatal "Error pruning node_modules"
fi

# Create a name for this version
VERSION_BRANCH_NAME="$(echo ${BRANCH_NAME:-"$(git rev-parse --abbrev-ref HEAD)"} | tr /_ - | cut -c 1-28)"
USER="$(git remote -v get-url origin | sed -E 's#(https?:\/\/|git@)github.com(\/|:)##' | sed 's#/.*$##')-"
if [[ "${USER}" == "web-platform-tests-" ]]; then USER=""; fi

VERSION="${USER}${VERSION_BRANCH_NAME}"
if [[ -n ${RELEASE} ]]
then
  # Use SHA for releases.
  VERSION="$(git rev-parse --short HEAD)"
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

if [[ -z "${QUIET}" ]]
then
    info "Deploy command:\n${COMMAND}"
    confirm "Execute?"
    if [[ "${?}" != "0" ]]; then fatal "User cancelled the deploy"; fi
fi

set -e

if [[ -z "${QUIET}" ]]; then info "Executing..."; fi
${COMMAND} || fatal "Deploy returned non-zero exit code $?"

exit 0
