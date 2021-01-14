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

The `default` service is a standard App Engine service while the other two are
Flex.

## Deploy the app

### To production

First log into the `wptdashboard` project. You need to be a project member with
at least Editor privileges. Then from the project root directory:

```sh
git checkout main
git pull
make deploy_production PROJECT=wptdashboard APP_PATH=webapp/web
make deploy_production PROJECT=wptdashboard APP_PATH=results-processor
make deploy_production PROJECT=wptdashboard APP_PATH=api/query/cache/service
```

If you've updated [`index.yaml`](../webapp/web/index.yaml),
[`queue.yaml`](../webapp/web/queue.yaml), or
[`dispatch.yaml`](../webapp/web/dispatch.yaml) you must also deploy them manually.

```sh
cd webapp/web
gcloud app deploy --project=wptdashboard index.yaml queue.yaml dispatch.yaml
```

### To staging

([Travis](../.travis.yml) deploys all services automatically,but not
`index.yaml`, `queue.yaml` or `dispatch.yaml`.)

To deploy manually, follow the same instructions as production but replace
`wptdashboard` with `wptdashboard-staging`, and use `make deploy_staging`
instead of `make deploy_production`.

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
