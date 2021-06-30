# Docker

We use Docker for two purposes: development and production. And we have a few
different Docker images.

## Development

The root [`Dockerfile`](../Dockerfile) is the image we use for [local
development](../README.md#development) and CI testing on GitHub
Actions. We have a [cron
job](https://github.com/web-platform-tests/wpt.fyi/actions?query=workflow%3A%22Update+Docker+image%22)
that rebuilds the image and pushes it to Docker Hub weekly (or whenever
`Dockerfile` changes) so that CI can pull the image directly instead of building
from scratch.

This image is big as it contains many development tools (e.g. the full `gcloud`
SDK, browsers to run WebDriver tests), so it is not suitable to be deployed.

## Production

All three AppEngine [services](app-engine.md) run in Docker containers, but we
only have `Dockerfile`s for the two Flex services (the standard runtime provides
a transparent container automatically):

* [processor](../results-processor/Dockerfile): Python with `gcloud` SDK (for
  `gsutil`)
* [searchcache](../api/query/cache/service/Dockerfile): Golang using the
  [builder pattern](https://docs.docker.com/develop/develop-images/multistage-build/)

These images are built as part of the deployment process (`gcloud app deploy`)
by Google Cloud Build. They are minimal images with few system tools.
