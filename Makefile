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
USE_FRAME_BUFFER := true
STAGING := false

GO_FILES := $(shell find $(WPTD_PATH) -type f -name '*.go')
GO_TEST_FILES := $(shell find $(WPTD_PATH) -type f -name '*_test.go')

# Recursively expanded variables so that USE_FRAME_BUFFER can be expanded.
# These two macros are intended to run in the same shell as the test runners,
# which means they need to be in the same (continued) line.
START_XVFB = if [ "$(USE_FRAME_BUFFER)" == "true" ]; then \
	export DISPLAY=:99; (Xvfb :99 -screen 0 1024x768x24 -ac +extension GLX +render -noreset &); fi
STOP_XVFB = if [ "$(USE_FRAME_BUFFER)" == "true" ]; then killall Xvfb; fi

build: go_build

test: go_test python_test

lint: go_lint eslint

prepush: go_build test lint

python_test: python3 tox
	cd $(WPTD_PATH)results-processor; tox

go_build: go_deps
	cd $(WPTD_GO_PATH); go build ./...

go_lint: go_deps golint_deps go_test_tag_lint
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

go_small_test: go_deps
	cd $(WPTD_GO_PATH); go test -tags=small -v ./...

go_medium_test: go_deps dev_appserver_deps
	cd $(WPTD_GO_PATH); go test -tags=medium -v $(FLAGS) ./...

# Use sub-make because otherwise make would only execute the first invocation
# of _go_webdriver_test. Variables will be passed into sub-make implicitly.
go_large_test:
	make go_firefox_test
	make go_chrome_test

go_firefox_test: BROWSER := firefox
go_firefox_test: firefox | _go_webdriver_test

go_chrome_test: BROWSER := chrome
go_chrome_test: chrome | _go_webdriver_test

# _go_webdriver_test is not intended to be used directly; use go_firefox_test or
# go_chrome_test instead.
_go_webdriver_test: var-BROWSER java go_deps xvfb node-web-component-tester webserver_deps
	# This Go test manages Xvfb itself, so we don't start/stop Xvfb for it.
	# The following variables are defined here because we don't know the
	# paths before installing node-web-component-tester as the paths
	# include version strings.
	GECKODRIVER_PATH="$(shell find $(NODE_SELENIUM_PATH)geckodriver/ -type f -name '*geckodriver')"; \
	CHROMEDRIVER_PATH="$(shell find $(NODE_SELENIUM_PATH)chromedriver/ -type f -name '*chromedriver')"; \
	cd $(WPTD_PATH)webdriver; \
	go test -v -tags=large -args \
		-firefox_path=$(FIREFOX_PATH) \
		-geckodriver_path=$$GECKODRIVER_PATH \
		-chrome_path=$(CHROME_PATH) \
		-chromedriver_path=$$CHROMEDRIVER_PATH \
		-frame_buffer=$(USE_FRAME_BUFFER) \
		-staging=$(STAGING) \
		-browser=$(BROWSER) \
		$(FLAGS)

web_components_test: xvfb firefox chrome node-web-component-tester webserver_deps
	$(START_XVFB); \
	cd $(WPTD_PATH)webapp; \
	npm test || (($(STOP_XVFB)) && exit 1); \
	$(STOP_XVFB)

sys_update: apt_update | sys_deps
	gcloud components update
	sudo npm install -g npm

apt_update:
	sudo apt-get --quiet update

# Dependencies for running dev_appserver.py.
webserver_deps: webapp_deps dev_appserver_deps

webapp_deps: go_deps bower_components

dev_appserver_deps: gcloud-app-engine-python gcloud-app-engine-go gcloud-cloud-datastore-emulator

chrome:
	if [[ -z "$$(which google-chrome)" ]]; then \
		if [[ -z "$$(which chromium)" ]]; then \
			make apt-get-chromium; \
		fi; \
		sudo ln -s "$$(which chromium)" $(CHROME_PATH); \
	fi

firefox:
	if [[ -z "$$(which firefox)" ]]; then \
		make firefox_install; \
	fi

firefox_install: firefox_deps bzip2 wget java
	$(WPTD_PATH)webdriver/install.sh $$HOME/browsers
	sudo ln -s $$HOME/browsers/firefox/firefox $(FIREFOX_PATH)

firefox_deps:
	sudo apt-get install -qqy --no-install-suggests $$(apt-cache depends firefox-esr | grep Depends | sed "s/.*ends:\ //" | tr '\n' ' ')

go_deps: go_packages $(GO_FILES)

go_packages: git
	cd $(WPTD_GO_PATH); go get -t -tags="small medium large" ./...

golint_deps: git
	if [ "$$(which golint)" == "" ]; then \
		go get -u golang.org/x/lint/golint; \
	fi

package_service: var-APP_PATH
	if [[ "$(APP_PATH)" == "revisions/service" || "$(APP_PATH)" == "api/spanner/service" ]]; then \
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
python3: apt-get-python3
python: apt-get-python
tox: apt-get-tox
wget: apt-get-wget
bzip2: apt-get-bzip2

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
		curl -sL https://deb.nodesource.com/setup_8.x | sudo -E bash -; \
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

gcloud-login: gcloud  $(WPTD_PATH)client-secret.json
	gcloud auth activate-service-account --key-file $(WPTD_PATH)client-secret.json

deploy_staging: gcloud-login webapp_deps package_service var-BRANCH_NAME var-APP_PATH var-PROJECT
	gcloud config set project $(PROJECT)
	cd $(WPTD_PATH); util/deploy.sh -q -b $(BRANCH_NAME) $(APP_PATH)
	rm -rf $(WPTD_PATH)revisions/service/wpt.fyi
	rm -rf $(WPTD_PATH)api/spanner/service/wpt.fyi

cleanup_staging_versions: gcloud-login
	$(WPTD_GO_PATH)/util/cleanup-versions.sh

deploy_production: gcloud webapp_deps package_service var-APP_PATH var-PROJECT
	gcloud config set project $(PROJECT)
	cd $(WPTD_PATH); util/deploy.sh -p $(APP_PATH)
	rm -rf $(WPTD_PATH)revisions/service/wpt.fyi
	rm -rf $(WPTD_PATH)api/spanner/service/wpt.fyi

bower_components: git node-bower
	cd $(WPTD_PATH)webapp; npm run bower-components

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

# symlinks the Go folder for the results-analysis project to (this) wpt.fyi folder's
# sibling results-analysis folder.
results_analysis_symlink: RESULTS_ANALYSIS_PATH := $(WPT_PATH)/results-analysis
results_analysis_symlink: RESULTS_ANALYSIS_GO_PATH := $(WPT_GO_PATH)/results-analysis
results_analysis_symlink:
	@if [[ -L $(RESULTS_ANALYSIS_GO_PATH) && -d $(RESULTS_ANALYSIS_GO_PATH) ]]; \
	then echo "Already a symlink"; \
	else \
		if [ -e $(RESULTS_ANALYSIS_GO_PATH) ]; then rm -r $(RESULTS_ANALYSIS_GO_PATH); fi; \
		ln -s $(RESULTS_ANALYSIS_PATH) $(RESULTS_ANALYSIS_GO_PATH); \
	fi

gcloud-%: gcloud
	gcloud components list --filter="state[name]=Installed AND id=$*" | grep " $* " \
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
