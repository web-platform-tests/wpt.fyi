FROM gcr.io/google-appengine/golang

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

ENV USER_HOME="/home/user"
ENV WPTD_PATH="${USER_HOME}/wpt.fyi"
ENV WPT_PATH="${USER_HOME}/web-platform-tests"
ENV WPTD_OUT_PATH="${USER_HOME}/wptdout"

# Setup go environment
RUN mkdir -p "${USER_HOME}/go"
ENV GOPATH="${USER_HOME}/go"
ENV WPTD_GO_PATH="${GOPATH}/src/github.com/web-platform-tests/wpt.fyi"

# Setup go + python binaries path
ENV PATH=${USER_HOME}/google-cloud-sdk/bin:$PATH:/usr/local/go/bin:$GOPATH/bin:${USER_HOME}/.local/bin

# Install sudo so that unpriv'd dev user can "sudo apt-get install ..." in from
# Makefile.
RUN apt-get update && apt-get install sudo make

# Put wpt.fyi code in GOPATH
RUN mkdir -p "${GOPATH}/src/github.com/web-platform-tests"
RUN ln -s "${WPTD_PATH}" "${GOPATH}/src/github.com/web-platform-tests/wpt.fyi"

RUN mkdir -p "${WPTD_PATH}"
RUN mkdir -p "${WPT_PATH}"

# Drop dev environment into source path
WORKDIR "${WPTD_PATH}"
