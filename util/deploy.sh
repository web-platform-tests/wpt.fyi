#!/bin/bash

# Helper script for using a standardized version flag when deploying.

set -e

REPO_DIR="$(dirname "$0")/.."
source "${REPO_DIR}/util/logging.sh"
source "${REPO_DIR}/util/path.sh"
WPTD_PATH=${WPTD_PATH:-$(absdir ${REPO_DIR})}

usage() {
  USAGE="Usage: deploy.sh [-p] [-q] [-b] [-r] [-i] [-d] [-h]
    -p Production deploy
    -q Quiet / no user prompts
    -b Branch name - defaults to current Git branch
    -r Repo slug (e.g. web-platform-tests/wpt.fyi), for making a Github comment
    -i Issue (PR) number, for making a Github comment
    -g Github token, for making a Github comment
    -d Print debug info
    -h Show (this) help information"
  echo "${USAGE}"
}

while getopts ':b:dphqr:i:g:' flag; do
  case "${flag}" in
    b) BRANCH_NAME="${OPTARG}" ;;
    p) PRODUCTION='true' ;;
    q) QUIET='true' ;;
    r) REPO_SLUG="${OPTARG}" ;;
    i) PULL_REQUEST="${OPTARG}" ;;
    g) GITHUB_TOKEN="${OPTARG}" ;;
    d) DEBUG='true' ;;
    h|*) usage && exit 0;;
  esac
done

# Ensure dependencies are installed.
if [[ -n ${DEBUG} ]]; then info "Installing dependencies..."; fi
cd ${WPTD_PATH}; make go_deps;

# Create a name for this version
BRANCH_NAME=${BRANCH_NAME:-"$(git rev-parse --abbrev-ref HEAD)"}
USER="$(git remote -v get-url origin | sed -E 's#(https?:\/\/|git@)github.com(\/|:)##' | sed 's#/.*$##')-"
if [[ "${USER}" == "web-platform-tests-" ]]; then USER=""; fi

VERSION="${USER}${BRANCH_NAME}"
PROMOTE="--no-promote"

if [[ -n ${PRODUCTION} ]]
then
  if [[ -n ${DEBUG} ]]; then info "Producing production configuration..."; fi
  if [[ "${USER}" != "web-platform-tests" ]]
  then
    if [[ -z ${QUIET} ]]
    then
      confirm "Are you sure you want to be deploying a non-web-platform-tests repo (${USER})?"
      if [ "${?}" != "0" ]; then exit "${?}"; fi
    fi
  fi
  # Use SHA for prod-pushes.
  VERSION="$(git rev-parse --short HEAD)"
  PROMOTE="--promote"
fi

if [[ -n ${QUIET} ]]
then
    QUIET_FLAG="-q"
else
    QUIET_FLAG=""
fi
COMMAND="gcloud app deploy ${PROMOTE} ${QUIET_FLAG} --version=${VERSION} ${WPTD_PATH}/webapp"

if [[ -n ${DEBUG} ]]; then info "Deploy command:\n${COMMAND}"; fi
if [[ -z ${QUIET} ]]
then
    confirm "Execute?"
    if [ "${?}" != "0" ]; then exit "${?}"; fi
fi

if [[ -n ${DEBUG} ]]; then info "Executing..."; fi
${COMMAND}

# Comment on the PR if running from Travis.
DEPLOYED_URL=$(gcloud app versions describe ${VERSION} -s default | grep -Po 'versionUrl: \K.*$')
echo "Deployed to ${DEPLOYED_URL}"

exit 0
