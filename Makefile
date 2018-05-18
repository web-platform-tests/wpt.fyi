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
WEBDRIVER_PATH ?= $(WPTD_GO_PATH)/webdriver
BROWSERS_PATH ?= $(HOME)/browsers
SELENIUM_PATH ?= $(BROWSERS_PATH)/selenium
FIREFOX_PATH ?= $(BROWSERS_PATH)/firefox/firefox
GECKODRIVER_PATH ?= $(BROWSERS_PATH)/geckodriver

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

go_webdriver_test: go_webdriver_deps bower_components
	cd $(WEBDRIVER_PATH); go test -v -tags=large \
			--selenium_path=$(SELENIUM_PATH) \
			--firefox_path=$(FIREFOX_PATH) \
			--geckodriver_path=$(GECKODRIVER_PATH)

go_webdriver_deps: go_deps webdriver_deps

webdriver_deps:
	cd $(WEBDRIVER_PATH); ./install.sh $(BROWSERS_PATH)

go_deps: $(find .  -type f | grep '\.go$' | grep -v '\.pb.go$')
	cd $(WPTD_GO_PATH); go get -t -tags="small medium large" ./...

eslint:
	cd $(WPTD_PATH)webapp; npm run lint

dev_data:
	cd $(WPTD_GO_PATH)/util; go get -t ./...
	go run util/populate_dev_data.go $(FLAGS)

deploy_staging: bower_components env-BRANCH_NAME env-APP_PATH
	gcloud config set project wptdashboard
	gcloud auth activate-service-account --key-file $(WPTD_PATH)client-secret.json
	cd $(WPTD_PATH); util/deploy.sh -q -b $(BRANCH_NAME) $(APP_PATH)

bower_components: bower
	cd $(WPTDPATH)webapp; npm run bower-components

bower:
	cd $(WPTDPATH)webapp; npm install bower

bower_components: bower
	cd $(WPTDPATH)webapp; npm run bower-components

bower:
	cd $(WPTDPATH)webapp; npm install bower

env-%:
	@ if [[ "${${*}}" = "" ]]; then echo "Environment variable $* not set"; exit 1; fi
