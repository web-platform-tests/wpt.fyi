# Copyright 2017 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

# Make targets in this file are intended to be run inside the Docker container
# environment.

# Make targets can be run in a host environment, but that requires ensuring
# the correct version of tools are installed and environment variables are
# set appropriately.

# Prefer simply expanded variables (:=) to avoid confusion caused by recursion.
# All variables can be overridden in command line by `make target FOO=bar`.

SHELL := /bin/bash
# WPTD_PATH will have a trailing slash, e.g. /home/user/wpt.fyi/
WPTD_PATH := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
NODE_SELENIUM_PATH := $(WPTD_PATH)webapp/node_modules/selenium-standalone/.selenium/
FIREFOX_PATH := /usr/bin/firefox
CHROME_PATH := /usr/bin/google-chrome
CHROMEDRIVER_PATH=/usr/bin/chromedriver
USE_FRAME_BUFFER := true
STAGING := false
VERBOSE := -v

GO_FILES := $(shell find $(WPTD_PATH) -type f -name '*.go')
GO_TEST_FILES := $(shell find $(WPTD_PATH) -type f -name '*_test.go')
# Golangci version should be updated periodically.
# See: https://golangci-lint.run/usage/install/#other-ci
GOLANGCI_LINT_VERSION := v1.60.3

build: go_build

test: go_test python_test

lint: eslint go_lint golangci_lint # TODO: Replace go_lint with golangci_lint

prepush: VERBOSE := $() # Empty out the verbose flag.
prepush: go_build go_test lint

python_test: python3 tox
	tox -c results-processor/

# Contains setup necessary only for github actions.
github_action_go_setup:
	# https://github.com/web-platform-tests/wpt.fyi/issues/3089
	if [ -d "/github/workspace" ]; then \
		echo "Avoiding buildvcs error for Go 1.18+ by marking github workspace safe."; \
		git config --global --add safe.directory /github/workspace ; \
	else \
		echo "Did not detect github workspace. Skipping." ; \
	fi
# NOTE: We prune before generate, because node_modules are embedded into the
# binary (and part of the build).
go_build: git mockgen github_action_go_setup webapp_node_modules_prod
	go generate ./...
	# Check all packages without producing any output.
	go build -v ./...
	# Build the webapp.
	go build -v ./webapp/web

go_build_dev:
	@ # Disable packr to always serve local node modules and dynamic components.
	@ # There's thus no need to prune node_modules.
	@ # Disable inlining and optimizations that can interfere with debugging.
	go build -v -tags skippackr -gcflags=all="-N -l" ./webapp/web

go_lint: golint go_test_tag_lint
	golint -set_exit_status ./api/...
	golint -set_exit_status ./shared/...
	golint -set_exit_status ./util/...
	golint -set_exit_status ./webapp/...
	golint -set_exit_status ./webdriver/...

# TODO: run on /shared/, /util/, /webapp/, /webdriver/
golangci_lint: golangci-lint github_action_go_setup
	golangci-lint cache clean
	golangci-lint run ./api/...

go_test_tag_lint:
	@ # Printing a list of test files without +build tag, asserting empty...
	@TAGLESS=$$(grep -PL '\/\/(\s?\+build|go:build) !?(small|medium|large|cloud)' $(GO_TEST_FILES)); \
	if [ -n "$$TAGLESS" ]; then echo -e "Files are missing '// +build TAG' or '//go:build TAG' tags:\n$$TAGLESS" && exit 1; fi

go_test: go_small_test go_medium_test

go_small_test: go_build gcc
	go test -tags=small $(VERBOSE) ./...

go_medium_test: go_build dev_appserver_deps gcc
	go test -tags=medium $(VERBOSE) $(FLAGS) ./...

# Use sub-make because otherwise make would only execute the first invocation
# of _go_webdriver_test. Variables will be passed into sub-make implicitly.
go_large_test:
	make go_firefox_test
	make go_chrome_test

go_firefox_test: firefox geckodriver
	make _go_webdriver_test BROWSER=firefox

go_chrome_test: chrome chromedriver
	make _go_webdriver_test BROWSER=chrome

go_cloud_test: go_build gcloud_login
	gcloud config set project wptdashboard-staging; \
	if [[ -f "$(WPTD_PATH)client-secret.json" ]]; then \
		echo "Running with client-secret.json credentials instead of possible system credentials. This should happen for CI runs."; \
		export GOOGLE_APPLICATION_CREDENTIALS="$(WPTD_PATH)client-secret.json"; \
	fi ; \
	GOOGLE_CLOUD_PROJECT=wptdashboard-staging GAE_SERVICE=test GAE_VERSION=1 go test -tags=cloud $(VERBOSE) $(FLAGS) ./...

puppeteer_chrome_test: go_build dev_appserver_deps webdriver_node_deps
	cd webdriver; npm test

webdriver_node_deps:
	cd webdriver; npm install

# _go_webdriver_test is not intended to be used directly; use go_firefox_test or
# go_chrome_test instead.
_go_webdriver_test: var-BROWSER java go_build xvfb geckodriver dev_appserver_deps gcc
	@ # This Go test manages Xvfb itself, so we don't start/stop Xvfb for it.
	@ # The following variables are defined here because we don't know the
	@ # path before installing geckodriver as it includes version strings.
	GECKODRIVER_PATH="$(shell find $(NODE_SELENIUM_PATH)geckodriver/ -type f -name '*geckodriver')"; \
	COMMAND="go test $(VERBOSE) -timeout=15m -tags=large ./webdriver -args \
		-firefox_path=$(FIREFOX_PATH) \
		-geckodriver_path=$$GECKODRIVER_PATH \
		-chrome_path=$(CHROME_PATH) \
		-chromedriver_path=$(CHROMEDRIVER_PATH) \
		-frame_buffer=$(USE_FRAME_BUFFER) \
		-staging=$(STAGING) \
		-browser=$(BROWSER) $(FLAGS)"; \
	if [ "$$UID" == "0" ]; then sudo -u browser $$COMMAND; else $$COMMAND; fi

# NOTE: psmisc includes killall, needed by wct.sh
web_components_test: xvfb firefox chrome webapp_node_modules_all psmisc
	util/wct.sh $(USE_FRAME_BUFFER)

dev_appserver_deps: gcloud-app-engine-go gcloud-cloud-datastore-emulator gcloud-beta java

# Note: If we change to downloading chrome from Chrome For Testing, modify the
# `chromedriver` target below to use the `known-good-versions-with-downloads.json` endpoint.
# More details can be found in the comment for the `chromedriver` target.
# TODO: pinning Chrome to 130 due to https://github.com/web-platform-tests/wpt.fyi/issues/4129
chrome: wget
	if [[ -z "$$(which google-chrome)" ]]; then \
		ARCHIVE=google-chrome-stable_130.0.6723.116-1_amd64.deb; \
		wget -q https://dl.google.com/linux/chrome/deb/pool/main/g/google-chrome-stable/$${ARCHIVE}; \
		sudo apt-get update; \
		sudo dpkg --install $${ARCHIVE} 2>/dev/null || true; \
		sudo apt-get install --fix-broken --fix-missing -qqy; \
		sudo dpkg --install $${ARCHIVE} 2>/dev/null; \
	fi

# Pull ChromeDriver from Chrome For Testing (CfT)
# Need to create the CHROMEDRIVER_PATH and then move the files in because the
# directory structure in chromedriver_linux64.zip has changed.
#
# CfT only has ChromeDriver URLs for chrome versions >=115. But assuming `chrome`
# target above remains pulling the latest stable, this will not be a problem.
#
# Until we also pull chrome from CfT, we should use the latest-patch-versions-per-build-with-downloads.json.
# When we make the switch, we can download from the known-good-versions-with-downloads.json endpoint too.
# More details: https://github.com/web-platform-tests/wpt.fyi/pull/3433/files#r1282787489
chromedriver: wget unzip chrome jq
	if [[ ! -f "$(CHROMEDRIVER_PATH)" ]]; then \
		CHROME_VERSION=$$(google-chrome --version | grep -ioE "[0-9]+\.[0-9]+\.[0-9]+"); \
		CHROMEDRIVER_URL=$$(curl -s https://googlechromelabs.github.io/chrome-for-testing/latest-patch-versions-per-build-with-downloads.json | jq -r ".builds[\"$${CHROME_VERSION}\"].downloads.chromedriver[] | select(.platform == \"linux64\") | .url"); \
		TEMP_DIR=$$(mktemp -d); \
		wget -q -O $${TEMP_DIR}/chromedriver_linux64.zip $${CHROMEDRIVER_URL}; \
		unzip -j $${TEMP_DIR}/chromedriver_linux64.zip -d $${TEMP_DIR}; \
		sudo mv $${TEMP_DIR}/chromedriver $(CHROMEDRIVER_PATH); \
		sudo chmod +x $(CHROMEDRIVER_PATH); \
	fi

firefox: bzip2 wget
	if [[ -z "$$(which firefox)" ]]; then \
		wget -O firefox.tar.xz -q "https://download.mozilla.org/?product=firefox-latest&os=linux64&lang=en-US"; \
		mkdir -p $$HOME/browsers; \
		tar -xaf firefox.tar.xz -C $$HOME/browsers; \
		sudo ln -s $$HOME/browsers/firefox/firefox $(FIREFOX_PATH); \
	fi

geckodriver: node-wct-local

golangci-lint: curl gpg
	if [ "$$(which golangci-lint)" == "" ]; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/${GOLANGCI_LINT_VERSION}/install.sh | sh -s -- -b $$(go env GOPATH)/bin; \
	fi

golint: git
	if [ "$$(which golint)" == "" ]; then \
		go install golang.org/x/lint/golint; \
	fi

mockgen: git
	if [ "$$(which mockgen)" == "" ]; then \
		go install go.uber.org/mock/mockgen; \
	fi

package_service: var-APP_PATH
	# Trim the potential "app.staging.yaml" suffix.
	if [[ "$(APP_PATH)" == "api/query/cache/service"* ]]; then \
		APP_PATH="api/query/cache/service"; \
	elif [[ "$(APP_PATH)" == "webapp/web"* ]]; then \
		APP_PATH="webapp/web"; \
	else \
		APP_PATH="$(APP_PATH)"; \
	fi ; \
	if [[ "$${APP_PATH}" == "api/query/cache/service" || "$${APP_PATH}" == "webapp/web" ]]; then \
		TMP_DIR=$$(mktemp -d); \
		rm -rf $(WPTD_PATH)$${APP_PATH}/wpt.fyi; \
		cp -r $(WPTD_PATH)* $${TMP_DIR}/; \
		mkdir $(WPTD_PATH)$${APP_PATH}/wpt.fyi; \
		cp -r $${TMP_DIR}/* $(WPTD_PATH)$${APP_PATH}/wpt.fyi/; \
		rm -rf $${TMP_DIR}; \
	fi

sys_deps: apt_update
	make gcloud
	make git
	make node

apt_update:
	sudo apt-get -qq update

bzip2: apt-get-bzip2
curl: apt-get-curl
gcc: apt-get-gcc
git: apt-get-git
jq: apt-get-jq
psmisc: apt-get-psmisc
python3: apt-get-python3.11
tox: apt-get-tox
unzip: apt-get-unzip
wget: apt-get-wget

java:
	@ # java has a different apt-get package name.
	if [[ "$$(which java)" == "" ]]; then \
		sudo apt-get install -qqy --no-install-suggests java-11-amazon-corretto-jdk; \
	fi

gpg:
	@ # gpg has a different apt-get package name.
	if [[ "$$(which gpg)" == "" ]]; then \
		sudo apt-get install -qqy --no-install-suggests gnupg; \
	fi

inotifywait:
	@ # inotifywait has a different apt-get package name.
	if [[ "$$(which inotifywait)" == "" ]]; then \
		sudo apt-get install -qqy --no-install-suggests inotify-tools; \
	fi

node: curl gpg
	if [[ "$$(which node)" == "" ]]; then \
		curl -sL https://deb.nodesource.com/setup_18.x | sudo -E bash -; \
		sudo apt-get install -qqy nodejs; \
	fi

gcloud: python3 curl gpg
	if [[ "$$(which gcloud)" == "" ]]; then \
		curl -s https://sdk.cloud.google.com > ./install-gcloud.sh; \
		bash ./install-gcloud.sh --disable-prompts --install-dir=$(HOME) > /dev/null; \
		rm -f ./install-gcloud.sh; \
		gcloud components install --quiet core gsutil; \
		gcloud config set disable_usage_reporting false; \
	fi

eslint: webapp_node_modules_all
	cd webapp; npm run lint

dev_data: FLAGS := -remote_host=staging.wpt.fyi
dev_data: git
	go run $(WPTD_PATH)util/populate_dev_data.go $(FLAGS)

gcloud_login: gcloud
	if [[ -z "$$(gcloud config list account --format "value(core.account)")" ]]; then \
		gcloud auth activate-service-account --key-file $(WPTD_PATH)client-secret.json; \
	fi

deployment_state: go_build gcloud_login package_service var-APP_PATH

deploy_staging: git apt-get-jq
deploy_staging: BRANCH_NAME := $$(git rev-parse --abbrev-ref HEAD)
deploy_staging: deployment_state var-BRANCH_NAME
	gcloud config set project wptdashboard-staging
	if [[ "$(BRANCH_NAME)" == "refs/heads/main" ]]; then \
		util/deploy.sh -q -r -p $(APP_PATH); \
	else \
		util/deploy.sh -q -b $(BRANCH_NAME) $(APP_PATH); \
	fi
	rm -rf $(WPTD_PATH)api/query/cache/service/wpt.fyi
	rm -rf $(WPTD_PATH)webapp/web/wpt.fyi

cleanup_staging_versions: gcloud_login
	$(WPTD_PATH)/util/cleanup-versions.sh

deploy_production: deployment_state
	gcloud config set project wptdashboard
	util/deploy.sh -r $(APP_PATH)
	rm -rf $(WPTD_PATH)api/query/cache/service/wpt.fyi
	rm -rf $(WPTD_PATH)webapp/web/wpt.fyi

webapp_node_modules_all: node
	cd webapp; npm install

webapp_node_modules_prod: webapp_node_modules_all
	cd webapp; npm prune --production

xvfb:
	if [[ "$(USE_FRAME_BUFFER)" == "true" && "$$(which Xvfb)" == "" ]]; then \
		sudo apt-get install -qqy --no-install-suggests xvfb; \
	fi

gcloud-%: gcloud
	gcloud components list --only-local-state --format="value(id)" 2>/dev/null | grep -q "$*" \
		|| gcloud components install --quiet $*

node-%: node
	@ echo "# Installing $*..."
	# Hack to (more quickly) detect whether a package is already installed (available in node).
	cd webapp; node -p "require('$*/package.json').version" 2>/dev/null || npm install --no-save $*

apt-get-%:
	if [[ "$$(which $*)" == "" ]]; then sudo apt-get install -qqy --no-install-suggests $*; fi

env-%:
	@ if [[ "${${*}}" = "" ]]; then echo "Environment variable $* not set"; exit 1; fi

var-%:
	@ if [[ "$($*)" = "" ]]; then echo "Make variable $* not set"; exit 1; fi
