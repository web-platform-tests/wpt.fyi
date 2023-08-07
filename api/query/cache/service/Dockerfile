# Production deployment spec for query cache service.

# Base golang 1.20.1 image.
FROM golang:1.20.7-bullseye as builder

RUN apt-get update
RUN apt-get install -qy --no-install-suggests git
WORKDIR /go/src/wpt.fyi

COPY wpt.fyi .
RUN CGO_ENABLED=0 /usr/local/go/bin/go build -o ../../bin/app ./api/query/cache/service

# Application image.
FROM gcr.io/distroless/static-debian11

COPY --from=builder /go/bin/app /usr/local/bin/app

CMD ["/usr/local/bin/app"]
