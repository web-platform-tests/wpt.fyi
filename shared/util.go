// Copyright 2017 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shared

import (
	"fmt"
	"log"
	"net/http"

	mapset "github.com/deckarep/golang-set"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	gaelog "google.golang.org/appengine/log"
)

// ExperimentalLabel is the implicit label present for runs marked 'experimental'.
const ExperimentalLabel = "experimental"

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

type Logger interface {
	Criticalf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warningf(format string, args ...interface{})
}

type gaeLogger struct {
	ctx context.Context
}

type stdLogger struct{}

type nilLogger struct{}

func (l gaeLogger) Criticalf(format string, args ...interface{}) {
	gaelog.Criticalf(l.ctx, format, args...)
}

func (l gaeLogger) Debugf(format string, args ...interface{}) {
	gaelog.Criticalf(l.ctx, format, args...)
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

func (l stdLogger) Criticalf(format string, args ...interface{}) {
	log.Printf("CRIT: %s", fmt.Sprintf(format, args...))
}

func (l stdLogger) Debugf(format string, args ...interface{}) {
	log.Printf("DEBG: %s", fmt.Sprintf(format, args...))
}

func (l stdLogger) Errorf(format string, args ...interface{}) {
	log.Printf("ERRO: %s", fmt.Sprintf(format, args...))
}

func (l stdLogger) Infof(format string, args ...interface{}) {
	log.Printf("INFO: %s", fmt.Sprintf(format, args...))
}

func (l stdLogger) Warningf(format string, args ...interface{}) {
	log.Printf("WARN: %s", fmt.Sprintf(format, args...))
}

func (l nilLogger) Criticalf(format string, args ...interface{}) {}

func (l nilLogger) Debugf(format string, args ...interface{}) {}

func (l nilLogger) Errorf(format string, args ...interface{}) {}

func (l nilLogger) Infof(format string, args ...interface{}) {}

func (l nilLogger) Warningf(format string, args ...interface{}) {}

// LoggerCtxKey is a key for attaching a Logger to a context.Context.
type LoggerCtxKey struct{}

var (
	gl  = gaeLogger{}
	sl  = stdLogger{}
	nl  = nilLogger{}
	lck = LoggerCtxKey{}
)

// NewGAELogger returns a Google App Engine Standard Environment logger bound to
// the given context.
func NewGAELogger(ctx context.Context) Logger {
	return gaeLogger{ctx}
}

// NewSTDLogger returns a new standard logger.
func NewSTDLogger() Logger {
	return sl
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

// NewAppEngineContext creates a new Google App Engine-based context bound to
// an http.Request.
func NewAppEngineContext(r *http.Request) context.Context {
	ctx := appengine.NewContext(r)
	ctx = context.WithValue(ctx, DefaultLoggerCtxKey(), NewGAELogger(ctx))
	return ctx
}

// NewTestContext creates a new context.Context for small tests.
func NewTestContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, DefaultLoggerCtxKey(), NewNilLogger())
	return ctx
}
