# Copyright 2017 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

# Make targets in this file are intended to be run inside the Docker container
# environment.

# Make targets can be run in a host environment, but that requires ensuring
# the correct version of tools are installed and environment variables are
# set appropriately.

SHELL := /bin/bash

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

test: go_test

lint: go_lint eslint

prepush: go_build test lint

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
	cd $(WPTD_GO_PATH); go test -tags=medium -v ./...

go_large_test: go_webdriver_test

integration_test: go_webdriver_test web_components_test

go_webdriver_test: go_webdriver_deps
	cd $(WPTD_PATH)webdriver; go test -v -tags=large \
			--selenium_path=$(SELENIUM_SERVER_PATH) \
			--firefox_path=$(FIREFOX_PATH) \
			--geckodriver_path=$(GECKODRIVER_PATH) \
			--frame_buffer=$(USE_FRAME_BUFFER)

web_components_test: webdriver_deps web_component_tester
	cd $(WPTD_PATH)webapp; export DISPLAY=:99.0; npm test

sys_update: sys_deps
	sudo apt-get update
	gcloud components update
	npm install -g npm

go_webdriver_deps: go_deps webdriver_deps webserver_deps

webdriver_deps: xvfb browser_deps webserver_deps web_component_tester

# Dependencies for running dev_appserver.py.
webserver_deps: build bower_components dev_appserver_deps

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

browser_deps: wget
	sudo apt-get install --assume-yes --no-install-suggests openjdk-8-jdk $$(apt-cache depends firefox-esr chromedriver |  grep Depends | sed "s/.*ends:\ //" | tr '\n' ' ')

go_deps: git gcloud $(GO_FILES)
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
python: apt-get-python
git: apt-get-git
wget: apt-get-wget

gpg:
	@ # gpg has a different apt-get package name.
	if [[ "$$(which gpg)" == "" ]]; then \
		sudo apt-get install --assume-yes --no-install-suggests gnupg; \
	fi

node: curl
	if [[ "$$(which node)" == "" ]]; then \
		curl -sL https://deb.nodesource.com/setup_10.x | sudo -E bash -; \
		sudo apt-get install -y nodejs; \
	fi

npm: apt-get-npm

gcloud: python curl gpg
	if [[ "$$(which gcloud)" == "" ]]; then \
		curl -s https://sdk.cloud.google.com > ./install-gcloud.sh; \
		bash ./install-gcloud.sh --disable-prompts --install-dir=$(HOME); \
		rm -f ./install-gcloud.sh; \
		gcloud components install --quiet core gsutil; \
		gcloud config set disable_usage_reporting false; \
	fi

eslint: node-babel-eslint node-eslint node-eslint-plugin-html
	cd $(WPTD_PATH)webapp; npm run lint

dev_data:
	cd $(WPTD_GO_PATH)/util; go get -t ./...
	go run util/populate_dev_data.go $(FLAGS)

deploy_staging: bower_components env-BRANCH_NAME env-APP_PATH
	gcloud config set project wptdashboard
	gcloud auth activate-service-account --key-file $(WPTD_PATH)client-secret.json
	cd $(WPTD_PATH); util/deploy.sh -q -b $(BRANCH_NAME) $(APP_PATH)

web_component_tester: chrome firefox node-web-component-tester bower_components

bower_components: node-bower
	cd $(WPTD_PATH)webapp; npm run bower-components

xvfb:
	if [[ "$(USE_FRAME_BUFFER)" == "true" && "$$(which Xvfb)" == "" ]]; then \
		sudo apt-get install --assume-yes --no-install-suggests xvfb; \
		export DISPLAY=99; Xvfb :99 -screen 0 1024x768x24 -ac +extension GLX +render -noreset & \
	fi

gcloud-%: gcloud
	gcloud components install --quiet $*

node-%: node npm
	@ echo "# Installing $*..."
	cd $(WPTD_PATH)webapp; node -p "require('$*/package.json').version" 2>/dev/null || npm install --no-save $*

apt-get-%:
	if [[ "$$(which $*)" == "" ]]; then sudo apt-get install --assume-yes --no-install-suggests $*; fi

env-%:
	@ if [[ "${${*}}" = "" ]]; then echo "Environment variable $* not set"; exit 1; fi
