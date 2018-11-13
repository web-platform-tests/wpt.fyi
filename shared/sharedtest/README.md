# sharedtest

`sharedtest` is a folder for shared test code.

Note that when authoring tests for `shared`, while also relying on utilities
available in `sharedtest`, you'll need to put the test in the `shared_test`
package, which is a Golang convention for "black box" testing of the `shared`
package. This is because we would otherwise have a circular dependency of

    shared
    sharedtest
    shared (test)

where `shared (test)` are `_test.go` files in the `shared` package.
