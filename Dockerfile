# vim: set expandtab sw=4
FROM golang:1.22.7-bookworm

# Create a non-priviledged user to run browsers as (Firefox and Chrome do not
# like to run as root).
RUN chmod a+rx $HOME && useradd --uid 9999 --user-group --create-home browser

# Sort the package names!
# firefox-esr: provides deps for Firefox (we don't use ESR directly)
# openjdk-17-jdk: provides JDK/JRE to Selenium & gcloud SDK
# python-crcmod: native module to speed up CRC checksum in gsutil
RUN apt-get update -qqy && apt-get install -qqy --no-install-suggests \
        curl \
        firefox-esr \
        lsb-release \
        openjdk-17-jdk \
        python3.11 \
        python3-crcmod \
        sudo \
        tox \
        wget \
        xvfb && \
    rm /usr/bin/firefox

# The base golang image adds Go paths to PATH, which cannot be inherited in
# sudo by default because of the `secure_path` directive. Overwrite sudoers to
# discard the setting.
RUN echo "root ALL=(ALL:ALL) ALL" > /etc/sudoers

# Node LTS
RUN curl -sL https://deb.nodesource.com/setup_18.x | bash - && \
    apt-get install -qqy nodejs

# Google Cloud SDK
# Based on https://github.com/GoogleCloudPlatform/cloud-sdk-docker/blob/master/Dockerfile
RUN export CLOUD_SDK_REPO="cloud-sdk-$(lsb_release -c -s)" && \
    echo "deb https://packages.cloud.google.com/apt $CLOUD_SDK_REPO main" > /etc/apt/sources.list.d/google-cloud-sdk.list && \
    echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" > /etc/apt/sources.list.d/google-cloud-sdk.list && \
    curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg && \
    apt-get update -y && apt-get install -y \
        google-cloud-cli \
        google-cloud-cli-app-engine-python \
        google-cloud-cli-app-engine-python-extras \
        google-cloud-cli-app-engine-go \
        google-cloud-cli-datastore-emulator && \
    gcloud config set core/disable_usage_reporting true && \
    gcloud config set component_manager/disable_update_check true && \
    gcloud --version
