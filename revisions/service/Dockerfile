# Production deployment spec for revisions announcer service.

# Base golang 1.10 image.
FROM gcr.io/gcp-runtimes/go1-builder:1.10 as builder

RUN apt-get update
RUN apt-get install -qy --no-install-suggests git
WORKDIR /go/src/app
COPY *.go .

RUN /usr/local/go/bin/go get -d ./...
RUN rm -rf /root/go/src/github.com/web-platform-tests/wpt.fyi/*
COPY wpt.fyi /root/go/src/github.com/web-platform-tests/wpt.fyi/
RUN /usr/local/go/bin/go build -o app .

# Application image.
FROM gcr.io/distroless/base:latest

COPY --from=builder /go/src/app/app /usr/local/bin/app

CMD ["/usr/local/bin/app"]
