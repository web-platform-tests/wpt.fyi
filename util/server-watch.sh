#!/bin/bash
set -e

./web &
SERVER_PID=$!
# node_modules is already served live by packr;
# components is served by nicehttp from disk.
while inotifywait -q -e modify -r webapp @webapp/node_modules @webapp/components; do
  kill $SERVER_PID
  make go_build_dev
  ./web &
  SERVER_PID=$!
done
