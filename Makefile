# Copyright 2017 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

# Make targets in this file are intended to be run inside the Docker container
# environment.

# Make targets can be run in a host environment, but that requires ensuring
# the correct version of tools are installed and environment variables are
# set appropriately.

SHELL := /bin/bash

START_XVFB = export DISPLAY=99; Xvfb :99 -screen 0 1024x768x24 -ac +extension GLX +render -noreset &
STOP_XVFB = killall Xvfb

export GOPATH=$(shell go env GOPATH)

# WPTD_PATH will have a trailing slash, e.g. /home/user/wpt.fyi/
WPTD_PATH := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
WPTD_GO_PATH ?= $(GOPATH)/src/github.com/web-platform-tests/wpt.fyi
NODE_SELENIUM_PATH ?= $(WPTD_PATH)webapp/node_modules/selenium-standalone/.selenium/
SELENIUM_SERVER_PATH ?= $(NODE_SELENIUM_PATH)selenium-server/3.8.1-server.jar
GECKODRIVER_PATH ?= $(NODE_SELENIUM_PATH)geckodriver/0.20.0-x64-geckodriver
FIREFOX_PATH ?= $$HOME/browsers/firefox/firefox
USE_FRAME_BUFFER ?= true
NVM_URL ?= https://raw.githubusercontent.com/creationix/nvm/v0.33.8/install.sh

GO_FILES := $(shell find $(WPTD_PATH) -type f -name '*.go')
GO_TEST_FILES := $(shell find $(WPTD_PATH) -type f -name '*_test.go')

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
	cd $(WPTD_GO_PATH); golint -set_exit_status api/
	cd $(WPTD_GO_PATH); golint -set_exit_status revisions/
	cd $(WPTD_GO_PATH); golint -set_exit_status shared/
	cd $(WPTD_GO_PATH); golint -set_exit_status util/
	cd $(WPTD_GO_PATH); golint -set_exit_status webapp/
	# Printing files with differences between current/gofmt'd output, asserting empty...
	cd $(WPTD_GO_PATH); ! gofmt -d $(GO_FILES) 2>&1 | read || ! echo $$(gofmt -l $(GO_FILES))

go_test_tag_lint:
	@ echo "# Printing a list of test files without +build tag, asserting empty..."
	TAGLESS=$$(grep -PL '\/\/\s?\+build !?(small|medium|large)' $(GO_TEST_FILES));
	if [ -n "$$TAGLESS" ]; then echo -e "Files are missing +build tags:\n$$TAGLESS" && exit 1; fi

go_test: go_small_test go_medium_test

go_small_test: go_deps
	cd $(WPTD_GO_PATH); go test -tags=small -v ./...

go_medium_test: go_deps dev_appserver_deps
	# Hack to work around https://github.com/golang/appengine/issues/136
	cd $(GOPATH)/src/github.com/golang/protobuf; git checkout ac606b1
	cd $(WPTD_GO_PATH); go test -tags=medium -v $(FLAGS) ./...

go_large_test: go_webdriver_test

integration_test: go_webdriver_test web_components_test

go_webdriver_test: STAGING := false
go_webdriver_test: go_deps xvfb firefox node-web-component-tester webserver_deps
	if [ "$(USE_FRAME_BUFFER)" == "true" ]; then ($(START_XVFB)); fi
	cd $(WPTD_PATH)webdriver; go test -v -tags=large \
			--selenium_path=$(SELENIUM_SERVER_PATH) \
			--firefox_path=$(FIREFOX_PATH) \
			--geckodriver_path=$(GECKODRIVER_PATH) \
			--frame_buffer=$(USE_FRAME_BUFFER) \
			--staging=$(STAGING)
	if [[ "$(USE_FRAME_BUFFER)" == "true" ]]; then $(STOP_XVFB); fi

web_components_test: xvfb firefox chrome node-web-component-tester webserver_deps
	$(START_XVFB)
	cd $(WPTD_PATH)webapp; export DISPLAY=:99.0; npm test
	$(STOP_XVFB)

sys_update: sys_deps
	sudo apt-get update
	gcloud components update
	npm install -g npm

# Dependencies for running dev_appserver.py.
webserver_deps: webapp_deps dev_appserver_deps

webapp_deps: go_deps bower_components

dev_appserver_deps: gcloud-app-engine-python gcloud-app-engine-go

chrome: browser_deps
	if [[ -z "$$(which google-chrome)" ]]; then \
		if [[ -z "$$(which chromium)" ]]; then \
			make apt-get-chromium; \
		fi; \
		sudo ln -s "$$(which chromium)" /usr/bin/google-chrome; \
	fi

firefox: browser_deps
	if [[ "$$(which firefox)" == "" ]]; then \
	  $(WPTD_PATH)webdriver/install.sh $$HOME/browsers; \
		sudo ln -s $(FIREFOX_PATH) /usr/bin/firefox; \
	fi

browser_deps: wget java
	sudo apt-get install --assume-yes --no-install-suggests $$(apt-cache depends firefox-esr chromedriver |  grep Depends | sed "s/.*ends:\ //" | tr '\n' ' ')

go_deps: gcloud go_packages $(GO_FILES)

go_packages: git
	cd $(WPTD_GO_PATH); go get -t -tags="small medium large" ./...

golint_deps: git go_deps
	# Manual git clone + install is a workaround for #85.
	if [ "$$(which golint)" == "" ]; then \
		mkdir -p "$(GOPATH)/src/golang.org/x"; \
		cd "$(GOPATH)/src/golang.org/x" && git clone https://github.com/golang/lint; \
		cd "$(GOPATH)/src/golang.org/x/lint" && go get ./... && go install ./...; \
	fi

sys_deps: curl gpg node gcloud git

curl: apt-get-curl
git: apt-get-git
python3: apt-get-python3
python: apt-get-python
tox: apt-get-tox
wget: apt-get-wget

java:
	@ # java has a different apt-get package name.
	if [[ "$$(which java)" == "" ]]; then \
		sudo apt-get install --assume-yes --no-install-suggests openjdk-8-jdk; \
	fi

gpg:
	@ # gpg has a different apt-get package name.
	if [[ "$$(which gpg)" == "" ]]; then \
		sudo apt-get install --assume-yes --no-install-suggests gnupg; \
	fi

node: curl gpg
	if [[ "$$(which node)" == "" ]]; then \
		curl -sL https://deb.nodesource.com/setup_8.x | sudo -E bash -; \
		sudo apt-get install -y nodejs; \
	fi

gcloud: python curl gpg
	@ echo "travis_fold:start:gcloud_install"
	if [[ "$$(which gcloud)" == "" ]]; then \
		curl -s https://sdk.cloud.google.com > ./install-gcloud.sh; \
		bash ./install-gcloud.sh --disable-prompts --install-dir=$(HOME); \
		rm -f ./install-gcloud.sh; \
		gcloud components install --quiet core gsutil; \
		gcloud config set disable_usage_reporting false; \
	fi
	@ echo "travis_fold:end:gcloud_install"

eslint: node-babel-eslint node-eslint node-eslint-plugin-html
	cd $(WPTD_PATH)webapp; npm run lint

dev_data:
	cd $(WPTD_GO_PATH)/util; go get -t ./...
	go run util/populate_dev_data.go $(FLAGS)

deploy_staging: gcloud webapp_deps var-BRANCH_NAME var-APP_PATH var-PROJECT $(WPTD_PATH)client-secret.json
	gcloud config set project $(PROJECT)
	gcloud auth activate-service-account --key-file $(WPTD_PATH)client-secret.json
	cd $(WPTD_PATH); util/deploy.sh -q -b $(BRANCH_NAME) $(APP_PATH)

bower_components: git node-bower
	cd $(WPTD_PATH)webapp; npm run bower-components

xvfb:
	if [[ "$(USE_FRAME_BUFFER)" == "true" && "$$(which Xvfb)" == "" ]]; then \
		sudo apt-get install --assume-yes --no-install-suggests xvfb; \
	fi

gcloud-%: gcloud
	gcloud components list --filter="state[name]=Installed AND id=$*" | grep " $* " \
		|| gcloud components install --quiet $*

node-%: node
	@ echo "# Installing $*..."
	# Hack to (more quickly) detect whether a package is already installed (available in node).
	cd $(WPTD_PATH)webapp; node -p "require('$*/package.json').version" 2>/dev/null || npm install --no-save $*

apt-get-%:
	if [[ "$$(which $*)" == "" ]]; then sudo apt-get install --quiet --assume-yes --no-install-suggests $*; fi

env-%:
	@ if [[ "${${*}}" = "" ]]; then echo "Environment variable $* not set"; exit 1; fi

var-%:
	@ if [[ "$($*)" = "" ]]; then echo "Make variable $* not set"; exit 1; fi
