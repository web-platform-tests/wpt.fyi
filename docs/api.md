# API endpoints documentation

This document covers the HTTP REST endpoints available on [wpt.fyi](http://wpt.fyi).

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
