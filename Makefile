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
GOPATH := $(shell go env GOPATH)
# WPTD_PATH will have a trailing slash, e.g. /home/user/wpt.fyi/
WPTD_PATH := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
WPT_PATH := $(dir $(WPTD_PATH)/../)
WPT_GO_PATH := $(GOPATH)/src/github.com/web-platform-tests
WPTD_GO_PATH := $(WPT_GO_PATH)/wpt.fyi
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
	cd $(WPTD_PATH)results-processor; tox

go_build: git mockgen
	cd $(WPTD_GO_PATH); go get ./...
	cd $(WPTD_GO_PATH); go generate ./...

go_build_test: go_build apt-get-gcc
	cd $(WPTD_GO_PATH); go get -t -tags="small medium large" ./...

go_lint: golint_deps go_test_tag_lint
	@echo "# Linting the go packages..."
	golint -set_exit_status $(WPTD_GO_PATH)/api/...
	# Skip revisions/test
	golint -set_exit_status $(WPTD_GO_PATH)/revisions/{announcer,api,epoch,git,service}/...
	golint -set_exit_status $(WPTD_GO_PATH)/shared/...
	golint -set_exit_status $(WPTD_GO_PATH)/util/...
	golint -set_exit_status $(WPTD_GO_PATH)/webapp/...
	golint -set_exit_status $(WPTD_GO_PATH)/webdriver/...

go_test_tag_lint:
	# Printing a list of test files without +build tag, asserting empty...
	@TAGLESS=$$(grep -PL '\/\/\s?\+build !?(small|medium|large)' $(GO_TEST_FILES)); \
	if [ -n "$$TAGLESS" ]; then echo -e "Files are missing +build tags:\n$$TAGLESS" && exit 1; fi

go_test: go_small_test go_medium_test

go_small_test: go_build_test
	cd $(WPTD_GO_PATH); go test -tags=small $(VERBOSE) ./...

go_medium_test: go_build_test dev_appserver_deps
	cd $(WPTD_GO_PATH); go test -tags=medium $(VERBOSE) $(FLAGS) ./...

# Use sub-make because otherwise make would only execute the first invocation
# of _go_webdriver_test. Variables will be passed into sub-make implicitly.
go_large_test:
	make go_firefox_test
	make go_chrome_test

go_firefox_test: BROWSER := firefox
go_firefox_test: firefox | _go_webdriver_test

go_chrome_test: BROWSER := chrome
go_chrome_test: chrome chromedriver | _go_webdriver_test

# _go_webdriver_test is not intended to be used directly; use go_firefox_test or
# go_chrome_test instead.
_go_webdriver_test: var-BROWSER java go_build_test xvfb node-web-component-tester webserver_deps
	# This Go test manages Xvfb itself, so we don't start/stop Xvfb for it.
	# The following variables are defined here because we don't know the
	# paths before installing node-web-component-tester as the paths
	# include version strings.
	GECKODRIVER_PATH="$(shell find $(NODE_SELENIUM_PATH)geckodriver/ -type f -name '*geckodriver')"; \
	cd $(WPTD_PATH)webdriver; \
	go test $(VERBOSE) -timeout=15m -tags=large -args \
		-firefox_path=$(FIREFOX_PATH) \
		-geckodriver_path=$$GECKODRIVER_PATH \
		-chrome_path=$(CHROME_PATH) \
		-chromedriver_path=$(CHROMEDRIVER_PATH) \
		-frame_buffer=$(USE_FRAME_BUFFER) \
		-staging=$(STAGING) \
		-browser=$(BROWSER) $(FLAGS)

# NOTE: psmisc includes killall, needed by wct.sh
web_components_test: xvfb firefox chrome webapp_node_modules_all apt-get-psmisc
	util/wct.sh $(USE_FRAME_BUFFER)

sys_update: apt_update | sys_deps
	gcloud components update
	sudo npm install -g npm

apt_update:
	sudo apt-get --quiet update

# Dependencies for running dev_appserver.py.
webserver_deps: webapp_deps dev_appserver_deps

webapp_deps: go_build webapp_node_modules

dev_appserver_deps: gcloud-app-engine-python gcloud-app-engine-go gcloud-cloud-datastore-emulator

chrome: wget
	if [[ -z "$$(which google-chrome)" ]]; then \
		ARCHIVE=google-chrome-stable_current_amd64.deb; \
		wget -q https://dl.google.com/linux/direct/$${ARCHIVE}; \
		sudo dpkg --install $${ARCHIVE} || true; \
		sudo apt-get install --fix-broken -qqy; \
		sudo dpkg --install $${ARCHIVE}; \
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

golint_deps: git
	if [ "$$(which golint)" == "" ]; then \
		go get -u golang.org/x/lint/golint; \
	fi

mockgen: git
	if [ "$$(which mockgen)" == "" ]; then \
		go get -u github.com/golang/mock/mockgen; \
	fi

package_service: var-APP_PATH
	# Trim the potential "app.staging.yaml" suffix.
	if [[ "$(APP_PATH)" == "api/query/cache/service"* ]]; then \
		APP_PATH="api/query/cache/service"; \
	fi ;\
	if [[ "$(APP_PATH)" == "revisions/service" || "$(APP_PATH)" == "api/query/cache/service" ]]; then \
		export TMP_DIR=$$(mktemp -d); \
		rm -rf $(WPTD_PATH)$(APP_PATH)/wpt.fyi; \
		cp -r $(WPTD_PATH)* $${TMP_DIR}/; \
		mkdir $(WPTD_PATH)$(APP_PATH)/wpt.fyi; \
		cp -r $${TMP_DIR}/* $(WPTD_PATH)$(APP_PATH)/wpt.fyi/; \
		rm -rf $${TMP_DIR}; \
	fi

sys_deps: curl gpg node gcloud git

curl: apt-get-curl
git: apt-get-git
python3: apt-get-python3.6
python: apt-get-python
tox: apt-get-tox
wget: apt-get-wget
bzip2: apt-get-bzip2
unzip: apt-get-unzip

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

eslint: node-babel-eslint node-eslint node-eslint-plugin-html
	cd $(WPTD_PATH)webapp; npm run lint

dev_data: FLAGS := -host=staging.wpt.fyi
dev_data: git
	cd $(WPTD_GO_PATH)/util; go get -t ./...
	go run $(WPTD_GO_PATH)/util/populate_dev_data.go $(FLAGS)

gcloud-login: gcloud
	if [[ -z "$$(gcloud config list account --format "value(core.account)")" ]]; then \
		gcloud auth activate-service-account --key-file $(WPTD_PATH)client-secret.json; \
	fi

deployment_state: gcloud-login webapp_deps package_service var-APP_PATH

deploy_staging: git
deploy_staging: BRANCH_NAME := $$(git rev-parse --abbrev-ref HEAD)
deploy_staging: deployment_state var-BRANCH_NAME
	gcloud config set project wptdashboard-staging
	if [[ "$(BRANCH_NAME)" == "master" ]]; then \
		cd $(WPTD_PATH); util/deploy.sh -q -r -p $(APP_PATH); \
	else \
		cd $(WPTD_PATH); util/deploy.sh -q -b $(BRANCH_NAME) $(APP_PATH); \
	fi
	rm -rf $(WPTD_PATH)revisions/service/wpt.fyi
	rm -rf $(WPTD_PATH)api/query/cache/service/wpt.fyi

cleanup_staging_versions: gcloud-login
	$(WPTD_GO_PATH)/util/cleanup-versions.sh

deploy_production: deployment_state
	gcloud config set project wptdashboard
	cd $(WPTD_PATH); util/deploy.sh -r $(APP_PATH)
	rm -rf $(WPTD_PATH)revisions/service/wpt.fyi
	rm -rf $(WPTD_PATH)api/query/cache/service/wpt.fyi

webapp_node_modules: node
	cd $(WPTD_PATH)webapp; npm install --production

webapp_node_modules_all: node
	cd $(WPTD_PATH)webapp; npm install

webapp_node_modules_prune: webapp_node_modules
	cd $(WPTD_PATH)webapp; npm prune --production

xvfb:
	if [[ "$(USE_FRAME_BUFFER)" == "true" && "$$(which Xvfb)" == "" ]]; then \
		sudo apt-get install -qqy --no-install-suggests xvfb; \
	fi

# symlinks the Go folder for the wpt.fyi project to (this) folder.
wpt_fyi_symlink:
	@if [[ -L $(WPTD_GO_PATH) && -d $(WPTD_GO_PATH) ]]; \
	then echo "Already a symlink"; \
	else \
		if [ -e $(WPTD_GO_PATH) ]; then rm -r $(WPTD_GO_PATH); fi; \
		ln -s $(WPTD_PATH) $(WPTD_GO_PATH); \
	fi

gcloud-%: gcloud
	gcloud components list --only-local-state --format="value(id)" 2>/dev/null | grep -q "$*" \
		|| gcloud components install --quiet $*

node-%: node
	@ echo "# Installing $*..."
	# Hack to (more quickly) detect whether a package is already installed (available in node).
	cd $(WPTD_PATH)webapp; node -p "require('$*/package.json').version" 2>/dev/null || npm install --no-save $*

apt-get-%:
	if [[ "$$(which $*)" == "" ]]; then sudo apt-get install -qqy --no-install-suggests $*; fi

env-%:
	@ if [[ "${${*}}" = "" ]]; then echo "Environment variable $* not set"; exit 1; fi

var-%:
	@ if [[ "$($*)" = "" ]]; then echo "Make variable $* not set"; exit 1; fi
