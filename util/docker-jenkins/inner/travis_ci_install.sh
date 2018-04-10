#!/bin/bash

# This file installs for WPT testrun in a container
# in the context of Travis CI.

set -x

pushd "${WPT_PATH}/.."
git clone --depth 1 https://github.com/w3c/web-platform-tests
popd

source "${WPT_PATH}/tools/ci/lib.sh"
hosts_fixup
