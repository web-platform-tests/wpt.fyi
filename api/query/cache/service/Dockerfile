# Production deployment spec for query cache service.

FROM golang:1.22.10-bookworm as builder

RUN apt-get update
RUN apt-get install -qy --no-install-suggests git
WORKDIR /go/src/wpt.fyi

COPY wpt.fyi .
RUN CGO_ENABLED=0 /usr/local/go/bin/go build -o ../../bin/app ./api/query/cache/service

# Application image.
FROM gcr.io/distroless/static-debian11

COPY --from=builder /go/bin/app /usr/local/bin/app

CMD ["/usr/local/bin/app"]
