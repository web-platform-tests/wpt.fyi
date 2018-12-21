// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	mapset "github.com/deckarep/golang-set"
	log "github.com/sirupsen/logrus"
	"google.golang.org/appengine"
	gaelog "google.golang.org/appengine/log"
)

// ExperimentalLabel is the implicit label present for runs marked 'experimental'.
const ExperimentalLabel = "experimental"

// LatestSHA is a helper for the 'latest' keyword/special case.
const LatestSHA = "latest"

// StableLabel is the implicit label present for runs marked 'stable'.
const StableLabel = "stable"

// BetaLabel is the implicit label present for runs marked 'beta'.
const BetaLabel = "beta"

// MasterLabel is the implicit label present for runs marked 'master',
// i.e. run from the master branch.
const MasterLabel = "master"

// PRBaseLabel is the implicit label for running just the affected tests on a
// PR but without the changes (i.e. against the base branch).
const PRBaseLabel = "pr_base"

// PRHeadLabel is the implicit label for running just the affected tests on the
// head of a PR (with the changes).
const PRHeadLabel = "pr_head"

// ProductChannelToLabel maps known product-specific channel names
// to the wpt.fyi model's equivalent.
func ProductChannelToLabel(channel string) string {
	switch channel {
	case "release", StableLabel:
		return StableLabel
	case BetaLabel:
		return BetaLabel
	case "dev", "nightly", "preview", ExperimentalLabel:
		return ExperimentalLabel
	}
	return ""
}

// GetDefaultProducts returns the default set of products to show on wpt.fyi
func GetDefaultProducts() ProductSpecs {
	browserNames := GetDefaultBrowserNames()
	products := make(ProductSpecs, len(browserNames))
	for i, name := range browserNames {
		products[i] = ProductSpec{}
		products[i].BrowserName = name
	}
	return products
}

// ToStringSlice converts a set to a typed string slice.
func ToStringSlice(set mapset.Set) []string {
	if set == nil {
		return nil
	}
	slice := set.ToSlice()
	result := make([]string, len(slice))
	for i, item := range slice {
		result[i] = item.(string)
	}
	return result
}

// IsLatest returns whether a SHA[0:10] is empty or "latest", both
// of which are treated as looking up the latest run for each browser.
func IsLatest(sha string) bool {
	return sha == "" || sha == "latest"
}

// Logger is an abstract logging interface that contains an intersection of
// logrus and GAE logging functionality.
type Logger interface {
	Debugf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warningf(format string, args ...interface{})
}

// SplitLogger is a logger that sends logging operations to both A and B.
type loggerMux struct {
	delegates []Logger
}

type gaeLogger struct {
	ctx context.Context
}

type nilLogger struct{}

// Debugf implements formatted debug logging to both A and B.
func (lm loggerMux) Debugf(format string, args ...interface{}) {
	for _, l := range lm.delegates {
		l.Debugf(format, args...)
	}
}

// Errorf implements formatted error logging to both A and B.
func (lm loggerMux) Errorf(format string, args ...interface{}) {
	for _, l := range lm.delegates {
		l.Errorf(format, args...)
	}
}

// Infof implements formatted info logging to both A and B.
func (lm loggerMux) Infof(format string, args ...interface{}) {
	for _, l := range lm.delegates {
		l.Infof(format, args...)
	}
}

// Warningf implements formatted warning logging to both A and B.
func (lm loggerMux) Warningf(format string, args ...interface{}) {
	for _, l := range lm.delegates {
		l.Warningf(format, args...)
	}
}

func (l gaeLogger) Debugf(format string, args ...interface{}) {
	gaelog.Debugf(l.ctx, format, args...)
}

func (l gaeLogger) Errorf(format string, args ...interface{}) {
	gaelog.Errorf(l.ctx, format, args...)
}

func (l gaeLogger) Infof(format string, args ...interface{}) {
	gaelog.Infof(l.ctx, format, args...)
}

func (l gaeLogger) Warningf(format string, args ...interface{}) {
	gaelog.Warningf(l.ctx, format, args...)
}

func (l nilLogger) Debugf(format string, args ...interface{}) {}

func (l nilLogger) Errorf(format string, args ...interface{}) {}

func (l nilLogger) Infof(format string, args ...interface{}) {}

func (l nilLogger) Warningf(format string, args ...interface{}) {}

// LoggerCtxKey is a key for attaching a Logger to a context.Context.
type LoggerCtxKey struct{}

var (
	gl  = gaeLogger{}
	nl  = nilLogger{}
	lck = LoggerCtxKey{}
)

// NewLoggerMux creates a multiplexing Logger that writes all log operations to
// all delegates.
func NewLoggerMux(delegates []Logger) Logger {
	if len(delegates) == 0 {
		return NewNilLogger()
	}
	return loggerMux{delegates}
}

// NewGAELogger returns a Google App Engine Standard Environment logger bound to
// the given context.
func NewGAELogger(ctx context.Context) Logger {
	return gaeLogger{ctx}
}

// NewNilLogger returns a new logger that silently ignores all Logger calls.
func NewNilLogger() Logger {
	return nl
}

// DefaultLoggerCtxKey returns the default key where a logger instance should be
// stored in a context.Context object.
func DefaultLoggerCtxKey() LoggerCtxKey {
	return lck
}

// GetLogger retrieves a non-nil Logger that is appropriate for use in ctx. If
// ctx does not provide a logger, then a nil-logger is returned.
func GetLogger(ctx context.Context) Logger {
	logger, ok := ctx.Value(DefaultLoggerCtxKey()).(Logger)
	if !ok || logger == nil {
		log.Warningf("Context without logger: %v; logs will be dropped", ctx)
		return NewNilLogger()
	}

	return logger
}

// NewAppEngineContext creates a new Google App Engine Standard-based
// context bound to an http.Request.
func NewAppEngineContext(r *http.Request) context.Context {
	ctx := appengine.NewContext(r)
	ctx = context.WithValue(ctx, DefaultLoggerCtxKey(), NewGAELogger(ctx))
	return ctx
}

// NewRequestContext creates a new  context bound to an *http.Request.
func NewRequestContext(r *http.Request) context.Context {
	ctx := appengine.NewContext(r)
	ctx = context.WithValue(ctx, DefaultLoggerCtxKey(), log.WithFields(log.Fields{
		"request": r,
	}))
	return ctx
}

// NewSetFromStringSlice is a helper for the inability to cast []string to []interface{}
func NewSetFromStringSlice(items []string) mapset.Set {
	if items == nil {
		return nil
	}
	set := mapset.NewSet()
	for _, i := range items {
		set.Add(i)
	}
	return set
}

// StringSliceContains returns true if the given slice contains the given string.
func StringSliceContains(ss []string, s string) bool {
	for _, i := range ss {
		if i == s {
			return true
		}
	}
	return false
}

// MapStringKeys returns the keys in the given string-keyed map.
func MapStringKeys(m interface{}) ([]string, error) {
	mapType := reflect.ValueOf(m)
	if mapType.Kind() != reflect.Map {
		return nil, errors.New("Interface is not a map type")
	}
	keys := mapType.MapKeys()
	strKeys := make([]string, len(keys))
	for i, key := range keys {
		var ok bool
		if strKeys[i], ok = key.Interface().(string); !ok {
			return nil, fmt.Errorf("Key %v was not a string type", key)
		}
	}
	return strKeys, nil
}

// GetResultsURL constructs the URL to the result of a single test file in the
// given run.
func GetResultsURL(run TestRun, testFile string) (resultsURL string) {
	resultsURL = run.ResultsURL
	if testFile != "" && testFile != "/" {
		// Assumes that result files are under a directory named SHA[0:10].
		resultsBase := strings.SplitAfter(resultsURL, "/"+run.Revision)[0]
		resultsPieces := strings.Split(resultsURL, "/")
		re := regexp.MustCompile("(-summary)?\\.json\\.gz$")
		product := re.ReplaceAllString(resultsPieces[len(resultsPieces)-1], "")
		resultsURL = fmt.Sprintf("%s/%s/%s", resultsBase, product, testFile)
	}
	return resultsURL
}

// CropString conditionally crops a string to the given length, if it is longer.
// Returns the original string otherwise.
func CropString(s string, i int) string {
	if len(s) <= i {
		return s
	}
	return s[:i]
}
