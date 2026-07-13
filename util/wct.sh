#!/bin/bash
# wct.sh [true|false]
# Run web component tests with or without Xvfb.

set -ex

USE_FRAME_BUFFER=$1

function stop_xvfb() {
  if [ "$USE_FRAME_BUFFER" == "true" ]; then
    # It's fine if Xvfb has already exited.
    killall Xvfb || true
  fi
}

trap stop_xvfb EXIT SIGINT SIGTERM

if [ "$USE_FRAME_BUFFER" == "true" ]; then
  export DISPLAY=:99
  (Xvfb :99 -screen 0 1024x768x24 -ac +extension GLX +render -noreset &)
fi

cd webapp
# Patch browser test harness to avoid cross-origin error in ChildRunner.get.
# See https://github.com/web-platform-tests/wpt.fyi/issues/4754
# This workaround can be phased out if/when wct-browser-legacy is updated or replaced.
for file in node_modules/wct-browser-legacy/browser.js node_modules/web-component-tester/browser.js; do
  if [ -f "$file" ]; then
    sed -i 's/return window.parent.WCT._ChildRunner.get(target, true);/try { return window.parent.WCT._ChildRunner.get(target, true); } catch (e) { return null; }/' "$file"
  fi
done
if [ "$UID" == "0" ]; then
  # Used in .github/actions/make-in-docker/Dockerfile
  sudo -u browser npm test
else
  npm test
fi
