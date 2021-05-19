# Triage Metadata Caching

The [wpt.fyi](https://wpt.fyi) dashboard has support for linking test results
for a specific test and browser to issues or bugs (or more generically, any
URL). The triaged data is stored in the
[wpt-metadata](https://github.com/web-platform-tests/wpt-metadata) repository
and is reflected back onto [wpt.fyi](https://wpt.fyi).

This section explains the caching mechanisms for triage data in webapp and
searchcache.

## webapp
Webapp hosts the following API endpoints:

- `/api/metadata` returns WPT Metadata to be displayed on the bottom of wpt.fyi pages and via inline icons on the test-result table

- `/api/metadata/triage` records triage information in the backend and sends PRs to the wpt-metadata repository which acts as the backing data store.

Both endpoints share the same in-memory copy of the wpt-metadata repository, and this copy expires [every 10 minutes](https://github.com/web-platform-tests/wpt.fyi/blob/9136dbf07414baf285c06787b6bf289632d27c83/api/metadata_cache.go#L29) in Redis. The only exception is that when users triage metadata, /api/metadata/triage will force an update of the copy.

- `/api/metadata/pending` retrieves pending metadata whose PRs are not merged yet in the wpt-metadata repository.

This endpoint stores pending metadata to Redis, with [a 7-day TTL](https://github.com/web-platform-tests/wpt.fyi/blob/9136dbf07414baf285c06787b6bf289632d27c83/api/metadata_handler.go#L82). If its pending PRs are merged or closed, pending metadata will be cleaned from Redis in [searchcache](https://github.com/web-platform-tests/wpt.fyi/blob/9136dbf07414baf285c06787b6bf289632d27c83/api/query/cache/poll/poll.go#L105).

## searchcache
Searchcache has a long-running polling thread that caches WPT Metadata [every 10 minutes](https://github.com/web-platform-tests/wpt.fyi/blob/207813b3ed18bae81068934caa478daffd782d36/api/query/cache/service/main.go#L151). When users search for triage information, the result is sometimes out-of-sync with webapp because the cache doesn't reflect what users have recently triaged.
