#!/bin/bash
set -e

make go_build_dev
./web &
SERVER_PID=$!
# node_modules is already served live by embed;
# components is served by nicehttp from disk.
while inotifywait -q -e modify -r . @.git @results-processor @webapp/node_modules @webapp/components; do
  # It's fine if the server has already died.
  kill $SERVER_PID || true
  # If we fail to build (quite likely as we are editing files), try again.
  make go_build_dev || continue
  ./web &
  SERVER_PID=$!
done
