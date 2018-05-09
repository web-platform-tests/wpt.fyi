# wpt.fyi API

This package defines and implements HTTP API endpoints for [wpt.fyi](https://wpt.fyi/), and this
document covers usage and parameters of those endpoints.

## Endpoints

An exhaustive list of the endpoints can be found in `routes.go`.

 - [/api/run](#apirun)
 - [/api/runs](#apiruns)
 - [/api/diff](#apidiff)
 - [/results](#results)

### /api/run

Gets a specific (single) TestRun metadata, for a given SHA[0:10] and platform.

__Parameters__

__`sha`__ :  SHA[0:10] of the runs to get, or the keyword `latest`. Defaults to `latest`.

__`platform`__ : browser[version[os[version]]]. e.g. `chrome-63.0-linux`

### /api/runs

Gets the TestRun metadata for all runs for a given SHA[0:10].

__Parameters__

__`sha`__ : SHA[0:10] of the runs to get, or the keyword `latest`. Defaults to `latest`.

__`max-count`__ : Maximum number of runs to get (for each browser). Only relevant when `sha` is `latest`. Maximum of 500.

### /api/diff

Computes a TestRun summary JSON blob of the differences between two TestRun
summary blobs.

__Parameters__

__`before`__ : [browser]@[sha] spec for the TestRun to use as the before state.

__`after`__ : [browser]@[sha] spec for the TestRun to use as the after state.

__`path`__ : Test path to filter by. `path` is a repeatable query parameter.

__`filter`__ : Differences to include in the summary.
 - `A` : Added - tests which are present after, but not before.
 - `D` : Deleted - tests which are present before, but not after.
 - `C` : Changed - tests which are present before and after, but the results summary is different.
 - `U` : Unchanged - tests which are present before and after, and the results summary count is not different.

### /results

Performs an HTTP redirect for the results summary JSON blob of the given TestRun.

__Parameters__

__`browser`__ : Browser to fetch the results for, e.g. `chrome`

__`sha`__ : SHA[0:10] of the TestRun to fetch, or the keyword `latest`. Defaults to `latest`.

### /api/manifest

Gets the JSON of the WPT manifest GitHub release asset, for a given sha (defaults to latest).

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
        "testharness": {...}
      },
    }

`manifest_item` is an **array** (nested in the map's `"file/path"` value's array) with varying contents. Loosely,

- For `testharness` entries: `[url, extras]`
  - `extras` example: `{"timeout": "long", "testdriver": True}`
- For `reftest` entries: `[url, references, extras]`
  - `references` example: `[[reference_url1, "=="], [reference_url2, "!="], ...]`
  - `extras` example: `{"timeout": "long", "viewport_size": ..., "dpi": ...}`
