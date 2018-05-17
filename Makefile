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
export PATH:=$(HOME)/google-cloud-sdk/bin:$(PATH)

# WPTD_PATH will have a trailing slash, e.g. /home/user/wpt.fyi/
WPTD_PATH := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
WPTD_GO_PATH ?= $(GOPATH)/src/github.com/web-platform-tests/wpt.fyi
WEBDRIVER_PATH ?= $(WPTD_GO_PATH)/webdriver
BROWSERS_PATH ?= $(HOME)/browsers
SELENIUM_PATH ?= $(BROWSERS_PATH)/selenium
FIREFOX_PATH ?= $(BROWSERS_PATH)/firefox/firefox
GECKODRIVER_PATH ?= $(BROWSERS_PATH)/geckodriver

GO_FILES := $(shell find $(WPTD_PATH) -type f -name '*.go')
GO_TEST_FILES := $(shell find $(WPTD_PATH) -type f -name '*_test.go')

build: go_build

test: go_test

lint: go_lint eslint

prepush: build test lint

go_build: go_deps
	cd $(WPTD_GO_PATH); go build ./...

go_lint: go_deps go_test_tag_lint
	@echo "# Linting the go packages..."
	@cd $(WPTD_GO_PATH); golint -set_exit_status api/
	@cd $(WPTD_GO_PATH); golint -set_exit_status revisions/
	@cd $(WPTD_GO_PATH); golint -set_exit_status shared/
	@cd $(WPTD_GO_PATH); golint -set_exit_status util/
	@cd $(WPTD_GO_PATH); golint -set_exit_status webapp/
	# Printing files with differences between current/gofmt'd output, asserting empty...
	@cd $(WPTD_GO_PATH); ! gofmt -d $(GO_FILES) 2>&1 | read || ! echo $$(gofmt -l $(GO_FILES))

go_test_tag_lint:
	# Printing a list of test files without +build tag, asserting empty...
	@TAGLESS=$$(grep -PL '\/\/ \+build !?(small|medium|large)' $(GO_TEST_FILES)); \
			if [ -n "$$TAGLESS" ]; then echo -e "Files are missing +build tags:\n$$TAGLESS" && exit 1; fi

go_test: go_small_test go_medium_test

go_small_test: go_deps
	cd $(WPTD_GO_PATH); go test -tags=small -v ./...

go_medium_test: go_deps
	cd $(WPTD_GO_PATH); go test -tags=medium -v ./...
	
go_large_test: go_webdriver_test

go_webdriver_test: go_webdriver_deps
	cd $(WEBDRIVER_PATH); go test -v -tags=large \
			--selenium_path=$(SELENIUM_PATH) \
			--firefox_path=$(FIREFOX_PATH) \
			--geckodriver_path=$(GECKODRIVER_PATH)

sys_update: sys_deps
	sudo apt-get update
	gcloud components update
	npm install -g npm

go_webdriver_deps: go_deps webdriver_deps

webdriver_deps:
	cd $(WPTD_PATH)webapp; npm install web-component-tester --unsafe-perm
	cd $(WEBDRIVER_PATH); ./install.sh $(BROWSERS_PATH)

go_deps: sys_deps $(find .  -type f | grep '\.go$' | grep -v '\.pb.go$')
	# Manual git clone + install is a workaround for #85.
	if [ "$$(which golint)" == "" ]; \
		then \
		mkdir -p "$(GOPATH)/src/golang.org/x"; \
		cd "$(GOPATH)/src/golang.org/x" && git clone https://github.com/golang/lint; \
		cd "$(GOPATH)/src/golang.org/x/lint" && go get ./... && go install ./...; \
	fi
	cd $(WPTD_GO_PATH); go get -t -tags="small medium large" ./...

sys_deps:
	if [[ "$$(which curl)" == "" ]]; \
		then \
		sudo apt-get install --assume-yes --no-install-suggests curl; \
	fi
	if [[ "$$(which git)" == "" ]]; \
		then \
		sudo apt-get install --assume-yes --no-install-suggests git; \
	fi
	if [[ "$$(which python)" == "" ]]; \
		then \
		sudo apt-get install --assume-yes --no-install-suggests python; \
	fi
	if [[ "$$(which gpg)" == "" ]]; \
	then \
		sudo apt-get install --assume-yes --no-install-suggests gnupg; \
	fi
	if [[ "$$(which gcloud)" == "" ]]; \
		then \
		curl -s https://sdk.cloud.google.com > ./install-gcloud.sh; \
		bash ./install-gcloud.sh --disable-prompts --install-dir=$(HOME); \
		rm -f ./install-gcloud.sh; \
		gcloud components install --quiet \
			app-engine-go \
			core \
			gsutil \
			app-engine-python; \
		gcloud config set disable_usage_reporting false; \
	fi
	if [[ "$$(which nodejs)" == "" ]]; \
	then \
		curl -sL https://deb.nodesource.com/setup_8.x | sudo -E bash -; \
		sudo apt-get install --assume-yes --no-install-suggests nodejs; \
	fi
	if [[ "$$(which npm)" == "" ]]; \
	then \
		sudo apt-get install --assume-yes --no-install-suggests npm; \
		npm install -g npm; \
	fi

eslint:
	cd $(WPTD_PATH)webapp; npm install eslint babel-eslint eslint-plugin-html
	cd $(WPTD_PATH)webapp; npm run lint

dev_data:
	cd $(WPTD_GO_PATH)/util; go get -t ./...
	go run util/populate_dev_data.go $(FLAGS)

webapp_deploy_staging: env-BRANCH_NAME
	gcloud config set project wptdashboard
	gcloud auth activate-service-account --key-file $(WPTD_PATH)client-secret.json
	cd $(WPTD_PATH); util/deploy.sh -q -b $(BRANCH_NAME)

env-%:
	@ if [[ "${${*}}" = "" ]]; then echo "Environment variable $* not set"; exit 1; fi
