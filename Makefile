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
WPT_PATH := $(dir $(WPTD_PATH)/../)
NODE_SELENIUM_PATH := $(WPTD_PATH)webapp/node_modules/selenium-standalone/.selenium/
FIREFOX_PATH := /usr/bin/firefox
CHROME_PATH := /usr/bin/google-chrome
CHROMEDRIVER_PATH=/usr/bin/chromedriver
USE_FRAME_BUFFER := true
STAGING := false
VERBOSE := -v

GO_FILES := $(shell find $(WPTD_PATH) -type f -name '*.go')
GO_TEST_FILES := $(shell find $(WPTD_PATH) -type f -name '*_test.go')

build: go_build

test: go_test python_test

lint: go_lint eslint

prepush: VERBOSE := $() # Empty out the verbose flag.
prepush: go_build go_test lint

python_test: python3 tox
	tox -c results-processor/

# NOTE: We prune before generate, because node_modules are packr'd into the
# binary (and part of the build).
go_build: git mockgen packr2
	make webapp_node_modules_prune
	go generate ./...
	go build ./...

go_lint: golint go_test_tag_lint
	@echo "# Linting the go packages..."
	golint -set_exit_status ./api/...
	# Skip revisions/test
	golint -set_exit_status ./revisions/{announcer,api,epoch,git,service}/...
	golint -set_exit_status ./shared/...
	golint -set_exit_status ./util/...
	golint -set_exit_status ./webapp/...
	golint -set_exit_status ./webdriver/...

go_test_tag_lint:
	# Printing a list of test files without +build tag, asserting empty...
	@TAGLESS=$$(grep -PL '\/\/\s?\+build !?(small|medium|large)' $(GO_TEST_FILES)); \
	if [ -n "$$TAGLESS" ]; then echo -e "Files are missing +build tags:\n$$TAGLESS" && exit 1; fi

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

puppeteer_chrome_test: chrome dev_appserver_deps webdriver_node_deps
	cd webdriver; npm test

webdriver_node_deps:
	cd webdriver; npm install

# _go_webdriver_test is not intended to be used directly; use go_firefox_test or
# go_chrome_test instead.
_go_webdriver_test: var-BROWSER java go_build xvfb geckodriver dev_appserver_deps gcc
	# This Go test manages Xvfb itself, so we don't start/stop Xvfb for it.
	# The following variables are defined here because we don't know the
	# path before installing geckodriver as it includes version strings.
	GECKODRIVER_PATH="$(shell find $(NODE_SELENIUM_PATH)geckodriver/ -type f -name '*geckodriver')"; \
	cd webdriver; \
	go test $(VERBOSE) -timeout=15m -tags=large -args \
		-firefox_path=$(FIREFOX_PATH) \
		-geckodriver_path=$$GECKODRIVER_PATH \
		-chrome_path=$(CHROME_PATH) \
		-chromedriver_path=$(CHROMEDRIVER_PATH) \
		-frame_buffer=$(USE_FRAME_BUFFER) \
		-staging=$(STAGING) \
		-test.timeout=30m \
		-browser=$(BROWSER) $(FLAGS)

# NOTE: psmisc includes killall, needed by wct.sh
web_components_test: xvfb firefox chrome webapp_node_modules_all psmisc
	util/wct.sh $(USE_FRAME_BUFFER)

dev_appserver_deps: gcloud-app-engine-python gcloud-app-engine-go gcloud-cloud-datastore-emulator

chrome: wget
	if [[ -z "$$(which google-chrome)" ]]; then \
		ARCHIVE=google-chrome-stable_current_amd64.deb; \
		wget -q https://dl.google.com/linux/direct/$${ARCHIVE}; \
		sudo dpkg --install $${ARCHIVE} 2>/dev/null || true; \
		sudo apt-get install --fix-broken -qqy; \
		sudo dpkg --install $${ARCHIVE} 2>/dev/null; \
	fi

# https://sites.google.com/a/chromium.org/chromedriver/downloads/version-selection
chromedriver: wget unzip chrome
	if [[ ! -f "$(CHROMEDRIVER_PATH)" ]]; then \
		CHROME_VERSION=$$(google-chrome --version | grep -ioE "[0-9]+\.[0-9]+\.[0-9]+"); \
		CHROMEDRIVER_VERSION=$$(curl https://chromedriver.storage.googleapis.com/LATEST_RELEASE_$${CHROME_VERSION}); \
		wget -q https://chromedriver.storage.googleapis.com/$${CHROMEDRIVER_VERSION}/chromedriver_linux64.zip; \
		sudo unzip chromedriver_linux64.zip -d $$(dirname $(CHROMEDRIVER_PATH)); \
		sudo chmod +x $(CHROMEDRIVER_PATH); \
	fi

firefox:
	if [[ -z "$$(which firefox)" ]]; then \
		make firefox_install; \
	fi

firefox_install: firefox_deps bzip2 wget java
	$(WPTD_PATH)webdriver/install.sh $$HOME/browsers
	sudo ln -s $$HOME/browsers/firefox/firefox $(FIREFOX_PATH)

firefox_deps:
	sudo apt-get install -qqy --no-install-suggests $$(apt-cache depends firefox | grep Depends | sed "s/.*ends:\ //" | tr '\n' ' ')

geckodriver: node-selenium-standalone
	cd webapp; `npm bin`/selenium-standalone install --singleDriverInstall=firefox

golint: git
	if [ "$$(which golint)" == "" ]; then \
		go install golang.org/x/lint/golint; \
	fi

mockgen: git
	if [ "$$(which mockgen)" == "" ]; then \
		go install github.com/golang/mock/mockgen; \
	fi

packr2: git
	if [ "$$(which packr2)" == "" ]; then \
		go install github.com/gobuffalo/packr/v2/packr2; \
	fi

package_service: var-APP_PATH
	# Trim the potential "app.staging.yaml" suffix.
	if [[ "$(APP_PATH)" == "api/query/cache/service"* ]]; then \
		APP_PATH="api/query/cache/service"; \
	else \
		APP_PATH="$(APP_PATH)"; \
	fi ; \
	if [[ "$${APP_PATH}" == "revisions/service" || "$${APP_PATH}" == "api/query/cache/service" ]]; then \
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
psmisc: apt-get-psmisc
python3: apt-get-python3.7
python: apt-get-python
tox: apt-get-tox
unzip: apt-get-unzip
wget: apt-get-wget

java:
	@ # java has a different apt-get package name.
	if [[ "$$(which java)" == "" ]]; then \
		sudo apt-get install -qqy --no-install-suggests openjdk-8-jdk; \
	fi

gpg:
	@ # gpg has a different apt-get package name.
	if [[ "$$(which gpg)" == "" ]]; then \
		sudo apt-get install -qqy --no-install-suggests gnupg; \
	fi

node: curl gpg
	if [[ "$$(which node)" == "" ]]; then \
		curl -sL https://deb.nodesource.com/setup_10.x | sudo -E bash -; \
		sudo apt-get install -qqy nodejs; \
	fi

gcloud: python curl gpg
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
	go run $(WPTD_PATH)/util/populate_dev_data.go $(FLAGS)

gcloud_login: gcloud
	if [[ -z "$$(gcloud config list account --format "value(core.account)")" ]]; then \
		gcloud auth activate-service-account --key-file $(WPTD_PATH)client-secret.json; \
	fi

deployment_state: go_build gcloud_login package_service var-APP_PATH

deploy_staging: git apt-get-jq
deploy_staging: BRANCH_NAME := $$(git rev-parse --abbrev-ref HEAD)
deploy_staging: deployment_state var-BRANCH_NAME
	gcloud config set project wptdashboard-staging
	if [[ "$(BRANCH_NAME)" == "master" ]]; then \
		util/deploy.sh -q -r -p $(APP_PATH); \
	else \
		util/deploy.sh -q -b $(BRANCH_NAME) $(APP_PATH); \
	fi
	rm -rf $(WPTD_PATH)revisions/service/wpt.fyi
	rm -rf $(WPTD_PATH)api/query/cache/service/wpt.fyi

cleanup_staging_versions: gcloud_login
	$(WPTD_PATH)/util/cleanup-versions.sh

deploy_production: deployment_state
	gcloud config set project wptdashboard
	util/deploy.sh -r $(APP_PATH)
	rm -rf $(WPTD_PATH)revisions/service/wpt.fyi
	rm -rf $(WPTD_PATH)api/query/cache/service/wpt.fyi

webapp_node_modules: node
	cd webapp; npm install --production

webapp_node_modules_all: node
	cd webapp; npm install

webapp_node_modules_prune: webapp_node_modules
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
