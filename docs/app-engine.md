# App Engine Documentation

The server serving [wpt.fyi](https://wpt.fyi) is an App Engine app.
Its entry point is [`main.go`](../main.go) and it is configured by
[`app.yaml`](../app.yaml).

## Deploy the app

First log into the `wpt.fyi` project. You need to be a project member
with at least Editor privileges.

```sh
gcloud init
```

Make sure you have the latest code and deploy the app.

```sh
git pull
./util/deploy.sh
```

If you've updated [`index.yaml`](../index.yaml) you must also deploy the indexes manually.

```sh
git pull
cd webapp
gcloud app deploy index.yaml
```
