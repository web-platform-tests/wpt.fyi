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
# Patch wct-browser-legacy to avoid cross-origin errors.
# See https://github.com/web-platform-tests/wpt.fyi/issues/4754 and #4956
# This workaround can be phased out if/when wct-browser-legacy is updated or replaced.
sed -i 's/return window.parent.WCT._ChildRunner.get(target, true);/try { return window.parent.WCT._ChildRunner.get(target, true); } catch (e) { return null; }/' node_modules/wct-browser-legacy/browser.js
sed -i 's/if (this.iframe.contentWindow.WCT) {/var hasWCT = false; try { hasWCT = !!this.iframe.contentWindow.WCT; } catch (e) {} if (hasWCT) {/' node_modules/wct-browser-legacy/browser.js
sed -i 's/return target.removeEventListener(type, listener);/try { return target.removeEventListener(type, listener); } catch (e) {}/' node_modules/wct-browser-legacy/browser.js
if [ "$UID" == "0" ]; then
  # Used in .github/actions/make-in-docker/Dockerfile
  sudo -u browser npm test
else
  npm test
fi
