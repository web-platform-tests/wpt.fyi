# Production deployment spec for query cache service.

# Base golang 1.18 image.
FROM golang:1.18-buster as builder

RUN apt-get update
RUN apt-get install -qy --no-install-suggests git
WORKDIR /go/src/wpt.fyi

COPY wpt.fyi .
RUN /usr/local/go/bin/go build -o ../../bin/app ./api/query/cache/service

# For debugging, you can use the debug tag, aka gcr.io/distroless/static-debian11:debug
# Otherwise, it will produce a image without a shell.
# Learn more about distroless at https://github.com/GoogleContainerTools/distroless
FROM gcr.io/distroless/static-debian11

COPY --from=builder /go/bin/app /usr/local/bin/app

CMD ["/usr/local/bin/app"]
