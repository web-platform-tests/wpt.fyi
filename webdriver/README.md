# wpt.fyi Webdriver tests

This directory containers integration tests for webapp/. These tests bring up
the full webserver then use a Golang Webdriver client,
[tebeka/selenium](https://github.com/tebeka/selenium), to load pages in a
browser (Chrome or Firefox) and assert that the webapp behaves as expected.

To run the tests, from the root `wpt.fyi` directory, run:

    make go_large_test

You can run just one of Chrome or Firefox via:

    make go_chrome_test
    make go_firefox_test

Note that when running these tests outside of docker (see [Running in
docker](#running-in-docker)), they will use your locally installed browser and
webdriver clients, so it is worth making sure they are the versions you expect.

## Debugging

Debugging the webdriver/ tests can be difficult as they are integration tests
and the problem can occur anywhere from controlling the browser, to the webapp
frontend, to the backend - and other weird bugs inbetween! This section
contains some tips on how to effectively debug them.

After running one of the above `make` commands at least once, one can directly
run the golang tests via:

```
# You can copy GECKODRIVER_PATH out of the make output; it should be installed
# locally under webapp/node_modules/selenium-standalone/...
go test -v -timeout=15m -tags=large ./webdriver -args \
    -firefox_path=/usr/bin/firefox \
    -geckodriver_path=$GECKODRIVER_PATH \
    -chrome_path=/usr/bin/google-chrome \
    -chromedriver_path=/usr/bin/chromedriver \
    -frame_buffer=true \
    -staging=false \
    -browser=chrome  # Or firefox
```

It is worth comparing this command-line against the Makefile, in case this
documentation becomes out of date.

### Running only one test

If you only need to run one test, you can use the golang test [`-run`
parameter](https://golang.org/pkg/testing/#hdr-Subtests_and_Sub_benchmarks).
For example:

```
go test -v -timeout=15m -tags=large ./webdriver \
    -run TestProductParam_Order/Order \
    -args \
    -firefox_path=/usr/bin/firefox \
    -geckodriver_path=$GECKODRIVER_PATH \
    -chrome_path=/usr/bin/google-chrome \
    -chromedriver_path=/usr/bin/chromedriver \
    -frame_buffer=true \
    -staging=false \
    -browser=chrome  # Or firefox
```

### Visual Output

Many of the tests run some javascript (or click on an element, etc) and expect
to find some resulting change on the page. When that doesn't occur, they
usually just timeout and it can be difficult to know why. One very useful trick
is to enable visual output, so that you can actually see the webpage as the
test runs.

To do this, set the `frame_buffer` argument to `false`, e.g.:

```
go test -v -timeout=15m -tags=large ./webdriver -args \
    -firefox_path=/usr/bin/firefox \
    -geckodriver_path=$GECKODRIVER_PATH \
    -chrome_path=/usr/bin/google-chrome \
    -chromedriver_path=/usr/bin/chromedriver \
    -frame_buffer=false \
    -staging=false \
    -browser=chrome  # Or firefox
```

### Verbose webdriver output

By default, webdriver output is hidden as it is very noisy. You can re-enable
it by passing `-debug=true` to the tests, e.g.:

```
go test -v -timeout=15m -tags=large ./webdriver -args \
    -firefox_path=/usr/bin/firefox \
    -geckodriver_path=$GECKODRIVER_PATH \
    -chrome_path=/usr/bin/google-chrome \
    -chromedriver_path=/usr/bin/chromedriver \
    -frame_buffer=true \
    -staging=false \
    -browser=chrome \
    -debug=true
```

Redirecting stderr to stdout and saving it to a log-file is recommended due to
the verbosity of webdriver logs (append `2>&1 | tee my-log.txt` to the above
command).

### Running in docker

Sometimes bugs only occur in a docker-like environment. This can be difficult
to reproduce, but a first step is to run the tests inside of docker. To do
this, first start the docker container in one terminal tab:

```
./util/docker-dev/run.sh
```

Then, in another tab, we need to get the instance id of the container, exec
'bash' inside of it, and run our test:

```
source util/commands.sh
wptd_exec_it bash
user@abce84dd426d:~/wpt.fyi$
[now you can run 'make go_chrome_test', or 'go test ...' directly, etc]
```

Note that this maps the host machine's wpt.fyi checkout into docker, so any
code edits you make on the host are reflected in the container and vice-versa.

### Debugging in docker

You can use VSCode to debug the web server running in Docker. To do so, first
start the docker container in one tab:
```sh
./util/docker-dev/run.sh
```

Then, in another tab, start the web server with the `-d` flag:
```sh
./util/docker-dev/web_server.sh -d
```

Afterwards, you can go to VSCode in the Run and Debug tab and click the play button
next to the Web server launch configuration. You can then set breakpoints, inspect
variables, pause and resume execution as usual.