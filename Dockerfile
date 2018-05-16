FROM gcr.io/google-appengine/golang

#
# Dockerfile suitable for development and continuous integration of all wpt.fyi
# services. It contains the union of all technologies that services (and their
# development environments) require.
#
# See Dockerfiles in sub-directories for individual service deployments.
#
# Caveats:
# - AppEngine Standard uses golang 1.8, whereas AppEngine Flex defaults to 
#   golang 1.10. This development environment uses the base image recommended 
#   for AppEngine Flex custom golang runtime, hence golang 1.10.
#

USER root

ENV USER_HOME="/home/user"
ENV WPTD_PATH="${USER_HOME}/wpt.fyi"
ENV WPT_PATH="${USER_HOME}/web-platform-tests"
ENV WPTD_OUT_PATH="${USER_HOME}/wptdout"

RUN apt-get update

# Install git, python-pip, virtualenv and unzip for setup below
RUN apt-get install --assume-yes --no-install-suggests \
    --no-install-recommends \
    git \
    make \
    python \
    python-pip \
    python-wheel \
    python-setuptools \
    virtualenv \
    unzip \
    dtrx

# Remove unwanted pre-installed Python packages
RUN apt-get remove --assume-yes \
    bzr \
    python-bzrlib \
    python-configobj \
    python-six \
    mercurial \
    mercurial-common

# Setup go environment
RUN mkdir -p "${USER_HOME}/go"
ENV GOPATH="${USER_HOME}/go"
ENV WPTD_GO_PATH="${GOPATH}/src/github.com/web-platform-tests/wpt.fyi"

# Setup go + python binaries path
ENV PATH=/opt/google-cloud-sdk/bin:$PATH:/usr/local/go/bin:$GOPATH/bin:${USER_HOME}/.local/bin

# Install go dependencies. Manual git clone + install is a workaround for #85.
ENV GOLANG_ORG_PATH="${GOPATH}/src/golang.org/x"
RUN mkdir -p ${GOLANG_ORG_PATH}
RUN cd ${GOLANG_ORG_PATH} && git clone https://github.com/golang/lint
RUN cd ${GOLANG_ORG_PATH}/lint && go get ./... && go install ./...

# Install curl & gcloud SDK.
RUN apt-get update -q
RUN apt-get install -qy curl
RUN curl -s https://sdk.cloud.google.com > install-gcloud.sh
RUN bash install-gcloud.sh --disable-prompts --install-dir=/opt
ENV PATH=/opt/google-cloud-sdk/bin:$PATH
RUN gcloud config set disable_usage_reporting false

# Update the SDK to the latest.
RUN gcloud components update

# Ensure Google Cloud Platform tools are installed
RUN gcloud components install \
    app-engine-go \
    bq \
    core \
    gsutil \
    app-engine-python

# Install npm, and use a user-relative global path.
RUN curl -sL https://deb.nodesource.com/setup_8.x | bash
RUN apt-get install -y nodejs

# Install wct
# Note that --unsafe-perm bypasses a post install script issue.
# See https://github.com/npm/npm/issues/17346
RUN npm install -g web-component-tester --unsafe-perm
RUN npm install -g eslint babel-eslint eslint-plugin-html

# Put wpt.fyi code in GOPATH
RUN mkdir -p "${GOPATH}/src/github.com/web-platform-tests"
RUN ln -s "${WPTD_PATH}" "${GOPATH}/src/github.com/web-platform-tests/wpt.fyi"

RUN mkdir -p "${WPTD_PATH}"
RUN mkdir -p "${WPT_PATH}"

# Drop dev environment into source path
WORKDIR "${WPTD_PATH}"
