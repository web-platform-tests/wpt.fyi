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

webapp/ has both lint tests and tests based on
[web-component-test](https://www.npmjs.com/package/web-component-tester). There
are `npm` aliases for many of the common tasks, listed below.

- `npm test`: This will run the linting task followed by the web-component-tester task.
- `npm run lint`: This will run _only_ the linting task.
- `npm run lint-fix`: This will run the linting task with automatic lint fixing.
- `npm run wct`: This will run _only_ the web-component-tester task.
- `npm run wctp`: This will run the web-component-tester task with the `-p` flag
  to leave the browser open after the tests have completed.

When using `npm run`, any additional flags or options will be passed to the
underlying command. For example, to run a specific test only on chrome:

- `npm run wct -l chrome path/to/test/test-file.html`
