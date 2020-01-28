# wpt.fyi API

This package defines and implements HTTP API endpoints for [wpt.fyi](https://wpt.fyi/), and this
document covers usage and parameters of those endpoints.

## Resource endpoints

Here's a list of endpoints to query various resources. An exhaustive list of
the endpoints can be found in `routes.go`.

 - [/api/runs](#apiruns)
 - [/api/runs/{id}](#apirunsid)
 - [/api/run](#apirun)
 - [/api/shas](#apishas)
 - [/api/diff](#apidiff)
 - [/api/results](#apiresults)
 - [/api/interop](#apiinterop)
 - [/api/manifest](#apimanifest)
 - [/api/search](#apisearch)
 - [/api/metadata](#apimetadata)

Also see [results creation](#results-creation) for endpoints to add new data.

## TestRun entities

`TestRun` entities represent metadata about an execution of the [wpt](https://github.com/web-platform-tests/wpt) test suite, on a particular product. Tests are run on a regular basis, and each entry in `/api/runs` annotates when the tests were executed, which product they were executed on, and the where the results are stored.

### /api/runs

Gets the TestRun metadata for all runs for a given SHA[0:10], sorted by `time_start` descending.

__Parameters__

__`sha`__ : SHA[0:10] of the runs to get, or the keyword `latest`. Defaults to `latest`.

__`product`__ : Product(s) to include (repeated param), e.g. `chrome` or `firefox-60`.

__`aligned`__ : boolean for whether to get only SHAs which were executed across all of the requested `product`s.

__`labels`__: A comma-separated list of labels, e.g. `firefox,stable`; only runs with all
the given labels will be returned. There are currently two kinds of labels supported,
browser names (`chrome`, `edge`, `firefox`, `safari`) and release channels (`experimental`
or `stable`).

__`from`__ : RFC3339 timestamp, for which to include runs that occured after the given time.
NOTE: Runs are sorted by `time_start` descending, so be wary when combining this parameter
with the `max-count` parameter below.

__`to`__ : RFC3339 timestamp, for which to include runs that occured before the given time.

__`max-count`__ : Maximum number of runs to get (for each browser). Maximum of 500.

#### staging.wpt.fyi only (Beta params)

__`pr`__ (Beta): GitHub PR number. Shows runs for commits that belong to the PR.

#### Examples

- https://wpt.fyi/api/runs?product=chrome&product=safari
- https://wpt.fyi/api/runs?product=chrome&from=2018-01-01T00:00:00Z&max-count=10

<details><summary><b>Example JSON</b></summary>

    [
      {
        "browser_name": "chrome",
        "browser_version": "67.0.3396.62",
        "os_name": "linux",
        "os_version": "4.4",
        "revision": "2bd11b91d4",
        "full_revision_hash": "2bd11b91d490ddd5237bcb6d8149a7f25faaa101",
        "results_url": "https://storage.googleapis.com/wptd/2bd11b91d4/chrome-stable-linux-summary.json.gz",
        "created_at": "2018-06-05T08:27:30.627865Z",
        "raw_results_url": "https://storage.googleapis.com/wptd-results/2bd11b91d490ddd5237bcb6d8149a7f25faaa101/chrome_67.0.3396.62_linux_4.4/report.json"
      }
    ]

</details>

### /api/runs/{id}

Gets a specific (single) TestRun metadata by its datastore ID.

#### Example

https://wpt.fyi/api/runs/5164888561287168

<details><summary><b>Example JSON</b></summary>

    {
      "id": "5164888561287168",
      "browser_name": "chrome",
      "browser_version": "67.0.3396.62",
      "os_name": "linux",
      "os_version": "4.4",
      "revision": "2bd11b91d4",
      "full_revision_hash": "2bd11b91d490ddd5237bcb6d8149a7f25faaa101",
      "results_url": "https://storage.googleapis.com/wptd/2bd11b91d4/chrome-stable-linux-summary.json.gz",
      "created_at": "2018-06-05T08:27:30.627865Z",
      "raw_results_url": "https://storage.googleapis.com/wptd-results/2bd11b91d490ddd5237bcb6d8149a7f25faaa101/chrome_67.0.3396.62_linux_4.4/report.json"
    }

</details>

### /api/run

Gets a specific (single) TestRun metadata by `product` and `sha`.

__Parameters__

__`sha`__ :  SHA[0:10] of the runs to get, or the keyword `latest`. Defaults to `latest`.

__`product`__ : browser[version[os[version]]]. e.g. `chrome-63.0-linux`

#### Example

https://wpt.fyi/api/run?sha=latest&product=chrome

<details><summary><b>Example JSON</b></summary>

    {
      "id": "5164888561287168",
      "browser_name": "chrome",
      "browser_version": "67.0.3396.62",
      "os_name": "linux",
      "os_version": "4.4",
      "revision": "2bd11b91d4",
      "full_revision_hash": "2bd11b91d490ddd5237bcb6d8149a7f25faaa101",
      "results_url": "https://storage.googleapis.com/wptd/2bd11b91d4/chrome-stable-linux-summary.json.gz",
      "created_at": "2018-06-05T08:27:30.627865Z",
      "raw_results_url": "https://storage.googleapis.com/wptd-results/2bd11b91d490ddd5237bcb6d8149a7f25faaa101/chrome_67.0.3396.62_linux_4.4/report.json"
    }

</details>

### /api/shas

Gets an array of revisions (SHA[0:10]), in reverse chronological order.
This method behaves similarly to [/api/runs](#apiruns) above, but projects the `revision` field's value.

__Parameters__

__`aligned`__ : boolean for whether to get only SHAs which were executed across all of the requested `product`s.

__`product`__ : Product(s) to include (repeated param), e.g. `chrome` or `firefox-60`

__`from`__ : RFC3339 timestamp, for which to include runs that occured after the given time.
NOTE: Runs are sorted by `time_start` descending, so be wary when combining this parameter
with the `max-count` parameter below.

__`to`__ : RFC3339 timestamp, for which to include runs that occured before the given time.

__`max-count`__ : Maximum number of runs to get (for each browser). Maximum of 500.

#### Example

https://wpt.fyi/api/shas?product=chrome

<details><summary><b>Example JSON</b></summary>

    [
      "98530fb944",
      "2bd11b91d4"//, ...
    ]

</details>

## Results summaries

The following methods apply to the results summaries JSON blobs, which are linked to from
[TestRun entities](#test-run-entities).

### /api/results

Performs an HTTP redirect for the results summary JSON blob of the given TestRun.

__Response format__

The summary JSON is in the format

    {
      "/path/to/test.html": [1, 1],
    }

Where the array contains [`number of passes`, `total tests`].

__Parameters__

__`product`__ : Product to fetch the results for, e.g. `chrome-66`

__`sha`__ : SHA[0:10] of the TestRun to fetch, or the keyword `latest`. Defaults to `latest`.

#### Example

https://wpt.fyi/api/results?product=chrome

<details><summary><b>Example JSON</b> (from the summary.json.gz output)</summary>

    {
      "/css/css-text/i18n/css3-text-line-break-opclns-213.html": [1, 1],
      "/css/css-writing-modes/table-progression-vrl-001.html": [1, 1],
      // ...
    }

</details>

### /api/diff

Computes a summary JSON blob of the differences between two TestRun summary blobs,
in the format of an array of [improved, regressed, total-delta].

__Parameters__

__`before`__ : [product]@[sha] spec for the TestRun to use as the before state.

__`after`__ : [product]@[sha] spec for the TestRun to use as the after state.

__`path`__ : Test path to filter by. `path` is a repeatable query parameter.

__`filter`__ : Differences to include in the summary.
 - `A` : Added - tests which are present after, but not before.
 - `D` : Deleted - tests which are present before, but not after.
 - `C` : Changed - tests which are present before and after, but the results summary is different.
 - `U` : Unchanged - tests which are present before and after, and the results summary count is not different.

### /api/interop

Gets TestRun interoperability metadata, for the runs that would be fetched
using the same params calling [/api/runs](#apiruns).

Note that if a subset of browsers are selected, the most recent interoperability
metadata that includes all of the browsers is return (which may have been
computed from more than the returned browsers). For example,
`/api/interop?product=chrome-67` will return interoperability metadata that
includes the results from the latest run of Chrome 67.0.

__Parameters__

__`sha`__ : SHA[0:10] of the runs to get, or the keyword `latest`. Defaults to `latest`.

__`product`__ : Product(s) to include (repeated param), e.g. `chrome` or `firefox-60`.

__`labels`__: A comma-separated list of labels, e.g. `firefox,stable`; only runs with all
the given labels will be returned. There are currently two kinds of labels supported,
browser names (`chrome`, `edge`, `firefox`, `safari`) and release channels (`experimental`
or `stable`).

__`from`__ : RFC3339 timestamp, for which to include runs that occured after the given time.
NOTE: Runs are sorted by `time_start` descending, so be wary when combining this parameter
with the `max-count` parameter below.

__`to`__: RFC3339 timestamp, for which to include runs that occured before the given time.

__`max-count`__ : Maximum number of runs to get (for each browser). Maximum of 500.

#### Examples

- https://wpt.fyi/api/interop
- https://wpt.fyi/api/interop?product=chrome-67
- https://wpt.fyi/api/interop?label=experimental

<details><summary><b>Example JSON</b></summary>

    {
      "test_runs": [
        {
          "id": 4829365045035008,
          "browser_name": "chrome",
          "browser_version": "69.0.3472.3 dev",
          "os_name": "linux",
          "os_version": "16.04",
          "revision": "9f00a60d91",
          "full_revision_hash": "9f00a60d91ba84e52dac35d6e08da2050774811d",
          "results_url": "https://storage.googleapis.com/wptd-staging/9f00a60d91ba84e52dac35d6e08da2050774811d/chrome-69.0.3472.3_dev-linux-16.04-904a25b130-summary.json.gz",
          "created_at": "2018-07-06T15:58:24.377035Z",
          "raw_results_url": "https://storage.googleapis.com/wptd-results-staging/9f00a60d91ba84e52dac35d6e08da2050774811d/chrome-69.0.3472.3_dev-linux-16.04-904a25b130/report.json",
          "labels": ["buildbot", "chrome", "experimental"]
        } //, ...
      ],
      "start_time": "2018-07-06T18:42:27.478781Z",
      "end_time": "2018-07-06T18:42:36.658149Z",
      "url": "https://storage.googleapis.com/wptd-metrics-staging/1530902547-1530902556/pass-rates.json.gz"
    }

</details>

## Test Manifest

The following methods apply to the retrieval and filtering of the Test Manifest in [WPT](https://github.com/web-platform-tests/wpt),
which contains metadata about each test.

### /api/manifest

Gets the JSON of the WPT manifest GitHub release asset, for a given `sha` (defaults to latest).

__Parameters__

__`sha`__ : SHA of the [WPT](https://github.com/web-platform-tests/wpt) repo PR for which to fetch,
    the manifest, or the keyword `latest`. (Defaults to `latest`.)

NOTE: The full SHA of the fetched manifest is returned in the HTTP response header `X-WPT-SHA`, e.g.

    X-WPT-SHA: abcdef0123456789abcdef0123456789abcdef01

__Response format__

The high-level structure of the `v4` manifest is as follows:

    {
      "items": {
        "manual": {
            "file/path": [
              manifest_item,
              ...
            ],
            ...
        },
        "reftest": {...},
        "testharness": {...},
        "visual", {...},
        "wdspec": {...},
      },
    }

`manifest_item` is an **array** (nested in the map's `"file/path"` value's array) with varying contents. Loosely,

- For `testharness` entries: `[url, extras]`
  - `extras` example: `{"timeout": "long", "testdriver": True}`
- For `reftest` entries: `[url, references, extras]`
  - `references` example: `[[reference_url1, "=="], [reference_url2, "!="], ...]`
  - `extras` example: `{"timeout": "long", "viewport_size": ..., "dpi": ...}`

## Results creation

### /api/results/upload

Uploads a wptreport to the dashboard to create the test run.

This endpoint only accepts POST requests. Requests need to be authenticated via HTTP basic auth.
Please contact [Ecosystem Infra](mailto:ecosystem-infra@chromium.org) if you want to register as a
"test runner", to upload results.

#### File payload

__Content type__: `multipart/form-data`

__Parameters__

__`labels`__: (Optional) A comma-separated string of labels for this test run. Currently recognized
labels are "experimental" and "stable" (the release channel of the tested browser).

__`callback_url`__: (Optional) A URL that the processor should `POST` when successful, which will
create the TestRun. Defaults to /api/results/create in the current project's environment (e.g. wpt.fyi for
wptdashboard, staging.wpt.fyi for wptdashboard-staging).

__`result_file`__: A **gzipped** JSON file produced by `wpt run --log-wptreport`.
This field can be repeated to include multiple files (for chunked reports).

__`screenshot_file`__: A **gzipped** screenshot database produced by `wpt run --log-screenshot`.
This field can be repeated to include multiple links (for chunked reports).

The JSON file roughly looks like this:

```json
{
  "results": [...],
  "time_start": MILLISECONDS_SINCE_EPOCH,
  "time_end": MILLISECONDS_SINCE_EPOCH,
  "run_info": {
    "revision": "WPT revision of the test run",
    "product": "your browser",
    "browser_version": "version of the browser",
    "os": "your os",
    "os_version": "OPTIONAL OS version",
    ...
  }
}
```

__Notes__

The `time_start` and `time_end` fields are numerical timestamps (in milliseconds since the UNIX epoch)
when the whole test run starts and finishes. They are optional, but encouraged. `wpt run` produces
them in the report by default.

`run_info.{revision,product,browser_version,os}` are required, and should be automatically
generated by `wpt run`. If for some reason the report does not contain these fields (e.g. old WPT
version, Sauce Labs, or custom runners), they can be overridden with the following *optional*
parameters in the POST payload (this is __NOT__ recommended; please include metadata in the reports
whenever possible):

* __`revision`__ (note this should be the full revision hash, not a 10-char truncation)
* __`browser_name`__ (note that it is not called `product` here)
* __`browser_version`__
* __`os_name`__ (note that it is not called `os` here)
* __`os_version`__

#### URL payload

__Content type__: `application/x-www-form-urlencoded`

__Parameters__

__`result_url`__: A URL to a **gzipped** JSON file produced by `wpt run --log-wptreport` (see above
for its format). This field can be repeated to include multiple links (for chunked reports).

__`screenshot_url`__: A URL to a **gzipped** screenshot database produced by `wpt run --log-screenshot`.
This field can be repeated to include multiple links (for chunked reports).

__`callback_url`__: (Optional) A URL that the processor should `POST` when successful, which will
create the TestRun. Defaults to /api/results/create in the current project's environment (e.g. wpt.fyi for
wptdashboard, staging.wpt.fyi for wptdashboard-staging).

__`labels`__: (Optional) A comma-separated string of labels for this test run. Currently recognized
labels are "experimental" and "stable" (the release channel of the tested browser).

### /api/results/create

This is an *internal* endpoint used by the results processor.

## Querying test results

### /api/search

Search for test results over some set of test runs.

__Parameters__

__`run_ids`__ : Array-separated list of numerical ids associated with the runs
over which to search. IDs associated with runs can be obtained by querying the
`/api/runs` API. Defaults to the default runs returned by `/api/runs`.

__`q`__: Query string for search. Only results data for tests that contain the
`q` value as a substring of the test name will be returned. Defaults to the
empty string, which will yield all test results for the selected runs.

#### Examples

- https://staging.wpt.fyi/api/search?run_ids=6311104602963968,5132783244541952&q=xyz

<details><summary><b>Example JSON</b></summary>

```json
{
  "runs": [
    {
      "id": 6.311104602964e+15,
      "browser_name": "chrome",
      "browser_version": "68.0.3440.106",
      "os_name": "linux",
      "os_version": "16.04",
      "revision": "2dda7b8c10",
      "full_revision_hash": "2dda7b8c10c7566fa6167a32b09c85d51baf2a85",
      "results_url": "https:\/\/storage.googleapis.com\/wptd-staging\/2dda7b8c10c7566fa6167a32b09c85d51baf2a85\/chrome-68.0.3440.106-linux-16.04-edf200244e-summary.json.gz",
      "created_at": "2018-08-17T08:12:29.219847Z",
      "time_start": "2018-08-17T06:26:52.33Z",
      "time_end": "2018-08-17T07:50:09.155Z",
      "raw_results_url": "https:\/\/storage.googleapis.com\/wptd-results-staging\/2dda7b8c10c7566fa6167a32b09c85d51baf2a85\/chrome-68.0.3440.106-linux-16.04-edf200244e\/report.json",
      "labels": [
        "buildbot",
        "chrome",
        "stable"
      ]
    },
    {
      "id": 5.132783244542e+15,
      "browser_name": "firefox",
      "browser_version": "61.0.2",
      "os_name": "linux",
      "os_version": "16.04",
      "revision": "2dda7b8c10",
      "full_revision_hash": "2dda7b8c10c7566fa6167a32b09c85d51baf2a85",
      "results_url": "https:\/\/storage.googleapis.com\/wptd-staging\/2dda7b8c10c7566fa6167a32b09c85d51baf2a85\/firefox-61.0.2-linux-16.04-75ff911c43-summary.json.gz",
      "created_at": "2018-08-17T08:31:38.580221Z",
      "time_start": "2018-08-17T06:47:29.643Z",
      "time_end": "2018-08-17T08:15:18.612Z",
      "raw_results_url": "https:\/\/storage.googleapis.com\/wptd-results-staging\/2dda7b8c10c7566fa6167a32b09c85d51baf2a85\/firefox-61.0.2-linux-16.04-75ff911c43\/report.json",
      "labels": [
        "buildbot",
        "firefox",
        "stable"
      ]
    }
  ],
  "results": [
    {
      "test": "\/html\/dom\/elements\/global-attributes\/lang-xyzzy.html",
      "legacy_status": [
        {
          "passes": 1,
          "total": 1
        },
        {
          "passes": 1,
          "total": 1
        }
      ]
    }
  ]
}
```

</details>

## Metadata results

### /api/metadata

API endpoint for fetching all of the `link` metadata stored in the wpt-metadata
repository, with the (normally file-sharded) data all flattened into a JSON
object which is keyed by test name.

This endpoint accepts POST and GET requests.

- GET request returns Metadata Link Information by product, and requires product parameters;

- POST request searches Metadata Link by link url, and requires product parameters and payload.

__URL Parameters__

__`product`__ : browser[version[os[version]]]. e.g. `chrome-63.0-linux`

#### JSON Request Payload
```json
[
  {
    "link": "[pattern]"
  }
]
```

  Where `[pattern]` is any substring of the url field of a wpt-metadata `link` node.

#### Get Examples

- /api/metadata?product=chrome&product=safari

<details><summary><b>Example JSON</b></summary>

```json
{
  "/FileAPI/blob/Blob-constructor.html": [
    {
      "url": "https://github.com/web-platform-tests/results-collection/issues/661",
      "product": "chrome",
      "results:" [
        {
          "subtest": "Blob with type \"image/gif;\"",
          "status": "UNKNOWN"
        },
        {
          "subtest": "Invalid contentType (\"text/plain\")",
          "status": "UNKNOWN"
        }
      ]
    }
  ],
  "/service-workers/service-worker/fetch-request-css-base-url.https.html": [
    {
      "url": "https://bugzilla.mozilla.org/show_bug.cgi?id=1201160",
      "product": "firefox",
    }
  ],
  "/service-workers/service-worker/fetch-request-css-images.https.html": [
    {
      "url": "https://bugzilla.mozilla.org/show_bug.cgi?id=1532331",
      "product": "firefox"
    }
  ]
}
```
</details>

#### Post Examples
- POST /api/metadata?product=chrome\&product=firefox \
    exists:='[{"link":"bugs.chromium.org"}]'

<details><summary><b>Example JSON</b></summary>

```json
{
  "/IndexedDB/bindings-inject-key.html": [
    {
      "url": "bugs.chromium.org/p/chromium/issues/detail?id=934844"
    }
  ],
  "/html/browsers/history/the-history-interface/007.html": [
    {
      "url": "bugs.chromium.org/p/chromium/issues/detail?id=592874"
    }
  ]
}
```
</details>
