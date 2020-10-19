package shared

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	gclog "cloud.google.com/go/logging"
	"github.com/sirupsen/logrus"
	gaelog "google.golang.org/appengine/log"
)

// Logger is an abstract logging interface that contains an intersection of
// logrus and GAE logging functionality.
type Logger interface {
	Debugf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warningf(format string, args ...interface{})
}

// LoggerCtxKey is a key for attaching a Logger to a context.Context.
type LoggerCtxKey struct{}

var lck = LoggerCtxKey{}

// DefaultLoggerCtxKey returns the default key where a logger instance should be
// stored in a context.Context object.
func DefaultLoggerCtxKey() LoggerCtxKey {
	return lck
}

// withLogger is a strongly-typed setter for the logger value on the context.
func withLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, DefaultLoggerCtxKey(), logger)
}

type nilLogger struct{}

var nl = nilLogger{}

func (l nilLogger) Debugf(format string, args ...interface{}) {}

func (l nilLogger) Errorf(format string, args ...interface{}) {}

func (l nilLogger) Infof(format string, args ...interface{}) {}

func (l nilLogger) Warningf(format string, args ...interface{}) {}

// NewNilLogger returns a new logger that silently ignores all Logger calls.
func NewNilLogger() Logger {
	return nl
}

type gaeLogger struct {
	ctx context.Context
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

// newGAELogger returns a Google App Engine Standard Environment logger bound to
// the given context.
func newGAELogger(ctx context.Context) Logger {
	return gaeLogger{ctx}
}

// NewAppEngineContext creates a new Google App Engine Standard-based
// context bound to an http.Request.
func NewAppEngineContext(r *http.Request) context.Context {
	ctx := r.Context()
	return withLogger(ctx, newGAELogger(ctx))
}

type gcLogger struct {
	childLogger *gclog.Logger
	traceID     string
	maxSeverity gclog.Severity
}

func (gcl *gcLogger) log(severity gclog.Severity, format string, params ...interface{}) {
	// "Severity levels are ordered, with numerically smaller levels treated as less severe than numerically larger levels".
	// https://pkg.go.dev/cloud.google.com/go/logging#Severity.
	if int(severity) > int(gcl.maxSeverity) {
		gcl.maxSeverity = severity
	}

	gcl.childLogger.Log(gclog.Entry{
		Severity: severity,
		Payload:  fmt.Sprintf(format, params...),
		Trace:    gcl.traceID,
	})
}

func (gcl *gcLogger) Debugf(format string, params ...interface{}) {
	gcl.log(gclog.Debug, format, params...)
}

func (gcl *gcLogger) Infof(format string, params ...interface{}) {
	gcl.log(gclog.Info, format, params...)
}

func (gcl *gcLogger) Warningf(format string, params ...interface{}) {
	gcl.log(gclog.Warning, format, params...)
}

func (gcl *gcLogger) Errorf(format string, params ...interface{}) {
	gcl.log(gclog.Error, format, params...)
}

// HandleWithGoogleCloudLogging handles the request with the given handler, setting the logger
// on the request's context to be a Google Cloud logging client for the given project.
//
// commonResource is an optional override to the monitored resource details appended to each log.
// e.g. in the Flex environment, it pays to override this value to type gae_app, to ensure finding
// logs is consistent between services.
func HandleWithGoogleCloudLogging(h http.HandlerFunc, project string, childLogger, parentLogger *gclog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// See https://cloud.google.com/appengine/docs/flexible/go/writing-application-logs
		traceID := strings.Split(r.Header.Get("X-Cloud-Trace-Context"), "/")[0]
		if traceID != "" {
			traceID = fmt.Sprintf("projects/%s/traces/%s", project, traceID)
		}

		gcl := gcLogger{
			childLogger: childLogger,
			traceID:     traceID,
			maxSeverity: gclog.Default,
		}
		h(w, r.WithContext(withLogger(r.Context(), &gcl)))

		end := time.Now()
		e := gclog.Entry{
			Timestamp: end,
			Trace:     traceID,
			Severity:  gcl.maxSeverity,
			HTTPRequest: &gclog.HTTPRequest{
				Request: r,
				// TODO(kyleju): Set Status based on w.
				Latency: end.Sub(start),
			},
		}

		parentLogger.Log(e)

	}
}

// GetLogger retrieves a non-nil Logger that is appropriate for use in ctx. If
// ctx does not provide a logger, then a nil-logger is returned.
func GetLogger(ctx context.Context) Logger {
	logger, ok := ctx.Value(DefaultLoggerCtxKey()).(Logger)
	if !ok || logger == nil {
		logrus.Warningf("Context without logger: %v; logs will be dropped", ctx)
		return NewNilLogger()
	}

	return logger
}
