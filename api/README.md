# wpt.fyi API

This package defines and implements HTTP API endpoints for [wpt.fyi](https://wpt.fyi/), and this
document covers usage and parameters of those endpoints.

## Endpoints

An exhaustive list of the endpoints can be found in `routes.go`.

 - [/api/runs](#apiruns)
 - [/api/runs/{id}](#apirunsid)
 - [/api/run](#apirun)
 - [/api/diff](#apidiff)
 - [/api/results](#apiresults)

## TestRun entities

`TestRun` entities represent metadata about an execution of the [wpt](https://github.com/web-platform-tests/wpt) test suite, on a particular product. Tests are run on a regular basis, and each entry in `/api/runs` annotates when the tests were executed, which product they were executed on, and the where the results are stored.

### /api/runs

Gets the TestRun metadata for all runs for a given SHA[0:10].

__Parameters__

__`sha`__ : SHA[0:10] of the runs to get, or the keyword `latest`. Defaults to `latest`.

__`product`__ : Product(s) to include (repeated param), e.g. `chrome` or `firefox-60`.

__`labels`__: A comma-separated list of labels, e.g. `firefox,stable`; only runs with all
the given labels will be returned. There are currently two kinds of labels supported,
browser names (`chrome`, `edge`, `firefox`, `safari`) and release channels (`experimental`
or `stable`).

__`from`__ : RFC3339 timestamp, for which to include runs that occured after the given time.

__`max-count`__ : Maximum number of runs to get (for each browser). Maximum of 500.

#### Examples

- https://wpt.fyi/api/runs?product=chrome&product=safari
- https://wpt.fyi/api/runs?product=chrome&from=2018-01-01T00:00:00Z&max-count=10

__Example JSON__

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

### /api/runs/{id}

Gets a specific (single) TestRun metadata by its datastore ID.

#### Example

https://wpt.fyi/api/runs/5164888561287168

__Example JSON__

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

### /api/run

Gets a specific (single) TestRun metadata by `product` and `sha`.

__Parameters__

__`sha`__ :  SHA[0:10] of the runs to get, or the keyword `latest`. Defaults to `latest`.

__`product`__ : browser[version[os[version]]]. e.g. `chrome-63.0-linux`

#### Example

https://wpt.fyi/api/run?sha=latest&product=chrome

__Example JSON__

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

### /api/shas

Gets an array of revisions (SHA[0:10]), in reverse chronological order.
This method behaves similarly to [/api/runs](#apiruns) above, but projects the `revision` field's value.

__Parameters__

__`complete`__ : boolean for whether to get only SHAs which were executed across all four of the default (stable) browsers. Not compatible with `product`.

__`product`__ : Product(s) to include (repeated param), e.g. `chrome` or `firefox-60`

__`from`__ : RFC3339 timestamp, for which to include runs that occured after the given time.

__`max-count`__ : Maximum number of runs to get (for each browser). Maximum of 500.

#### Example

https://wpt.fyi/api/shas?product=chrome

__Example JSON__

    [
      "98530fb944",
      "2bd11b91d4"//, ...
    ]

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

__Example JSON__ (from the summary.json.gz output):

    {
      "/css/css-text/i18n/css3-text-line-break-opclns-213.html": [1, 1],
      "/css/css-writing-modes/table-progression-vrl-001.html": [1, 1],
      // ...
    }

### /api/diff

Computes a TestRun summary JSON blob of the differences between two TestRun
summary blobs.

__Parameters__

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

NOTE: The full SHA of the fetched manifest is returned in the HTTP response header `x-wpt-sha`, e.g.

    x-wpt-sha: abcdef0123456789abcdef0123456789abcdef01

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

## /api/results/upload

Uploads a wptreport to the dashboard to create the test run.

This endpoint only accepts POST requests. Requests need to be authenticated via HTTP basic auth.
Please contact [Ecosystem Infra](mailto:ecosystem-infra@chromium.org) if you want to register as a
"test runner", to upload results.

### File payload

__Content type__: `multipart/form-data`

__Parameters__

__`result_file`__: A **gzipped** JSON file produced by `wpt run --log-wptreport`.

__`labels`__: (Optional) A comma-separated string of labels for this test run. Currently recognized
labels are "experimental" and "stable" (the release channel of the tested browser).

The JSON file roughly looks like this:

```json
{
  "results": [...],
  "time_start": MICROSECONDS_SINCE_EPOCH,
  "time_end": MICROSECONDS_SINCE_EPOCH,
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

The `time_start` and `time_end` fields are numerical timestamps (in microseconds since the UNIX epoch)
when the whole test run starts and finishes. They are optional, but encouraged. `wpt run` produces
them in the report by default.

`run_info.{revision,product,browser_version,os}` are required, and should be automatically
generated by `wpt run`. If for some reason the report does not contain these fields (e.g. old WPT
version, Sauce Labs, or custom runners), they can be overridden with the following *optional*
parameters in the POST payload (this is __NOT__ recommended; please include metadata in the reports
whenever possible):

* __`revision`__
* __`browser_name`__ (note that it is not called `product` here)
* __`browser_version`__
* __`os_name`__ (note that it is not called `os` here)
* __`os_version`__

### URL payload

__Content type__: `application/x-www-form-urlencoded`

__Parameters__

TODO
