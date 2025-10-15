# Production deployment spec for the webapp.

FROM golang:1.25.3-bookworm as builder

RUN apt-get update
RUN apt-get install -qy --no-install-suggests git sudo
WORKDIR /go/src/wpt.fyi

# Default to the existing path when running inside the development docker container.
ARG LOCAL_SRC=wpt.fyi
COPY $LOCAL_SRC .
RUN CGO_ENABLED=0 make go_build

# Application image.
FROM nginx:1.26.3-alpine3.20

# Copy your application binary
COPY --from=builder /go/src/wpt.fyi/web /usr/local/bin/app

# Copy static assets
COPY --from=builder /go/src/wpt.fyi/webapp/dynamic-components /usr/share/nginx/html/dynamic-components/
COPY --from=builder /go/src/wpt.fyi/webapp/components /usr/share/nginx/html/components/
COPY --from=builder /go/src/wpt.fyi/webapp/static /usr/share/nginx/html/static/
COPY --from=builder /go/src/wpt.fyi/webapp/templates /usr/share/nginx/html/templates/
COPY --from=builder /go/src/wpt.fyi/webapp/views /usr/share/nginx/html/views/
RUN chmod -R a+r /usr/share/nginx/html

# Copy the Nginx configuration (see below)
COPY nginx.conf /etc/nginx/nginx.conf

# create log dir configured in nginx.conf
RUN mkdir -p /var/log/app_engine

CMD ["/bin/sh",  "-c",  "nginx -g 'daemon off;' & PORT=8081 /usr/local/bin/app"]
