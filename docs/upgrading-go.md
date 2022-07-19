# Maintenance: Upgrading Go

This document details the files to change and the necessary steps when upgrading Go. At the time of the writing, we only upgrade on minor version changes to Go, not patch changes. If that changes, please update this document.

## Step 1 - Change the Runtime Version in webapp's app.staging.yaml and app.yaml files

Ensure that the desired version is available. Go to the [Go standard runtime docs](https://cloud.google.com/appengine/docs/standard/go/runtime). There you should see the latest versions available.

Once you have confirmed that:
- Open [app.yaml](../webapp/web/app.yaml) and [app.staging.yaml](../webapp/web/app.staging.yaml)
- Change the `runtime` line to match the new version of the Go.


## Step 2 - Change the version in the Dockerfiles
- tooling [Dockerfile](../Dockerfile) at the root of the repo
- searchcache [Dockerfile](../api/query/cache/service/Dockerfile)

The tooling image and the first stage of searchcache use the same golang image. Check out the golang [page](https://hub.docker.com/_/golang?tab=tags) on Docker Hub for the latest tags. Currently, we are using the `buster` [release](https://wiki.debian.org/DebianReleases) of Debian. As a result pick the `golang:<latest stable version>-buster` tag. If buster is superseded by a new version, you should change that as well.

## Step 3 - Change the version in go.mod

There is a line with the go version in the [go.mod](../go.mod) file. Change it to the latest major and minor version.

## Step 4 - Run go mod tidy

Now, we are in a chicken and egg problem. We need to run `go mod tidy` but `webplatformtests/wpt.fyi:latest` won't have the changes for the latest version of go from your changes to the tooling image in step 1. As a result, you will need to build the tooling image locally and start it.

```sh
# From the root of the repository
docker build --tag webplatformtests/wpt.fyi:latest .
```

*Expand section for directions*
<!-- TODO add more instructions for setups like local and docker compose  -->

<details>
  <summary>Docker</summary>
  
  # Step 1 - Terminal 1
  Follow the steps in the main [README.md](../README.md) to start up the instance. It will use this locally built version of Dockerfile

  # Step 2 - Terminal 2
  ```sh
  docker exec wptd-dev-instance go mod tidy
  ```
  This will update your go.mod and go.sum.

</details>