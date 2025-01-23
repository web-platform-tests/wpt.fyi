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
 - [/api/manifest](#apimanifest)
 - [/api/search](#apisearch)
 - [/api/metadata](#apimetadata)
 - [/api/metadata/pending](#apimetadatapending)
 - [/api/metadata/triage](#apimetadatatriage)
 - [/api/bsf](#apibsf)
 - [/api/history](#apihistory)

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
        "results_url": "https://storage.googleapis.com/wptd/2bd11b91d4/chrome-stable-linux-summary_v2.json.gz",
        "created_at": "2018-06-05T08:27:30.627865Z",
        "raw_results_url": "https://storage.googleapis.com/wptd-results/2bd11b91d490ddd5237bcb6d8149a7f25faaa101/chrome_67.0.3396.62_linux_4.4/report.json"
      }
    ]

</details>

### /api/runs/{id}

Gets a specific (single) TestRun metadata by its datastore ID.

#### Example

https://wpt.fyi/api/runs/5184362994728960

<details><summary><b>Example JSON</b></summary>

    {
      "id": "5164888561287168",
      "browser_name": "chrome",
      "browser_version": "67.0.3396.62",
      "os_name": "linux",
      "os_version": "4.4",
      "revision": "2bd11b91d4",
      "full_revision_hash": "2bd11b91d490ddd5237bcb6d8149a7f25faaa101",
      "results_url": "https://storage.googleapis.com/wptd/2bd11b91d4/chrome-stable-linux-summary_v2.json.gz",
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
      "results_url": "https://storage.googleapis.com/wptd/2bd11b91d4/chrome-stable-linux-summary_v2.json.gz",
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

The summary JSON format has been updated as of July 2022, and all requisite
summary files should now follow this  newformat. Summary files with the new format
are denoted with the `_v2` file name suffix. This change was made to
differentiate a test's overall status value from the subtest passes and
totals.

The v2 summary JSON is in the format

    {
      "/path/to/test.html": {
        "s": "O",
        "c": [1, 1]
      },
    }

Each test path has two properties.

`s`, or status, which is an abbreviated value to the test's overall status.

`c`, or counts, which is an array containing
[`number of subtest passes`, `total subtests`].

__Status abbrevations__

| Status              | Abbreviation |
|---------------------|--------------|
| OK                  | O            |
| PASS                | P            |
| FAIL                | F            |
| SKIP                | S            |
| ERROR               | E            |
| NOTRUN              | N            |
| CRASH               | C            |
| TIMEOUT             | T            |
| PRECONDITION_FAILED | PF           |

Any summary files before this update follow the old JSON format (v1). The v1
summary format has no additional name suffix, unlike v2.

The v1 JSON is in the format

    {
      "/path/to/test.html": [2, 2],
    }

Where the array contains [`number of subtest passes`, `total subtests`].
The test's overall status is added with these subtest values. A passing status
value (`OK` or `PASS`) will increment the number of subtest passes.

__Parameters__

__`product`__ : Product to fetch the results for, e.g. `chrome-66`

__`sha`__ : SHA[0:10] of the TestRun to fetch, or the keyword `latest`. Defaults to `latest`.

#### Example

https://wpt.fyi/api/results?product=chrome

<details><summary><b>Example JSON</b> (from the summary_v2.json.gz output)</summary>

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

__`run_ids`__ : Exactly two numerical IDs for the "before" and "after" runs (in
that order), separted by a comma. IDs associated with runs can be obtained by
querying the `/api/runs` API. This overrides the `before` and `after` params.

__`before`__ : [product]@[sha] spec for the TestRun to use as the before state.

__`after`__ : [product]@[sha] spec for the TestRun to use as the after state.

__`path`__ : Test path to filter by. `path` is a repeatable query parameter.

__`filter`__ : Differences to include in the summary.
 - `A` : Added - tests which are present after, but not before.
 - `D` : Deleted - tests which are present before, but not after.
 - `C` : Changed - tests which are present before and after, but the results summary is different.
 - `U` : Unchanged - tests which are present before and after, and the results summary count is not different.

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
Please [file an issue](https://github.com/web-platform-tests/wpt.fyi/issues/new) if you want to
register as a "test runner", to upload results.

#### File payload

__Content type__: `multipart/form-data`

__Parameters__

__`labels`__: (Optional) A comma-separated string of labels for this test run. Currently recognized
labels are "experimental", "stable" (the release channel of the tested browser) and "master" (test run
from the master branch).

__`callback_url`__: (Optional) A URL that the processor should `POST` when successful, which will
create the TestRun. Defaults to /api/results/create in the current project's environment (e.g. wpt.fyi for
wptdashboard, staging.wpt.fyi for wptdashboard-staging).

__`result_file`__: A **gzipped** JSON file, with the filename ending with `.gz` extension, produced by `wpt run --log-wptreport`.
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

__`archive_url`__: A URL to a ZIP archive containing files like `wpt_report*.json` and
`wpt_screenshot*.json`, similar to `result_url` and `screenshot_url` respectively. This field can
be repeated to include multiple links (for chunked reports). This field cannot co-exist with
`result_url` or `screenshot_url`.

__`callback_url`__: (Optional) A URL that the processor should `POST` when successful, which will
create the TestRun. Defaults to /api/results/create in the current project's environment (e.g. wpt.fyi for
wptdashboard, staging.wpt.fyi for wptdashboard-staging).

__`labels`__: (Optional) A comma-separated string of labels for this test run. Currently recognized
labels are "experimental" and "stable" (the release channel of the tested browser).

### /api/results/create

This is an *internal* endpoint used by the results processor.

## Querying test results

### /api/search

Search for test results over some set of test runs. This endpoint accepts POST and GET requests.

- POST requests are forwarded to the searchcache for structured queries, with
  `run_ids` and `query` fields in the JSON payload; see [search query](./query/README.md#apisearch)
  documentaton for more information.

- GET requests are unstructured queries with the following parameters:


__Parameters__

__`run_ids`__ : (Optional) A comma-separated list of numerical ids associated
with the runs over which to search. IDs associated with runs can be obtained by
querying the `/api/runs` API. Defaults to the default runs returned by
`/api/runs`. NOTE: This is not the same set of runs as is shown on wpt.fyi by
default.

__`q`__: (Optional) A query string for search. Only results data for tests that
contain the `q` value as a substring of the test name will be returned. Defaults
to the empty string, which will yield all test results for the selected runs.
NOTE: structured search queries are not supported.

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
      "results_url": "https:\/\/storage.googleapis.com\/wptd-staging\/2dda7b8c10c7566fa6167a32b09c85d51baf2a85\/chrome-68.0.3440.106-linux-16.04-edf200244e-summary_v2.json.gz",
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
      "results_url": "https:\/\/storage.googleapis.com\/wptd-staging\/2dda7b8c10c7566fa6167a32b09c85d51baf2a85\/firefox-61.0.2-linux-16.04-75ff911c43-summary_v2.json.gz",
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
    exists:='[{"link":"issues.chromium.org"}]'

<details><summary><b>Example JSON</b></summary>

```json
{
  "/IndexedDB/bindings-inject-key.html": [
    {
      "url": "issues.chromium.org/issues/934844"
    }
  ],
  "/html/browsers/history/the-history-interface/007.html": [
    {
      "url": "issues.chromium.org/issues/592874"
    }
  ]
}
```
</details>

### /api/metadata/pending
API endpoint for retrieving pending metadata whose PRs are not merged yet. This endpoint is used along with the /api/metadata endpoint to retrieve all metadata, pending or non-pending. It accepts GET requests without any parameters. It returns the same JSON response as [/api/metadata](#apimetadata).

This endpoint is a best-effort API, because in some rare cases, e.g. both the Redis server and its replica go down, pending metadata information can be lost temporarily.

### /api/metadata/triage

This API is available for trusted third parties.

To use the Triage Metadata API, you first need to sign in to [wpt.fyi](https://wpt.fyi/) (top-right corner; 'Sign in with GitHub'). For more information on wpt.fyi login, see [here](https://docs.google.com/document/d/1iRkaK6cGgXp3DKbNbPMVsYGMaOHO-5CfqEuLPUR_2HM).

The logged-in user also needs to belong to the ['web-platform-tests' GitHub organization](https://github.com/orgs/web-platform-tests/people). To join, please [file an issue](https://github.com/web-platform-tests/wpt/issues/new?), including the reason you need access to the Triage Metadata API.

Once logged in, you can send a request to /api/metadata/triage to triage metadata. This endpoint only accepts PATCH requests and appends a triage JSON object to the existing Metadata YML files. The JSON object is a flattened YAML `Links` structure that is keyed by test name [Test path](https://docs.google.com/document/d/1oWYVkc2ztANCGUxwNVTQHlWV32zq6Ifq9jkkbYNbSAg/edit#heading=h.t7ysbpr8er1y); see below for an example.

This endpoint returns the URL of a PR that is created in the wpt-metadata repo.

<details><summary><b>Example JSON Body</b></summary>

```json
{
  "/FileAPI/blob/Blob-constructor.html": [
    {
      "url": "https://github.com/web-platform-tests/results-collection/issues/661",
      "product": "chrome",
      "results:" [
        {
          "subtest": "Blob with type \"image/gif;\"",
          "status": 6
        },
        {
          "subtest": "Invalid contentType (\"text/plain\")",
          "status": 0
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

## Browser Specific Failure

### /api/bsf
Gets the BSF data of Chrome, Firefox, Safari for the home directory.

The endpoint accepts GET requests.

__Parameters__

__`from`__ : (Optional) RFC3339 timestamp, for which to include BSF data that occured after the given time inclusively.

__`to`__ : (Optional) RFC3339 timestamp, for which to include BSF data that occured before the given time exclusively.

__`experimental`__ : A boolean to return BSF data for experimental or stable runs. Defaults to false.

__JSON Response__

The response has three top-level fields:

`lastUpdateRevision` indicates the latest WPT Revision updated in `data`.

`fields` corresponds to the fields (columns) in the `data` table and has the format of an array of:

- sha, date, [product-version, product-score]+

`data` returns BSF data in chronological order.

<details><summary><b>Example JSON</b></summary>

```json
{
   "lastUpdateRevision":"eea0b54014e970a2f94f1c35ec6e18ece76beb76",
   "fields":[
      "sha",
      "date",
      "chrome-version",
      "chrome",
      "firefox-version",
      "firefox",
      "safari-version",
      "safari"
   ],
   "data":[
      [
         "eea0b54014e970a2f94f1c35ec6e18ece76beb76",
         "2018-08-07",
         "70.0.3510.0 dev",
         "602.0505256721168",
         "63.0a1",
         "1617.1788882804883",
         "12.1",
         "2900.3438625831423"
      ],
      [
         "203c34855f6871d6e55eaf7b55b50dad563f781f",
         "2018-08-18",
         "70.0.3521.2 dev",
         "605.3869030161061",
         "63.0a1",
         "1521.908686731921",
         "12.1",
         "2966.686195133767"
      ]
   ]
}
```
</details>

## Test History

### /api/history

This endpoint accepts POST requests. It returns historical test run information for a given test name.

#### JSON Request Payload
```json
{
    "test_name": "example test name"
}
```
#### JSON Response
The returned JSON will contain a history of test runs for each major browser: Chrome, Firefox, Edge, and Safari.

Each individual subtest run will have a `date`, `status`, and `run_id`.

The first test entry for each browser is represented with an empty string. This represents the `Harness Status` if there are multiple tests, or the `Test Status` if there is only one test.

<details><summary><b>Example JSON</b></summary>

```json
{
    "results": {
        "chrome": {
            "": [
                {
                    "date": "2022-06-02T06:02:55.000Z",
                    "status": "TIMEOUT",
                    "run_id": "5074677897101312"
                }
            ],
            "subtest_name_1": [
                {
                    "date": "2022-06-02T06:02:55.000Z",
                    "status": "PASS",
                    "run_id": "5074677897101312"
                }
            ]
        },
        "firefox": {
            "": [
                {
                    "date": "2022-06-02T06:02:55.000Z",
                    "status": "OK",
                    "run_id": "5074677897101312"
                }
            ],
            "subtest_name_1": [
                {
                    "date": "2022-06-02T06:02:55.000Z",
                    "status": "PASS",
                    "run_id": "5074677897101312"
                }
            ]
        }
    }
}
```
</details>
