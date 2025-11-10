# Production deployment spec for query cache service.

FROM golang:1.25.4-bookworm as builder

RUN apt-get update
RUN apt-get install -qy --no-install-suggests git
WORKDIR /go/src/wpt.fyi

# Default to the existing path when running inside the development docker container.
ARG LOCAL_SRC=wpt.fyi
COPY $LOCAL_SRC .
RUN CGO_ENABLED=0 /usr/local/go/bin/go build -o ../../bin/app ./api/query/cache/service

# Application image.
FROM gcr.io/distroless/static-debian12

COPY --from=builder /go/bin/app /usr/local/bin/app

CMD ["/usr/local/bin/app"]
