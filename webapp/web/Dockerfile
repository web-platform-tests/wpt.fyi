# Production deployment spec for the webapp.

FROM golang:1.22.11-bookworm as builder

RUN apt-get update
RUN apt-get install -qy --no-install-suggests git sudo
WORKDIR /go/src/wpt.fyi

# Default to the existing path when running inside the development docker container.
ARG LOCAL_SRC=wpt.fyi
COPY $LOCAL_SRC .
RUN make go_build

# Application image.
FROM gcr.io/distroless/static-debian11

COPY --from=builder /go/src/wpt.fyi/webapp/web /usr/local/bin/app

CMD ["/usr/local/bin/app"]
