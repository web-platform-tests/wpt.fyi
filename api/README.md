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
 - [/api/revisions/epochs](#apirevisionsepochs)
 - [/api/revisions/latest](#apirevisionslatest)
 - [/api/revisions/list](#apirevisionslist)

Also see [results creation](#results-creation) for endpoints to add new data.

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

__`max-count`__ : Maximum number of runs to get (for each browser). Maximum of 500.

#### Examples

- https://wpt.fyi/api/interop
- https://wpt.fyi/api/interop?product=chrome-67
- https://wpt.fyi/api/interop?label=experimental

__Example response JSON__

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

## Results creation

### /api/results/upload

Uploads a wptreport to the dashboard to create the test run.

This endpoint only accepts POST requests. Requests need to be authenticated via HTTP basic auth.
Please contact [Ecosystem Infra](mailto:ecosystem-infra@chromium.org) if you want to register as a
"test runner", to upload results.

#### File payload

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

#### URL payload

__Content type__: `application/x-www-form-urlencoded`

__Parameters__

__`result_url`__: A URL to a **gzipped** JSON file produced by `wpt run --log-wptreport` (see above
for its format). This field can be repeated to include multiple links (for chunked reports).

__`labels`__: (Optional) A comma-separated string of labels for this test run. Currently recognized
labels are "experimental" and "stable" (the release channel of the tested browser).

### /api/results/create

This is an *internal* endpoint used by the results processor.

## Announcement of revisions-of-interest

The `/api/revisions` namespace contains APIs for accessing WPT
_revisions-of-interest_. The primary use case for this API is synchronizing the
WPT revision used by active test runners.

### /api/revisions/epochs

Get the collection of epochs over which revisions are announced. For example,
the `two_hourly` epoch announces the last revision prior to every two-hour
interval (for each day: last revision prior to midnight, 2AM, 4AM, etc.). Weeks
start on Monday. All epochs calculated relative to UTC times.

__Parameters__

None

__Example JSON__

```json
[
	{
		"id": "weekly",
		"label": "Once per week (weekly)",
		"description": "The last PR merge commit of each week, by UTC commit timestamp on master. Weeks start on Monday.",
		"min_duration_sec": 604800,
		"max_duration_sec": 604800
	},
	{
		"id": "daily",
		"label": "Once per day (daily)",
		"description": "The last PR merge commit of each day, by UTC commit timestamp on master.",
		"min_duration_sec": 86400,
		"max_duration_sec": 86400
	},
	{
		"id": "eight_hourly",
		"label": "Once every eight hours",
		"description": "The last PR merge commit of eight-hour partition of the day, by UTC commit timestamp on master. E.g., epoch changes at 00:00:00, 00:08:00, etc..",
		"min_duration_sec": 28800,
		"max_duration_sec": 28800
	},
	{
		"id": "four_hourly",
		"label": "Once every four hours",
		"description": "The last PR merge commit of four-hour partition of the day, by UTC commit timestamp on master. E.g., epoch changes at 00:00:00, 00:04:00, etc..",
		"min_duration_sec": 14400,
		"max_duration_sec": 14400
	},
	{
		"id": "two_hourly",
		"label": "Once every two hours",
		"description": "The last PR merge commit of two-hour partition of the day, by UTC commit timestamp on master. E.g., epoch changes at 00:00:00, 00:02:00, etc..",
		"min_duration_sec": 7200,
		"max_duration_sec": 7200
	},
	{
		"id": "hourly",
		"label": "Once per hour (hourly)",
		"description": "The last PR merge commit of each hour, by UTC commit timestamp on master.",
		"min_duration_sec": 3600,
		"max_duration_sec": 3600
	}
]
```

### /api/revisions/latest

Get the latest announced revision for all epochs. For convenience, the metadata
for the epochs is included as well.

__Parameters__

None

__Example JSON__

```json
{
	"revisions": {
		"daily": {
			"hash": "5462552a420cba8886cf50bb9d9674d7a79fdc4e",
			"commit_time": "2018-08-13T23:36:57Z"
		},
		"eight_hourly": {
			"hash": "1f8b6c9a44e5c6b64bac140c542b570360f886ac",
			"commit_time": "2018-08-14T15:14:39Z"
		},
		"four_hourly": {
			"hash": "1f8b6c9a44e5c6b64bac140c542b570360f886ac",
			"commit_time": "2018-08-14T15:14:39Z"
		},
		"hourly": {
			"hash": "1f8b6c9a44e5c6b64bac140c542b570360f886ac",
			"commit_time": "2018-08-14T15:14:39Z"
		},
		"two_hourly": {
			"hash": "1f8b6c9a44e5c6b64bac140c542b570360f886ac",
			"commit_time": "2018-08-14T15:14:39Z"
		},
		"weekly": {
			"hash": "d31eacaff0c4d96f8c125c21faac6e0f75dd683c",
			"commit_time": "2018-08-11T18:20:16Z"
		}
	},
	"epochs": [
		{
			"id": "hourly",
			"label": "Once per hour (hourly)",
			"description": "The last PR merge commit of each hour, by UTC commit timestamp on master.",
			"min_duration_sec": 3600,
			"max_duration_sec": 3600
		},
		{
			"id": "two_hourly",
			"label": "Once every two hours",
			"description": "The last PR merge commit of two-hour partition of the day, by UTC commit timestamp on master. E.g., epoch changes at 00:00:00, 00:02:00, etc..",
			"min_duration_sec": 7200,
			"max_duration_sec": 7200
		},
		{
			"id": "four_hourly",
			"label": "Once every four hours",
			"description": "The last PR merge commit of four-hour partition of the day, by UTC commit timestamp on master. E.g., epoch changes at 00:00:00, 00:04:00, etc..",
			"min_duration_sec": 14400,
			"max_duration_sec": 14400
		},
		{
			"id": "eight_hourly",
			"label": "Once every eight hours",
			"description": "The last PR merge commit of eight-hour partition of the day, by UTC commit timestamp on master. E.g., epoch changes at 00:00:00, 00:08:00, etc..",
			"min_duration_sec": 28800,
			"max_duration_sec": 28800
		},
		{
			"id": "daily",
			"label": "Once per day (daily)",
			"description": "The last PR merge commit of each day, by UTC commit timestamp on master.",
			"min_duration_sec": 86400,
			"max_duration_sec": 86400
		},
		{
			"id": "weekly",
			"label": "Once per week (weekly)",
			"description": "The last PR merge commit of each week, by UTC commit timestamp on master. Weeks start on Monday.",
			"min_duration_sec": 604800,
			"max_duration_sec": 604800
		}
	]
}
```

### /api/revisions/list

List a particular range of revision. This API allows the client to query for
announced revisions for particular epochs, a particular time range, etc..

__Parameters__

__`epochs`__ : A potentially repeated parameter. Each parameter value contains
the `id` of some epoch known by the announcer. Defaults to all known epochs.

__`num_revision`__: The number of epochal revisions _for each epoch in `epochs`
values_ to include in the response. Defaults to 100. Response will include an
`error` field when fewer than the requested number could be found for some
epoch(s).

__`at`__ : An RFC3339-encoded timestamp describing the upper limit on the time
range for fetching epochal revisions. Defaults to now.

__`start`__: An RFC3339-encoded timestamp describing the lower limit on the time
range for fetching epochal revisions. Defaults to the date which is
`num_revisions * longest(epochs).max_duration_sec` seconds prior to now.

#### Examples

- https://wpt.fyi/api/revisions/list?epochs=hourly&epochs=two_hourly&num_revisions=10&at=2018-01-10T00:00:00Z&start=2018-01-01T00:00:00Z
- https://wpt.fyi/api/revisions/list?epochs=daily&num_revisions=10

__Example JSON__

```json
{
	"revisions": {
		"daily": [
			{
				"hash": "5462552a420cba8886cf50bb9d9674d7a79fdc4e",
				"commit_time": "2018-08-13T23:36:57Z"
			},
			{
				"hash": "d31eacaff0c4d96f8c125c21faac6e0f75dd683c",
				"commit_time": "2018-08-11T18:20:16Z"
			},
			{
				"hash": "b382ac7192087da0a7439902e20be76ab7587ee8",
				"commit_time": "2018-08-10T21:32:20Z"
			},
			{
				"hash": "9f51afc215d4f882a7ae069494ed37ea2c9503b1",
				"commit_time": "2018-08-09T22:03:24Z"
			}
		],
		"hourly": [
			{
				"hash": "1f8b6c9a44e5c6b64bac140c542b570360f886ac",
				"commit_time": "2018-08-14T15:14:39Z"
			},
			{
				"hash": "39aac0cde328471b8a97b136c26a5293f55771b3",
				"commit_time": "2018-08-14T14:56:57Z"
			},
			{
				"hash": "c02862684bb2faac9000b1ec1ad785464c97f5d9",
				"commit_time": "2018-08-14T13:19:39Z"
			},
			{
				"hash": "a20165544242305af9b699fbe5d1be2ec78243cd",
				"commit_time": "2018-08-14T10:12:12Z"
			}
		]
	},
	"epochs": [
		{
			"id": "hourly",
			"label": "Once per hour (hourly)",
			"description": "The last PR merge commit of each hour, by UTC commit timestamp on master.",
			"min_duration_sec": 3600,
			"max_duration_sec": 3600
		},
		{
			"id": "daily",
			"label": "Once per day (daily)",
			"description": "The last PR merge commit of each day, by UTC commit timestamp on master.",
			"min_duration_sec": 86400,
			"max_duration_sec": 86400
		}
	]
}
```
