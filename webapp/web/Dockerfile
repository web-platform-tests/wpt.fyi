# Production deployment spec for webapp.

# BUILD_ENV represents where this docker image is built.
# Options:
# - gcloud - App Engine
# - local - Locally
ARG BUILD_ENV=gcloud

# GO_IMAGE represents the base go image to use.
# Tags: https://hub.docker.com/_/golang
# For more information, check out the upgrading-go.md document.
ARG GO_IMAGE=golang:1.18-buster

# Stage 1a - App Engine builder
# For the gcloud image, wpt.fyi was already created by the `package_service`
# Makefile target ran via ./util/deploy-staging.sh (which runs make deploy_staging -> make deployment_state).
FROM ${GO_IMAGE} as builder_gcloud
WORKDIR /webapp
ONBUILD COPY wpt.fyi .

# Stage 1b - Local builder
# For the local build, have to need to run `make package_service` manually first.
# This will create the wpt.fyi folder
# Afterwards copy the files into the webapp folder, just like the gcloud image
FROM ${GO_IMAGE} as builder_local
WORKDIR /webapp-temp
ONBUILD COPY . .
ONBUILD RUN make package_service APP_PATH=webapp/web/
ONBUILD WORKDIR /webapp
ONBUILD RUN cp -a /webapp-temp/webapp/web/wpt.fyi/. /webapp

# Stage 2 - Use the selected builder (1a or 1b), which is dependent upon the BUILD_ENV build arg
FROM builder_${BUILD_ENV} as builder

# Node LTS
RUN curl -sL https://deb.nodesource.com/setup_16.x | bash - && \
    apt-get install -qqy nodejs

WORKDIR /webapp
# Create a build
RUN make go_build

# Cleanup to make a minimal image
# Use git to help with filtering out the files
# By using the existing .gcloudignore as .gitignore, `git clean` can filter out the same files
RUN git init && \
    # Copy the .gcloudignore as a .gitignore
    cp /webapp/webapp/web/.gcloudignore .gitignore && cat .gitignore && \
    # Prior to running git clean, manually clean up some files no longer needed
    # but were needed during the push to App Engine (aka files that cannot be in .gcloudignore).
    # These are extra files that should removed because the next stage will copy the whole directory.
    # > Start manual clean up section
    # .go files - Now that it is built, can remove any Go files needed for building.
    find . -name "*.go" -type f -delete && \
    rm go.mod go.sum && \
    # js files needed for building
    rm webapp/package-lock.json webapp/package.json && \
    # Makefile - No longer need Makefile
    rm Makefile && \
    # webapp/.gitignore - Remove nested gitignore
    # If this file remains, `git clean` will delete the generated static/<browser>_64x64.png files
    rm webapp/.gitignore && \
    # > End manual clean up section
    # Add the files then do a clean
    git add .  && git clean -fdx && git status && \
    # Clean up the git metadata since we don't need it anymore
    rm -rf .git

# Stage 3
# For debugging, you can use the debug tag, aka gcr.io/distroless/static-debian11:debug
# Otherwise, it will produce a image without a shell.
# Learn more about distroless at https://github.com/GoogleContainerTools/distroless
FROM gcr.io/distroless/static-debian11:debug as webapp_release
WORKDIR /webapp
COPY --from=builder /webapp /webapp

EXPOSE 8080
ENTRYPOINT [ "/webapp/web" ]
