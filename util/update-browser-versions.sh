#!/bin/bash
# Copyright 2026 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

# Updates the browser versions that the Makefile installs for the tests:
#   CHROME_VERSION   to the latest Chrome for Testing Stable release,
#   FIREFOX_VERSION  to the latest Firefox release, and
#   GECKODRIVER_TAG  to the latest geckodriver release.
#
# Usage: util/update-browser-versions.sh
# Prints what changed. When run in GitHub Actions, it also writes `changed`
# and `<browser>_old`/`<browser>_new` for chrome, firefox and geckodriver to
# $GITHUB_OUTPUT. Set GITHUB_TOKEN to avoid GitHub API rate limits.

set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."
MAKEFILE=Makefile

CFT_VERSIONS_URL=https://googlechromelabs.github.io/chrome-for-testing/last-known-good-versions.json
FIREFOX_VERSIONS_URL=https://product-details.mozilla.org/1.0/firefox_versions.json
GECKODRIVER_RELEASE_URL=https://api.github.com/repos/mozilla/geckodriver/releases/latest

current() {
  sed -n "s/^$1 := //p" "${MAKEFILE}"
}

chrome_old=$(current CHROME_VERSION)
firefox_old=$(current FIREFOX_VERSION)
geckodriver_old=$(current GECKODRIVER_TAG)
if [[ -z "${chrome_old}" || -z "${firefox_old}" || -z "${geckodriver_old}" ]]; then
  echo "CHROME_VERSION, FIREFOX_VERSION or GECKODRIVER_TAG not found in ${MAKEFILE}" >&2
  exit 1
fi

chrome_new=$(curl -fsSL "${CFT_VERSIONS_URL}" | jq -r '.channels.Stable.version')
firefox_new=$(curl -fsSL "${FIREFOX_VERSIONS_URL}" | jq -r '.LATEST_FIREFOX_VERSION')
auth=()
if [[ -n "${GITHUB_TOKEN:-}" ]]; then
  auth=(-H "Authorization: Bearer ${GITHUB_TOKEN}")
fi
geckodriver_new=$(curl -fsSL "${auth[@]}" "${GECKODRIVER_RELEASE_URL}" | jq -r '.tag_name')

# Only accept well-formed versions, so a bad response can't break the Makefile.
if [[ ! "${chrome_new}" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Unexpected Chrome version: '${chrome_new}'" >&2
  exit 1
fi
if [[ ! "${firefox_new}" =~ ^[0-9]+\.[0-9]+(\.[0-9]+)?$ ]]; then
  echo "Unexpected Firefox version: '${firefox_new}'" >&2
  exit 1
fi
if [[ ! "${geckodriver_new}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Unexpected geckodriver tag: '${geckodriver_new}'" >&2
  exit 1
fi

# Make sure the release can actually be downloaded before switching to it.
curl -fsSI -o /dev/null \
  "https://storage.googleapis.com/chrome-for-testing-public/${chrome_new}/linux64/chrome-linux64.zip"
curl -fsSI -o /dev/null \
  "https://archive.mozilla.org/pub/firefox/releases/${firefox_new}/linux-x86_64/en-US/firefox-${firefox_new}.tar.xz"
curl -fsSIL -o /dev/null \
  "https://github.com/mozilla/geckodriver/releases/download/${geckodriver_new}/geckodriver-${geckodriver_new}-linux64.tar.gz"

sed -i \
  -e "s/^CHROME_VERSION := .*/CHROME_VERSION := ${chrome_new}/" \
  -e "s/^FIREFOX_VERSION := .*/FIREFOX_VERSION := ${firefox_new}/" \
  -e "s/^GECKODRIVER_TAG := .*/GECKODRIVER_TAG := ${geckodriver_new}/" \
  "${MAKEFILE}"

changed=false
if [[ "${chrome_old}" != "${chrome_new}" || "${firefox_old}" != "${firefox_new}" ||
      "${geckodriver_old}" != "${geckodriver_new}" ]]; then
  changed=true
fi
echo "Chrome:      ${chrome_old} -> ${chrome_new}"
echo "Firefox:     ${firefox_old} -> ${firefox_new}"
echo "geckodriver: ${geckodriver_old} -> ${geckodriver_new}"
echo "changed:     ${changed}"

if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  {
    echo "changed=${changed}"
    echo "chrome_old=${chrome_old}"
    echo "chrome_new=${chrome_new}"
    echo "firefox_old=${firefox_old}"
    echo "firefox_new=${firefox_new}"
    echo "geckodriver_old=${geckodriver_old}"
    echo "geckodriver_new=${geckodriver_new}"
  } >> "${GITHUB_OUTPUT}"
fi
