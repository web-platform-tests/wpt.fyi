//go:build tools

package util

import (
	// Import all the tools we use in order to track the deps in go.mod as recommended by
	// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
	// https://github.com/go-modules-by-example/index/blob/master/010_tools/README.md
	_ "go.uber.org/mock/mockgen"
	_ "golang.org/x/lint/golint"
)
