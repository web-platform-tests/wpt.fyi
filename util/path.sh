#!/bin/bash

function absdir() {
  pushd "${1}" > /dev/null
  if [ "${?}" != "0" ]; then
    echo "${1}"
    return 1
  fi
  pwd
  popd > /dev/null
}
