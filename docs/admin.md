# Admin

## Flags

[`/admin/flags`](https://wpt.fyi/admin/flags) allows site-wide settings to be set, similar to the per-user [`/flags`](https://wpt.fyi/flags).

## Flushing caches

[`/admin/cache/flush`](https://wpt.fyi//admin/cache/flush) flushes some caches used by the webapp. Note that this does not affect the [searchcache](https://github.com/web-platform-tests/wpt.fyi/blob/master/api/query/cache/README.md), which runs as a separate service.

## Uploading results

[`/admin/results/upload`](https://wpt.fyi//admin/results/upload) is a form for upload results manually. Please file an issue on this project if you require credentials to upload results manually or via the API.
