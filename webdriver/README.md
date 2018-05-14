# wpt.fyi Webdriver tests

This directory covers Webdriver tests for the `webapp`, written in a 3rd-party
Golang Webdriver client, [tebeka/selenium](https://github.com/tebeka/selenium).

To run the tests, from the root `wpt.fyi` directory, run:

    make go_webdriver_test

If you want to actually see the tests in action, disable the frame buffer.

    make webdriver_deps
    go test --frame_buffer=false -tags=large ./webdriver

If you want to use a custom installed location of selenium / browser / driver
binaries, the required flags are shown in [the Makefile](../Makefile)
`go_webdriver_test' rule.
