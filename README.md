# [web-platform-tests dashboard](https://wpt.fyi/) 📈 [![Build Status](https://travis-ci.com/web-platform-tests/wpt.fyi.svg?branch=master)](https://travis-ci.com/web-platform-tests/wpt.fyi)

wpt.fyi is a dashboard of cross-browser results for [web-platform-tests](https://github.com/web-platform-tests/wpt), the data for which is uploaded by external services, primarily via execution of the [results-collection](https://github.com/web-platform-tests/results-collection) repo.

**Backend**: An [App Engine app](webapp/) for storing test run metadata and serving HTML

**Frontend**: [Polymer elements](webapp/components/) for loading and visualizing test results

## Setting up your environment

You'll need [Docker](https://www.docker.com/). With Docker installed, build the base image and development image, and start a development server instance:

```sh
docker build -t wptd-dev .
./util/docker-dev/run.sh
```

This starts a Docker instance named `wptd-dev-instance`.

## Running locally

Once the instance is running, you can fire up the web server in another terminal:

```sh
./util/docker-dev/web_server.sh
```

This will build dependencies and start the Google App Engine development server inside `wptd-dev-instance`.

With the webserver running, you'll also need to populate the app datastore with some initial data. In another terminal,
execute the script which leverages `util/populate_dev_data.go` by running:

```sh
./util/docker-dev/dev_data.sh
```

See [CONTRIBUTING.md](/CONTRIBUTING.md) for more information on local development.

# Filesystem and network output

- This script will only write files under `config['build_path']`.
- One run will write approximately 111MB to the filesystem.
- If --upload is specified, it will upload that 111MB of results to GCS.
- To upload results, you must be logged in with `gcloud` in the `wpt.fyi` project.

## Using the data

All test result data is public. Please use our APIs to explore the data. For example, use the [results API](/api/README.md#apiresults) to download result summaries, and use the [runs API](/api/README.md#apiruns) to query runs and their metadata, which include links to other data like raw full reports.

### Large-scale analysis

There is no public API for TestRuns, so if you need to access only the most recent results, looking at
the main page will give you the latest test SHAs. If you need to access earlier results, an
exhaustive search is the only way to do that (see issue [#73](https://github.com/web-platform-tests/wpt.fyi/issues/73) and [#43](https://github.com/web-platform-tests/wpt.fyi/issues/43)).

## Miscellaneous

#### WPT documentation page for each browser

- Chromium: https://chromium.googlesource.com/chromium/src/+/master/docs/testing/web_platform_tests.md
- Firefox: https://wiki.mozilla.org/Auto-tools/Projects/web-platform-tests
- WebKit: https://trac.webkit.org/wiki/WebKitW3CTesting

#### Location of the WPT in each browser’s source tree

- Chromium: [`src/third_party/WebKit/LayoutTests/external/wpt`](https://cs.chromium.org/chromium/src/third_party/WebKit/LayoutTests/external/wpt/)
- Firefox: [`testing/web-platform/tests`](https://dxr.mozilla.org/mozilla-central/source/testing/web-platform/tests)
- WebKit: [`LayoutTests/imported/w3c/web-platform-tests`](https://trac.webkit.org/browser/trunk/LayoutTests/imported/web-platform-tests/wpt)

#### You can run almost any WPT test on w3c-test.org

Try out http://w3c-test.org/html/semantics/forms/the-input-element/checkbox.html

This doesn't work with some HTTPS tests. Also be advised that the server is not intended for frequent large-scale test runs.

#### Sources of inspiration

- ECMAScript 6 compatibility table - https://kangax.github.io/compat-table/es6/
- https://html5test.com/

# Appendix

## Terminology

### Platform ID

These are the keys in [`webapp/browsers.json`](webapp/browsers.json). They're used to identify a tuple (browser name, browser version, os name, os version).
