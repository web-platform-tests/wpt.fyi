#!/bin/bash

DOCKER_INSTANCE="${DOCKER_INSTANCE:-wptd-dev-instance}"

function wptd_chown() {
  docker exec -u 0:0 "${DOCKER_INSTANCE}" chown -R $(id -u $USER):$(id -g $USER) $1
}
function wptd_useradd() {
  docker exec -u 0:0 "${DOCKER_INSTANCE}" groupadd -g "$(id -g $USER)" user
  docker exec -u 0:0 "${DOCKER_INSTANCE}" useradd -u "$(id -u $USER)" -g "$(id -g $USER)" user
  docker exec -u 0:0 "${DOCKER_INSTANCE}" usermod -a -G user root
  docker exec -u 0:0 "${DOCKER_INSTANCE}" sh -c 'echo "%user ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers'
}
function wptd_exec() {
  docker exec -u $(id -u $USER):$(id -g $USER) "${DOCKER_INSTANCE}" "$@"
}
function wptd_exec_it() {
  docker exec -it -u $(id -u $USER):$(id -g $USER) "${DOCKER_INSTANCE}" "$@"
}
# function wptd_run() {}
function wptd_stop() {
  docker stop "${DOCKER_INSTANCE}"
}
function wptd_rm() {
  docker rm "${DOCKER_INSTANCE}"
}
