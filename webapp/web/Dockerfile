# Production deployment spec for webapp.
FROM golang:1.18-buster as builder

# Node LTS
RUN curl -sL https://deb.nodesource.com/setup_16.x | bash - && \
    apt-get install -qqy nodejs

WORKDIR /webapp
# Create a build
COPY . .
RUN make go_build

# Cleanup to make a minimal image
# Use git to help with filtering out the files
# By using the existing .gcloudignore as .gitignore we can filter out the same files
RUN git init && \
    # Copy the .gcloudignore as a .gitignore
    cp /webapp/webapp/web/.gcloudignore .gitignore && \
    cat .gitignore && echo "new" && \
    # > Start exception section.
    # These are extra files that should removed
    # .go files
    find . -name "*.go" -type f -delete && \
    rm go.mod go.sum && \
    # Makefile
    rm Makefile && \
    # webapp gitignore which would exclude the static/<browser>_64x64.png
    rm webapp/.gitignore && \
    # > End exception section
    # Add the files then do a clean
    # git ls-files --ignored --exclude-standard && echo "pause" && \
    git add .  && git clean -fdx && git status && \
    # Clean up the git metadata since we don't need it anymore
    rm -rf .git

# For debugging, you can use the debug tag, aka gcr.io/distroless/static-debian11:debug
# Otherwise, it will produce a image without a shell.
# Learn more about distroless at https://github.com/GoogleContainerTools/distroless
FROM gcr.io/distroless/static-debian11 as webapp_release
WORKDIR /webapp
COPY --from=webapp_build /webapp /webapp

EXPOSE 8080
ENTRYPOINT [ "/webapp/web" ]