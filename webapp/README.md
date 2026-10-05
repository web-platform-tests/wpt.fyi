# wpt.fyi

## Testing webapp/

### Prerequisites:

1. [Setting up your environment](https://github.com/web-platform-tests/wpt.fyi#setting-up-your-environment)
2. [Running locally](https://github.com/web-platform-tests/wpt.fyi#running-locally)

Once the above steps are completed, run the following commands from within `webapp/`:

```sh
npm install
```

### Test commands

webapp/ has lint checks and component tests. The component tests are the
`components/test/*.test.js` files, and they run in Chrome and Firefox with
[Web Test Runner](https://modern-web.dev/docs/test-runner/overview/). There are
`npm` aliases for the common tasks, listed below.

- `npm test`: This will run the component tests.
- `npm run lint`: This will run _only_ the linting task.
- `npm run lint-fix`: This will run the linting task with automatic lint fixing.

When using `npm test`, any additional flags or options after `--` will be
passed to `web-test-runner`. The `WTR_BROWSERS` environment variable selects
the browsers (default: `chrome,firefox`), and `FIREFOX_PATH` can point to a
Firefox binary that is not at `/usr/bin/firefox`. For example:

- `WTR_BROWSERS=chrome npm test -- --files components/test/path.test.js` runs a
  single test file, only in Chrome.
- `npm test -- --watch` reruns the tests when files change.
- `npm test -- --manual --open` serves the tests for debugging in your own
  browser.

### Running web_components_test
To run `web_components_test` in any platform, first start a Docker instance.
Once the instance is running, execute the following in another terminal:
```sh
source util/commands.sh
wptd_exec make web_components_test
```
