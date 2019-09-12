# vim: set expandtab sw=4
# WORKDIR is assumed to be a checkout of the repo (usually a volume).
FROM golang:1.12-buster

RUN mkdir -p ${GOPATH}/src/github.com/web-platform-tests && \
    ln -s $(pwd) ${GOPATH}/src/github.com/web-platform-tests/wpt.fyi

# Sort the package names!
# python-crcmod: native module to speed up CRC checksum in gsutil
# sudo: used in Makefile (no-op inside Docker)
RUN apt-get update -qqy && apt-get install -qqy \
        curl \
        lsb-release \
        python-crcmod \
        python3.7 \
        sudo \
        tox \
        wget

# Node LTS
RUN curl -sL https://deb.nodesource.com/setup_10.x | bash - && \
    apt-get install -qqy nodejs

# Google Cloud SDK
# Based on https://github.com/GoogleCloudPlatform/cloud-sdk-docker/blob/master/Dockerfile
RUN export CLOUD_SDK_REPO="cloud-sdk-$(lsb_release -c -s)" && \
    echo "deb https://packages.cloud.google.com/apt $CLOUD_SDK_REPO main" > /etc/apt/sources.list.d/google-cloud-sdk.list && \
    curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add - && \
    apt-get update -qqy && apt-get install -qqy \
        google-cloud-sdk \
        google-cloud-sdk-app-engine-python \
        google-cloud-sdk-app-engine-python-extras \
        google-cloud-sdk-app-engine-go && \
    gcloud config set core/disable_usage_reporting true && \
    gcloud config set component_manager/disable_update_check true && \
    gcloud --version

# Go tools (sort the lines!)
RUN go get -u \
    github.com/golang/mock/mockgen \
    golang.org/x/lint/golint
