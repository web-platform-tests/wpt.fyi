# [web-platform-tests dashboard](https://wpt.fyi/) 📈

[![GitHub Actions](https://github.com/web-platform-tests/wpt.fyi/workflows/Continuous%20Integration/badge.svg)](https://github.com/web-platform-tests/wpt.fyi/actions?query=workflow%3A%22Continuous+Integration%22+branch%3Amaster)

wpt.fyi is a dashboard of cross-browser results for [web-platform-tests](https://github.com/web-platform-tests/wpt), the data for which is uploaded by external services, primarily from various CI integrations in the wpt repo.

**Backend**: An [App Engine app](webapp/) for storing test run metadata and serving HTML

**Frontend**: [Polymer elements](webapp/components/) for loading and visualizing test results

## Using the data

All test result data is public. Please use our APIs to explore the data. For example, use the [results API](/api/README.md#apiresults) to download result summaries, and use the [runs API](/api/README.md#apiruns) to query runs and their metadata, which include links to other data like raw full reports.

### Product ID

This is a tuple of browser name, browser version, os name, os version, serialized in the form of `browser[-version[-os[-version]]]` (`[]` means optional), widely used in our APIs as the `product` parameter.

## Development

### Setting up your environment

You'll need [Docker](https://www.docker.com/). With Docker installed, start the development container:

```sh
docker pull webplatformtests/wpt.fyi:latest   # Optional: this forces fetching the latest version, instead of using the locally cached version.
./util/docker-dev/run.sh
```

This starts a Docker instance named `wptd-dev-instance`.

### Running locally

Once the instance is running, you can fire up the web server in another terminal:

```sh
./util/docker-dev/web_server.sh
```

This will build dependencies and start the Google App Engine development server inside `wptd-dev-instance`.

Meanwhile, you'll also need to populate the app datastore with some initial data. In another terminal,
execute the script which leverages `util/populate_dev_data.go` by running:

```sh
./util/docker-dev/dev_data.sh
```

See [CONTRIBUTING.md](/CONTRIBUTING.md) for more information on local development.

## Miscellaneous

### WPT documentation page for each browser

- Chromium: https://chromium.googlesource.com/chromium/src/+/master/docs/testing/web_platform_tests.md
- Firefox: https://wiki.mozilla.org/Auto-tools/Projects/web-platform-tests
- WebKit: https://trac.webkit.org/wiki/WebKitW3CTesting

### Location of the WPT in each browser’s source tree

- Chromium: [`src/third_party/blink/web_tests/external/wpt`](https://cs.chromium.org/chromium/src/third_party/blink/web_tests/external/wpt/)
- Firefox: [`testing/web-platform/tests`](https://dxr.mozilla.org/mozilla-central/source/testing/web-platform/tests)
- WebKit: [`LayoutTests/imported/w3c/web-platform-tests`](https://trac.webkit.org/browser/trunk/LayoutTests/imported/web-platform-tests/wpt)

### You can run almost any WPT test on wpt.live

Try out http://wpt.live/html/semantics/forms/the-input-element/checkbox.html

This doesn't work with some HTTPS tests. Also be advised that the server is not intended for frequent large-scale test runs.

### Sources of inspiration

- ECMAScript 6 compatibility table - https://kangax.github.io/compat-table/es6/
- https://html5test.com/
