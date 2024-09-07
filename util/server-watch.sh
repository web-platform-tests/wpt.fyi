#!/bin/bash
set -e

usage() {
  USAGE="Usage: server-watch.sh [-d]
    -d : Start a debugging session with Delve"
  echo "${USAGE}"
}

while getopts ':dh' flag; do
  case "${flag}" in
    d) DEBUG='true' ;;
    h|*) usage && exit 0;;
  esac
done
UTIL_DIR=$(dirname $0)
source "${UTIL_DIR}/logging.sh"

make go_build_dev
if [[ ${DEBUG} != "true" ]];
then
  ./web &
  SERVER_PID=$!
else
  if [[ $(which dlv) == "" ]]; then \
    go install github.com/go-delve/delve/cmd/dlv@latest
  fi
  info "Starting debugger on port 12345"
  dlv debug github.com/web-platform-tests/wpt.fyi/webapp/web --headless --listen=:12345 --output /tmp/web.debug
  exit 0
fi

# node_modules is already served live by packr;
# components is served by nicehttp from disk.
while inotifywait -q -e modify -r . @.git @results-processor @webapp/node_modules @webapp/components; do
  # It's fine if the server has already died.
  kill $SERVER_PID || true
  # If we fail to build (quite likely as we are editing files), try again.
  make go_build_dev || continue
  ./web &
  SERVER_PID=$!
done
