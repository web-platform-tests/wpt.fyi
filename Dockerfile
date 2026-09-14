# vim: set expandtab sw=4
# Go 1.27 is required for taskcluster >= v107.0.0
FROM golang:1.27.1-bookworm

# Create a non-priviledged user to run browsers as (Firefox and Chrome do not
# like to run as root).
RUN chmod a+rx $HOME && useradd --uid 9999 --user-group --create-home browser

# Add apt repositories for Java, Node.js and Google Cloud CLI
RUN export DISTRO_CODENAME=$(awk -F= '/^VERSION_CODENAME/{print$2}' /etc/os-release) && \
    echo "deb [signed-by=/usr/share/keyrings/corretto.gpg] https://apt.corretto.aws stable main" > /etc/apt/sources.list.d/corretto.list && \
    curl -s https://apt.corretto.aws/corretto.key | gpg --dearmor -o /usr/share/keyrings/corretto.gpg && \
    export NODE_VERSION="20.x" && \
    export ARCH=$(dpkg --print-architecture) && \
    echo "deb [arch=$ARCH signed-by=/usr/share/keyrings/nodesource.gpg] https://deb.nodesource.com/node_$NODE_VERSION nodistro main" > /etc/apt/sources.list.d/nodesource.list && \
    curl -s https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key | gpg --dearmor -o /usr/share/keyrings/nodesource.gpg && \
    echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk-$DISTRO_CODENAME main" > /etc/apt/sources.list.d/google-cloud-sdk.list && \
    curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg

# Sort the package names!
# firefox-esr: provides deps for Firefox (we don't use ESR directly)
# java-11-amazon-corretto-jdk: provides JDK/JRE to Selenium & gcloud SDK
# python-crcmod: native module to speed up CRC checksum in gsutil
RUN apt-get update -qqy && apt-get install -qqy --no-install-suggests \
        curl \
        firefox-esr \
        java-11-amazon-corretto-jdk \
        nodejs \
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

# We must stick to 527.0.0 because the datastore emulator only requires Java 11.
# Versions 528.0.0 and onwards require Java 21.
# However, we can't upgrade to a newer version of Java because when Chrome
# runs with WCT/Selenium, it only runs with java 11 and not 17 or 21.
ENV CLOUD_SDK_VERSION=527.0.0
# Google Cloud SDK configuration
# Installed from archive tarball because Debian apt repo pruned version 527.0.0.
RUN curl -sSL https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-${CLOUD_SDK_VERSION}-linux-x86_64.tar.gz | tar -xz -C /usr/local && \
    /usr/local/google-cloud-sdk/bin/gcloud components install --quiet \
        app-engine-python \
        app-engine-python-extras \
        app-engine-go \
        cloud-datastore-emulator \
        beta && \
    /usr/local/google-cloud-sdk/bin/gcloud config set core/disable_usage_reporting true && \
    /usr/local/google-cloud-sdk/bin/gcloud config set component_manager/disable_update_check true && \
    chmod -R a+rwX /usr/local/google-cloud-sdk && \
    ln -s /usr/local/google-cloud-sdk/bin/* /usr/bin/ && \
    gcloud --version
