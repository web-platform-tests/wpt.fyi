# wpt.fyi GitHub Checks integration

This directory implements the wpt.fyi (and staging.wpt.fyi) integration with
the GitHub Checks API for commits made to
[WPT](https://github.com/web-platform-tests/wpt). The goal of this integration
is to provide summary data (reporting regressions, etc) for test-suite runs
performed by our CI systems (Azure Pipelines and Taskcluster).

The wpt.fyi GitHub Checks code owns only computing the summary data and pushing
it to GitHub, not ingesting the results from the CI systems. For those, see the
[`/api/azure`](/api/azure/) and [`/api/taskcluster`](/api/taskcluster)
directories.

## Links

* [Design doc](https://docs.google.com/document/d/1EsMmll5s5ZA4kvaCeFUKFfdjG8DMxGANX8JDPl8rKFE/edit)
