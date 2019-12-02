// +build tools

package util

import (
	// Import all the tools we use in order to track the deps in go.mod as recommended by
	// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
	// https://github.com/go-modules-by-example/index/blob/master/010_tools/README.md
	_ "github.com/gobuffalo/packr/v2/packr2"
	_ "github.com/golang/mock/mockgen"
	_ "golang.org/x/lint/golint"
)
