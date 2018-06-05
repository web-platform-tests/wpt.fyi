FROM gcr.io/gcp-runtimes/go1-builder:1.10

#
# Dockerfile suitable for development and continuous integration of all wpt.fyi
# services. It contains an environment suitable for installing and running
# services using the project-level Makefile.
#
# See Dockerfiles in sub-directories for individual service deployments.
#
# Caveats:
# - AppEngine Standard uses golang 1.8, whereas AppEngine Flex defaults to
#   golang 1.10. This development environment uses the base image recommended
#   for AppEngine Flex custom golang runtime, hence golang 1.10 is the default
#   golang toolchain. However, when using the gcloud dev_appserver toolchain,
#   it will internally use a custom golang 1.8 environment.
#

USER root

# Expected layout: /home/user/web-platform-tests/{wpt.fyi,other_repos...}
ENV USER_HOME="/home/user"
ENV WPT_PATH="${USER_HOME}/web-platform-tests"
ENV WPTD_PATH="${WPT_PATH}/wpt.fyi"
ENV WPTD_OUT_PATH="${USER_HOME}/wptdout"

# Setup go environment
ENV GOPATH="${USER_HOME}/go"
RUN mkdir -p "${GOPATH}"
ENV GCLOUD_PATH="${USER_HOME}/google-cloud-sdk"
ENV WPTD_GO_PATH="${GOPATH}/src/github.com/web-platform-tests/wpt.fyi"

# Setup go + python binaries path
ENV PATH=$PATH:/usr/local/go/bin:$GOPATH/bin:${USER_HOME}/.local/bin:${GCLOUD_PATH}/bin

# Install sudo so that unpriv'd dev user can "sudo apt-get install ..." in from
# Makefile.
RUN apt-get update && apt-get install sudo make

# Put wpt.fyi code in GOPATH
RUN mkdir -p "${GOPATH}/src/github.com"
RUN ln -s "${WPT_PATH}" "${GOPATH}/src/github.com/web-platform-tests"

RUN mkdir -p "${WPT_PATH}"
RUN mkdir -p "${WPTD_PATH}"

# Drop dev environment into source path
WORKDIR "${WPTD_PATH}"
