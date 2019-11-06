# Searchcache

This directory contains the implementation of the searchcache for wpt.fyi. The
searchcache runs as a separate service from the main dashboard, providing a
caching layer between the dashboard and the backing test runs datastore for
complex search queries.

## Developing locally

It is possible to run the searchcache locally when developing, though it is
difficult to also have it talk to a local build of the dashboard. However
running the searchcache and then communicating with it via `curl` serves most
development needs.

### Prerequisites

Outside of the standard setup for developing wpt.fyi (see the main
CONTRIBUTING.md file), searchcache development requires:

1. At least 4GB of RAM (8GB+ preferable). The searchcache loves RAM more than
   Chrome does.
1. A GCP [service account](https://cloud.google.com/iam/docs/understanding-service-accounts)
   with the role Cloud Datastore User for the `wptdashboard-staging` project.

### Building

From the `service/` subdirectory, run:

```sh
go build
```

This should create a new binary, `service`.

### Running

**Do not run the searchcache binary without flags!**. It is configured to grab
pretty much all of the RAM on the machine it runs on - great for a VM, bad for
your development machine.

The following are required flags when running `service`:

* `--project_id=wptdashboard-staging` - set the project to the staging instance
  of [wpt.fyi](https://wpt.fyi).
* `--gcp_credentials_file=gcp.json` - provide the necessary credentials to
  access the datastore. See Prerequisites above.
* `--max_heap_bytes=256000000` - sets a soft limit on memory usage. Emphasis on
  'soft' - with a 256MB heap set here, the author had their searchcache
  stabilize at approximately 2 **GB** of memory.

Putting it together:

```sh
./service \
    --project_id=wptdashboard-staging \
    --gcp_credentials_file=gcp.json \
    --max_heap_bytes=256000000
```

Other flags can be found by passing `--help` to the `service` binary.

**Tip**: The `service` binary is quite noisy (as it slurps up test runs from
the datastore); it can be useful to pipe its output via `tee` and store the log
locally to let you easily grep through it.

### Interacting with searchcache

By default, searchcache is hosted on `localhost:8080`. It can be communicated
with using HTTP requests. For example, using `curl`:

```sh
curl -H "Content-Type: application/json" \
    -X POST \
    -d '{"run_ids":[267810084, 255750007],"query":{"exists":[{"is":"different"}]}}' \
    http://localhost:8080/api/search/cache
```
