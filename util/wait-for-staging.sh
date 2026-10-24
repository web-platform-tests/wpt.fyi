#!/bin/bash
# Copyright 2026 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

# Waits until the staging webapp consistently serves the app, so that the
# WebDriver tests that run against it right after a deploy don't hit it while
# App Engine is still moving traffic to the new version.
#
# Each round loads the results and runs pages (which must contain their app
# element and the import map) and two modules they load, all served by the Go
# app. The script succeeds once every round has passed for WAIT_STABLE seconds
# in a row; a failed round starts the period over. It fails if that hasn't
# happened within WAIT_TIMEOUT seconds.
#
# Usage: util/wait-for-staging.sh [base URL]
#   base URL       defaults to https://staging.wpt.fyi
#   WAIT_STABLE    seconds in a row that every round must pass (default 240)
#   WAIT_INTERVAL  seconds between rounds (default 3)
#   WAIT_TIMEOUT   seconds before giving up (default 600)

set -u

UTIL_DIR="$(dirname "${BASH_SOURCE[0]}")"
source "${UTIL_DIR}/logging.sh"

BASE_URL="${1:-https://staging.wpt.fyi}"
WAIT_STABLE="${WAIT_STABLE:-240}"
WAIT_INTERVAL="${WAIT_INTERVAL:-3}"
WAIT_TIMEOUT="${WAIT_TIMEOUT:-600}"

# Prints the body of a URL, or fails (non-2xx status, network error, timeout).
function fetch() {
  curl --silent --show-error --fail --max-time 30 \
    --header 'Cache-Control: no-cache' "${BASE_URL}$1"
}

# Checks that a page is served and contains the given element and the import
# map that resolves Lit. Prints the reason when it doesn't.
function check_page() {
  local path="$1" element="$2" body
  if ! body="$(fetch "${path}" 2>&1)"; then
    echo "${path}: ${body}"
    return 1
  fi
  if [[ "${body}" != *"<${element}"* ]]; then
    echo "${path}: the page has no <${element}>"
    return 1
  fi
  if [[ "${body}" != *'type="importmap"'* ]]; then
    echo "${path}: the page has no import map"
    return 1
  fi
}

# Checks that a module is served. Prints the reason when it isn't.
function check_module() {
  local path="$1" output
  if ! output="$(fetch "${path}" 2>&1 >/dev/null)"; then
    echo "${path}: ${output}"
    return 1
  fi
}

function check_round() {
  check_page "/results/?label=master" "wpt-app" &&
    check_page "/runs" "wpt-runs" &&
    check_module "/node_modules/lit/index.js" &&
    check_module "/node_modules/@polymer/polymer/polymer-element.js"
}

info "Waiting for ${BASE_URL} to pass every check for ${WAIT_STABLE}s in a row (timeout ${WAIT_TIMEOUT}s)..."
START=${SECONDS}
STABLE_SINCE=""
FAILED=0
while true; do
  if (( SECONDS - START >= WAIT_TIMEOUT )); then
    fatal "${BASE_URL} didn't pass every check for ${WAIT_STABLE}s in a row within ${WAIT_TIMEOUT}s (${FAILED} failed checks)."
  fi
  if reason="$(check_round)"; then
    STABLE_SINCE="${STABLE_SINCE:-${SECONDS}}"
    if (( SECONDS - STABLE_SINCE >= WAIT_STABLE )); then
      break
    fi
  else
    FAILED=$((FAILED + 1))
    warn "Check failed: ${reason}"
    STABLE_SINCE=""
  fi
  sleep "${WAIT_INTERVAL}"
done
info "${BASE_URL} passed every check for ${WAIT_STABLE}s in a row after $((SECONDS - START))s (${FAILED} failed checks)."
