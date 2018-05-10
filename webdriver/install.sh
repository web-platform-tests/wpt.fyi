#!/usr/bin/env bash

# Fetch needed webdriver dependencies.

SCRIPT_DIR=$(dirname "$0")
source "${SCRIPT_DIR}/../util/logging.sh"
source "${SCRIPT_DIR}/../util/path.sh"

set -e

usage() {
  USAGE="Usage: install.sh [-r] [path]
    -r   - Reinstall
    path - Path to install (default: ~/browsers)"
  info ${USAGE}
}

INSTALL_DIR=${1:-~/browsers}

REINSTALL="false"
while getopts ':r' flag; do
  case "${flag}" in
    r) REINSTALL='true' ;;
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
        debug "wget $1"
        wget "$1"
    fi
}

info "Changing into ${INSTALL_DIR}..."
if [[ ! -e ${INSTALL_DIR} ]];
then
  mkdir ${INSTALL_DIR}
fi
cd ${INSTALL_DIR}

# Selenium standalone.
SELENIUM_STANDALONE="selenium-server-standalone-3.8.1.jar"
SELENIUM_STANDALONE_URL="http://selenium-release.storage.googleapis.com/3.8/${SELENIUM_STANDALONE}"

info "Getting ${SELENIUM_STANDALONE} binary..."
fetch "${SELENIUM_STANDALONE_URL}" "${SELENIUM_STANDALONE}"

# Gecko driver
GECKO_DRIVER="geckodriver-v0.19.1"
UNAME_OUT="$(uname -s)"
case "${UNAME_OUT}" in
    Darwin*)   GECKO_DRIVER_OS="macos";;
    Linux*|*)  GECKO_DRIVER_OS="linux64";;
esac
GECKO_DRIVER_TAR="${GECKO_DRIVER}-${GECKO_DRIVER_OS}.tar"
GECKO_DRIVER_GZ="${GECKO_DRIVER_TAR}.gz"
GECKO_DRIVER_URL="https://github.com/mozilla/geckodriver/releases/download/v0.19.1/${GECKO_DRIVER_GZ}"

info "Getting ${GECKO_DRIVER} binary..."
if [[ ! -e ${GECKO_DRIVER} || "${REINSTALL}" == "true" ]]
then
    info "Downloading ${GECKO_DRIVER_URL}..."
    fetch "${GECKO_DRIVER_URL}" "${GECKO_DRIVER_GZ}"

    debug "Unzipping ${GECKO_DRIVER_GZ}..."
    if [[ ! -e ${GECKO_DRIVER_TAR} ]]; then gunzip -k ${GECKO_DRIVER_GZ}; fi

    debug "Unzipping ${GECKO_DRIVER_TAR}..."
    if [[ ! -e geckodriver || "${REINSTALL}" == "true" ]]; then tar -xf ${GECKO_DRIVER_TAR}; fi

    if [[ -e "${GECKO_DRIVER}" && "${REINSTALL}" == "true" ]]
    then
        info "Removing existing ${GECKO_DRIVER}..."
        rm ${GECKO_DRIVER}
    fi

    debug "Renaming to ${GECKO_DRIVER}..."
    mv geckodriver ${GECKO_DRIVER}
fi

# Firefox 58
FIREFOX="firefox-58.0"
case "${UNAME_OUT}" in
    Darwin*)
        FIREFOX_OS="mac"
        FIREFOX_DMG="Firefox 58.0.dmg"
        FIREFOX_SRC="${FIREFOX_DMG}"
        ;;
    Linux*|*)
        FIREFOX_OS="linux-x86_64"
        FIREFOX_TAR="${FIREFOX}.tar"
        FIREFOX_BZ="${FIREFOX_TAR}.bz2"
        FIREFOX_SRC="${FIREFOX_BZ}"
        ;;
esac
FIREFOX_URL="https://releases.mozilla.org/pub/firefox/releases/58.0/${FIREFOX_OS}/en-US/${FIREFOX_SRC}"

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
    elif [[ "${FIREFOX_BZ}" != "" ]]
    then
        debug "Unzipping ${FIREFOX_BZ}..."
        if [[ ! -e ${FIREFOX_TAR} || "${REINSTALL}" == "true" ]]; then bzip2 -dkf ${FIREFOX_BZ}; fi

        debug "Unzipping ${FIREFOX_TAR}..."
        if [[ ! -e firefox || "${REINSTALL}" == "true" ]]; then tar -xf ${FIREFOX_TAR}; fi

        debug "Renaming to ${FIREFOX}..."
        mv firefox ${FIREFOX}
    fi
fi
