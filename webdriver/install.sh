#!/usr/bin/env bash

# Fetch needed webdriver dependencies.

SCRIPT_DIR=$(dirname "$0")
source "${SCRIPT_DIR}/../util/logging.sh"
source "${SCRIPT_DIR}/../util/path.sh"

set -e

usage() {
  USAGE="Usage: install.sh [-r] [-p PATH]
    -r   - Reinstall
    -p   - Path to install (default: ~/browsers)"
  info "${USAGE}"
}

INSTALL_DIR=~/browsers
REINSTALL="false"
while getopts ':rp:' flag; do
  case "${flag}" in
    r) REINSTALL='true' ;;
    p) INSTALL_DIR="${OPTARG}" ;;
    h|*) usage && exit 0;;
  esac
done

# fetch [url] [filename]
# Downloads [url] if [filename] doesn't already exist.
function fetch () {
    if [[ -e $2 ]]
    then
        info "$2 already present."
    else
        debug "wget -q $1"
        wget -q "$1"
    fi
}

info "Changing into ${INSTALL_DIR}..."
if [[ ! -e ${INSTALL_DIR} ]];
then
  mkdir ${INSTALL_DIR}
fi
cd ${INSTALL_DIR}

# Firefox 60
FIREFOX="firefox"
case "${UNAME_OUT}" in
    Darwin*)
        FIREFOX_OS="mac"
        FIREFOX_DMG="Firefox 60.0.dmg"
        FIREFOX_SRC="${FIREFOX_DMG}"
        ;;
    Linux*|*)
        FIREFOX_OS="linux-x86_64"
        FIREFOX_TBZ="${FIREFOX}-60.0.tar.bz2"
        FIREFOX_SRC="${FIREFOX_TBZ}"
        ;;
esac
FIREFOX_URL="https://releases.mozilla.org/pub/firefox/releases/60.0/${FIREFOX_OS}/en-US/${FIREFOX_SRC}"

info "Getting ${FIREFOX} binary..."
if [[ ! -e ${FIREFOX} || "${REINSTALL}" == "true" ]]
then
    if [[ -e ${FIREFOX} && "${REINSTALL}" == "true" ]]
    then
        warn "Removing existing ${FIREFOX} dir..."
        rm -r ${FIREFOX}
    fi

    info "Downloading ${FIREFOX_URL}..."
    fetch "${FIREFOX_URL}" "${FIREFOX_SRC}"

    if [[ "${FIREFOX_DMG}" != "" ]]
    then
        hdiutil attach "${FIREFOX_DMG}"
        FIREFOX_CP_CMD="cp -R /Volumes/Firefox/Firefox.app ${FIREFOX}"
        debug "${FIREFOX_CP_CMD}"
        ${FIREFOX_CP_CMD}
        hdiutil detach "/Volumes/Firefox"
    elif [[ "${FIREFOX_TBZ}" != "" ]]
    then
        debug "Unzipping ${FIREFOX_TBZ}..."
        if [[ ! -e firefox || "${REINSTALL}" == "true" ]]; then tar -xjf ${FIREFOX_TBZ}; fi
    fi
fi
