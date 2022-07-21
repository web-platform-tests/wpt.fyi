# Maintenance: Upgrading Golang

This document details the files to change and the necessary steps when upgrading Golang. At the time of the writing, we only upgrade on minor version changes to Golang, not patch changes. If that changes, please update this document.

## Step 1 - Change the Runtime Version in webapp's app.staging.yaml and app.yaml files

Ensure that the desired version is available. Go to the [Golang standard runtime docs](https://cloud.google.com/appengine/docs/standard/go/runtime) to see the latest versions available.

Once you have confirmed that the desired version is available:
- Open [app.yaml](../webapp/web/app.yaml) and [app.staging.yaml](../webapp/web/app.staging.yaml)
- Change the `runtime` line to match the new version of Golang seen in the App Engine documentation.


## Step 2 - Change the version in the Dockerfiles
- tooling [Dockerfile](../Dockerfile) at the root of the repo
- searchcache [Dockerfile](../api/query/cache/service/Dockerfile)

The tooling image and the first stage of searchcache use the same Golang image. Check out the Golang [page](https://hub.docker.com/_/golang?tab=tags) on Docker Hub for the latest tags. Currently, we are using the `buster` [release](https://wiki.debian.org/DebianReleases) of Debian. As a result pick the `golang:<latest stable version>-buster` tag. If buster is superseded by a new version, you should change that as well.

## Step 3 - Change the version in go.mod

There is a line with the Golang version in the [go.mod](../go.mod) file. Change it to the latest major and minor version.

## Step 4 - Run go mod tidy

*Expand section for directions*
<!-- TODO add more instructions for setups like local and docker compose  -->

<details>
  <summary>Running "go mod tidy" in webplatformtests/wpt.fyi Docker container</summary>
  
  ### Terminal 1

  You need to run `go mod tidy` **but** `webplatformtests/wpt.fyi:latest` won't have the changes for the latest version of Golang from your changes to the tooling image in step 2. As a result, you will need to build the tooling image locally.

  ```sh
  # From the root of the repository
  docker build --tag webplatformtests/wpt.fyi:latest .
  ```

  Follow the steps in the main [README.md](../README.md) to start up the instance. It will use this locally built version of Dockerfile

  ### Terminal 2
  ```sh
  docker exec wptd-dev-instance go mod tidy
  ```
  This will update your go.mod and go.sum.

</details>