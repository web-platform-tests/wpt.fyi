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

BQ_LIB_REPO ?= github.com/GoogleCloudPlatform/protoc-gen-bq-schema
PB_LIB_DIR ?= ../protobuf/src
PB_BQ_LIB_DIR ?= $(WPTD_PATH)vendor/$(BQ_LIB_REPO)
PB_LOCAL_LIB_DIR ?= protos
PB_BQ_OUT_DIR ?= bq-schema
PB_PY_OUT_DIR ?= run/protos
PB_GO_OUT_DIR ?= generated
PB_GO_PKG_MAP ?= Mbq_table_name.proto=$(BQ_LIB_REPO)/protos

PROTOS=$(wildcard $(PB_LOCAL_LIB_DIR)/*.proto)

GO_FILES := $(wildcard $(WPTD_PATH)**/*.go)
GO_FILES := $(filter-out $(wildcard $(WPTD_PATH)generated/**/*.go), $(GO_FILES))
GO_FILES := $(filter-out $(wildcard $(WPTD_PATH)vendor/**/*.go), $(GO_FILES))

build: go_build

test: go_test

lint: go_lint eslint

prepush: build test lint

go_build: go_deps
	cd $(WPTD_GO_PATH); go build ./...

go_lint: go_deps
	cd $(WPTD_GO_PATH); golint -set_exit_status $(GO_FILES)
	# Print differences between current/gofmt'd output, check empty.
	cd $(WPTD_GO_PATH); ! gofmt -d $(GO_FILES) 2>&1 | read

go_test: go_deps
	cd $(WPTD_GO_PATH); go test -tags=medium -v ./...

go_deps: $(find .  -type f | grep '\.go$' | grep -v '\.pb.go$')
	cd $(WPTD_GO_PATH); go get -t ./...

eslint:
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
