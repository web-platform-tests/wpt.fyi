# Choose the same base Debian release as google/cloud-sdk.
FROM golang:1.12-stretch AS golang

FROM google/cloud-sdk:latest
COPY --from=golang /usr/local/go /usr/local/go
COPY --from=golang /go /go
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

# Keep package names sorted!
RUN apt-get update -qqy && apt-get install -qqy --no-install-recommends \
	g++ \
	gcc \
	make \
	pkg-config \
	sudo \
	tox

# Node LTS
RUN curl -sL https://deb.nodesource.com/setup_10.x | bash - && \
	apt-get install -qqy nodejs

# Go tools (sort the lines!)
RUN go get -u \
	github.com/golang/mock/mockgen \
	golang.org/x/lint/golint
