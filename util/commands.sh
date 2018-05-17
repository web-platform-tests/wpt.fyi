#!/bin/bash

function wptd_chown() {
  docker exec -u 0:0 wptd-dev-instance chown -R $(id -u $USER):$(id -g $USER) $1
}
function wptd_useradd() {
  docker exec -u 0:0 wptd-dev-instance groupadd -g "$(id -g $USER)" user
  docker exec -u 0:0 wptd-dev-instance useradd -u "$(id -u $USER)" -g "$(id -g $USER)" user
  docker exec -u 0:0 wptd-dev-instance sh -c 'echo "%user ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers'
}
function wptd_exec() {
  docker exec -u $(id -u $USER):$(id -g $USER) wptd-dev-instance $1
}
function wptd_exec_it() {
  docker exec -it -u $(id -u $USER):$(id -g $USER) wptd-dev-instance $1
}
# function wptd_run() {}
function wptd_stop() {
  docker stop wptd-dev-instance
}
function wptd_rm() {
  docker rm wptd-dev-instance
}
