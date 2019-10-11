# vim: set expandtab sw=4
FROM golang:1.12-buster

# Sort the package names!
# openjdk-11-jdk: provides JDK/JRE to Selenium & gcloud SDK
# python-crcmod: native module to speed up CRC checksum in gsutil
# sudo: used in Makefile (no-op inside Docker)
RUN apt-get update -qqy && apt-get install -qqy --no-install-suggests \
        curl \
        equivs \
        lsb-release \
        openjdk-11-jdk \
        python-crcmod \
        python3.7 \
        sudo \
        tox \
        wget \
        xvfb

# Provide a fake openjdk-8-jdk for google-cloud-sdk-datastore-eumlator
# google-cloud-sdk-datastore-emulator incorrectly depends on openjdk-8-jdk which does not exist in buster.
RUN echo "Package: openjdk-8-jdk-dummy\nProvides: openjdk-8-jdk\nDepends: openjdk-11-jdk" > dummy.control && \
    cat dummy.control && \
    equivs-build dummy.control && \
    ls && \
    dpkg -i openjdk-8-jdk-dummy_1.0_all.deb

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
        google-cloud-sdk-app-engine-go \
        google-cloud-sdk-datastore-emulator && \
    gcloud config set core/disable_usage_reporting true && \
    gcloud config set component_manager/disable_update_check true && \
    gcloud --version
