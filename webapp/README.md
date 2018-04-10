# wpt.fyi


## Setup local development for wpt.fyi

Prerequisites: 

1. [Setting up your environment](https://github.com/web-platform-tests/wpt.fyi#setting-up-your-environment)
2. [Running locally](https://github.com/web-platform-tests/wpt.fyi#running-locally)

Once prerequisites are completed, run the following commands from within `webapp/`:

```sh
npm install -g bower
bower install
npm install -g web-component-tester
npm install
```

## Commands

- `npm test`: This will run the linting task followed by the web-component-tester task.
- `npm run lint`: This will run _only_ the linting task.
- `npm run lint-fix`: This will run the linting task with automatic lint fixing.
- `npm run wct`: This will run _only_ the web-component-tester task.
- `npm run wctp`: This will run the web-component-tester task with the `-p` flag to leave the browser open after the tests have completed. 
