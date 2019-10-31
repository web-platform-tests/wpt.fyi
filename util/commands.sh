#!/bin/bash

DOCKER_INSTANCE="${DOCKER_INSTANCE:-wptd-dev-instance}"

function wptd_chown() {
  docker exec -u 0:0 "${DOCKER_INSTANCE}" chown -R $(id -u $USER):$(id -g $USER) $1
}
function wptd_useradd() {
  # Allow the exit code of groupadd to be 4 (GID not unique).
  docker exec -u 0:0 "${DOCKER_INSTANCE}" groupadd -g $(id -g $USER) user || [ $? == 4 ]
  # Add user to audio & video groups to ensure Chrome can use sandbox.
  docker exec -u 0:0 "${DOCKER_INSTANCE}" useradd -u $(id -u $USER) -g $(id -g $USER) -G audio,video user
  docker exec -u 0:0 "${DOCKER_INSTANCE}" sh -c 'echo "user ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers'
}
function wptd_exec() {
  docker exec -u $(id -u $USER) "${DOCKER_INSTANCE}" sh -c "$*"
}
function wptd_exec_it() {
  docker exec -it -u $(id -u $USER) "${DOCKER_INSTANCE}" sh -c "$*"
}
# function wptd_run() {}
function wptd_stop() {
  docker stop "${DOCKER_INSTANCE}"
}
function wptd_rm() {
  docker rm "${DOCKER_INSTANCE}"
}
