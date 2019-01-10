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
npm test
