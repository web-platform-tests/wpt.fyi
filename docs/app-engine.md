# App Engine Documentation

The project runs on Google App Engine. It contains the following three services,
each of which has an `app.yaml` file in its directory.

1. **default**: `/webapp/`, the default service serves the `wpt.fyi` frontend and
   APIs.
2. **processor**: `/results-processor/`, the internal backend of the Results
   Receiver (not accessible externally) which processes the incoming results.
3. **searchcache**: `/api/query/cache/service/`, an in-memory cache and query
   executor for [structured searches](../api/query/README.md).

The `default` service is a standard AppEngine service while the other two are
Flex.

## Deploy the app

First log into the `wptdashboard` project. You need to be a project member with
at least Editor privileges. Then from the project root directory:

```sh
git checkout master
git pull
make deploy_production PROJECT=wptdashboard APP_PATH=webapp
make deploy_production PROJECT=wptdashboard APP_PATH=results-processor
make deploy_production PROJECT=wptdashboard APP_PATH=api/query/cache/service
```

If you've updated [`index.yaml`](../webapp/index.yaml),
[`queue.yaml`](../webapp/queue.yaml), or
[`dispatch.yaml`](../webapp/dispatch.yaml) you must also deploy them manually.

```sh
cd webapp
gcloud app deploy --project=wptdashboard index.yaml queue.yaml dispatch.yaml
```
