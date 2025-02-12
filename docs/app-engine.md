# App Engine Documentation

The project runs on Google App Engine. It contains the following three services,
each of which has an `app.yaml` file in its directory and in some cases an
`app.staging.yaml` for the staging project.

1. **default**: `/webapp/web/`, the default service serves the `wpt.fyi` frontend and
   APIs.
2. **processor**: `/results-processor/`, the internal backend of the Results
   Receiver (not accessible externally) which processes the incoming results.
3. **searchcache**: `/api/query/cache/service/`, an in-memory cache and query
   executor for [structured searches](../api/query/README.md).

All services are App Engine Flex services.

## Deploy the app

### To production

You need to be a `wptdashboard` GCP project member with
at least Editor privileges. Then from the project root directory:

```sh
git checkout main
git pull
util/deploy-production.sh
```

Then follow the script’s instructions. You can see all available command line options by passing `-h`.

If there are changes to deploy but the checks have failed, it will provide the proper links to investigate the failures. If the failures should not block deployment (e.g. intermittent failures), it will suggest that you rerun the script with the `-f` flag to force deployment.

If the deployment fails during one of the build or deployment steps in the docker VM (for transient or intermittent errors), you can skip the GitHub bug creation and proceed straight to retrying the build by using the `-b` flag.

### To staging

([GitHub Actions](../.github/workflows/deploy.yml) deploys all services automatically, but not
`index.yaml`, `queue.yaml` or `dispatch.yaml`.)

To deploy manually, follow the same instructions as production but replace
`wptdashboard` with `wptdashboard-staging`, use `make deploy_staging`
instead of `make deploy_production` and use `app.staging.yaml` instead of `app.yaml`:

```sh
wptd_exec_it make deploy_staging PROJECT=wptdashboard-staging APP_PATH=webapp/web/app.staging.yaml
wptd_exec_it make deploy_staging PROJECT=wptdashboard-staging APP_PATH=results-processor/app.staging.yaml
wptd_exec_it make deploy_staging PROJECT=wptdashboard-staging APP_PATH=api/query/cache/service/app.staging.yaml
```

## Out-of-repo configurations

There are more configurations required in Google Cloud in addition to the YAML
files above. They need to be done using the `gcloud` CLI or on the GCP
dashboard, and are not currently version-controlled (TODO: consider using
Terraform).

### Serverless VPC Access

By default, App Engine **standard** environment is separated from the "internal"
(VPC) network (including Compute Engine and any resource with an internal IP
such as Cloud Memorystore). To connect to these resources, follow this doc to
enable Serverless VPC Access and configure the connector:
https://cloud.google.com/appengine/docs/standard/go/connecting-vpc (note: we do
not use "Shared VPC")

### Cloud Memorystore (Redis)

Follow this doc to set up Cloud Memorystore (Redis):
https://cloud.google.com/appengine/docs/standard/go/using-memorystore#setup_redis_db
