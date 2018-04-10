#!/bin/bash

# Helper script for using a standardized version flag when deploying.

REPO_DIR="$(dirname "$0")/.."
source "${REPO_DIR}/util/logging.sh"
source "${REPO_DIR}/util/path.sh"
WPTD_PATH=${WPTD_PATH:-$(absdir ${REPO_DIR})}

usage() {
  info "Usage: deploy.sh [-p] [-h]";
}

while getopts ':ph' flag; do
  case "${flag}" in
    p) PRODUCTION='true' ;;
    h|*) usage && exit 0;;
  esac
done

# Ensure dependencies are installed.
info "Installing dependencies..."
cd ${WPTD_PATH}; make go_deps;

# Create a name for this version
BRANCH_NAME="$(git rev-parse --abbrev-ref HEAD)"
USER="$(git remote -v get-url origin | sed -E 's#(https?:\/\/|git@)github.com(\/|:)##' | sed 's#/.*$##')"
VERSION="${USER}-${BRANCH_NAME}"
PROMOTE="--no-promote"

if [[ ${PRODUCTION} == 'true' ]]
then
  info "Producing production configuration..."
  if [[ "${USER}" != "w3c" ]]
  then
    confirm "Are you sure you want to be deploying a non-w3c repo (${USER})?"
    if [ "${?}" != "0" ]; then exit "${?}"; fi
  fi
  # Use SHA for prod-pushes.
  VERSION="$(git rev-parse --short HEAD)"
  PROMOTE="--promote"
fi

COMMAND="gcloud app deploy ${PROMOTE} --version=${VERSION} ${WPTD_PATH}/webapp"

info "Deploy command:\n${COMMAND}"
confirm "Execute?"
if [ "${?}" != "0" ]; then exit "${?}"; fi

info "Executing..."
${COMMAND}
exit 0
