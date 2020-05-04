# Production deployment spec for query cache service.

# Base golang 1.12 image.
FROM gcr.io/gcp-runtimes/go1-builder:1.12 as builder

RUN apt-get update
RUN apt-get install -qy --no-install-suggests git
WORKDIR /go/src/wpt.fyi

COPY wpt.fyi .
RUN /usr/local/go/bin/go build -o ../../bin/app ./api/query/cache/service

# Application image.
FROM gcr.io/distroless/base:latest

COPY --from=builder /go/bin/app /usr/local/bin/app

CMD ["/usr/local/bin/app"]
