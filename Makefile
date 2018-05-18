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

# WPTD_PATH will have a trailing slash, e.g. /home/jenkins/wpt.fyi/
WPTD_PATH := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
WPTD_GO_PATH ?= $(GOPATH)/src/github.com/web-platform-tests/wpt.fyi
NODE_SELENIUM_PATH=$(WPTD_PATH)webapp/node_modules/selenium-standalone/.selenium/
SELENIUM_SERVER_PATH ?= $(NODE_SELENIUM_PATH)selenium-server/3.8.1-server.jar
GECKODRIVER_PATH ?= $(NODE_SELENIUM_PATH)geckodriver/0.20.0-x64-geckodriver
FIREFOX_PATH ?= $$(which firefox)
USE_FRAME_BUFFER ?= true
NVM_URL=https://raw.githubusercontent.com/creationix/nvm/v0.33.8/install.sh

BQ_LIB_REPO ?= github.com/GoogleCloudPlatform/protoc-gen-bq-schema
PB_LIB_DIR ?= ../protobuf/src
PB_BQ_LIB_DIR ?= $(WPTD_PATH)vendor/$(BQ_LIB_REPO)
PB_LOCAL_LIB_DIR ?= protos
PB_BQ_OUT_DIR ?= bq-schema
PB_PY_OUT_DIR ?= run/protos
PB_GO_OUT_DIR ?= generated
PB_GO_PKG_MAP ?= Mbq_table_name.proto=$(BQ_LIB_REPO)/protos

PROTOS=$(wildcard $(PB_LOCAL_LIB_DIR)/*.proto)

GO_FILES := $(shell find $(WPTD_PATH) -type f -name '*.go')
GO_FILES := $(filter-out $(wildcard $(WPTD_PATH)generated/**/*.go), $(GO_FILES))
GO_FILES := $(filter-out $(wildcard $(WPTD_PATH)vendor/**/*.go), $(GO_FILES))
GO_TEST_FILES := $(shell find $(WPTD_PATH) -type f -name '*_test.go')
GO_TEST_FILES := $(filter-out $(wildcard $(WPTD_PATH)generated/**/*.go), $(GO_TEST_FILES))
GO_TEST_FILES := $(filter-out $(wildcard $(WPTD_PATH)vendor/**/*.go), $(GO_TEST_FILES))

build: go_build bower_components

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

integration_test: go_webdriver_test web_components_test

go_webdriver_test: go_webdriver_deps
	cd $(WPTD_PATH)webdriver; go test -v -tags=large \
			--selenium_path=$(SELENIUM_SERVER_PATH) \
			--firefox_path=$(FIREFOX_PATH) \
			--geckodriver_path=$(GECKODRIVER_PATH) \
			--frame_buffer=$(USE_FRAME_BUFFER)

web_components_test: webdriver_deps web_component_tester
	cd $(WPTD_PATH)webapp; export DISPLAY=:99.0; npm test

go_webdriver_deps: go_deps webdriver_deps

webdriver_deps: bower_components xvfb web_component_tester

go_deps: $(find .  -type f | grep '\.go$' | grep -v '\.pb.go$')
	cd $(WPTD_GO_PATH); go get -t -tags="small medium large" ./...

eslint: eslint_node_modules
	cd $(WPTD_PATH)webapp; npm run lint

eslint_node_modules: node-babel-eslint node-eslint node-eslint-plugin-html

dev_data:
	cd $(WPTD_GO_PATH)/util; go get -t ./...
	go run util/populate_dev_data.go $(FLAGS)

webapp_deploy_staging: bower_components env-BRANCH_NAME
	gcloud config set project wptdashboard
	gcloud auth activate-service-account --key-file $(WPTD_PATH)client-secret.json
	cd $(WPTD_PATH); util/deploy.sh -q -b $(BRANCH_NAME)

web_component_tester: node-web-component-tester bower_components

bower_components: node-bower
	cd $(WPTD_PATH)webapp; npm run bower-components

node-%: node
	cd $(WPTD_PATH)webapp; if [[ "$$(node -p "require('$*/package.json').version")" == "" ]]; then npm install $*; fi

node: nvm
	if [[ "$$(which node)" == "" ]]; then source $$HOME/.nvm/nvm.sh; nvm install 6 && node --version; fi

nvm:
	if [[ ! -e $$HOME/.nvm/nvm.sh ]];	then wget -qO- $(NVM_URL) | bash;	fi

xvfb: node
	if [[ "$$(which Xvfb)" == "" ]]; then cd $(WPTD_PATH)webapp; npm install xvfb; fi

env-%:
	@ if [[ "${${*}}" = "" ]]; then echo "Environment variable $* not set"; exit 1; fi
