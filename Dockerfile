# vim: set expandtab sw=4
# Go 1.27 is required for taskcluster >= v107.0.0
FROM golang:1.27.3-bookworm

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
# java-21-amazon-corretto-jdk: provides the JRE that the Cloud Datastore emulator needs
# python-crcmod: native module to speed up CRC checksum in gsutil
RUN apt-get update -qqy && apt-get upgrade -qqy && apt-get install -qqy --no-install-suggests \
        curl \
        firefox-esr \
        java-21-amazon-corretto-jdk \
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

# Google Cloud SDK configuration: the latest release, from the archive tarball.
# The Cloud Datastore emulator in releases from 528.0.0 on needs Java 21+.
RUN curl -sSL https://dl.google.com/dl/cloudsdk/channels/rapid/google-cloud-sdk.tar.gz | tar -xz -C /usr/local && \
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
