package shared

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	gclog "cloud.google.com/go/logging"
	"github.com/sirupsen/logrus"
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

// NewAppEngineContext creates a new Google App Engine Standard-based
// context bound to an http.Request.
func NewAppEngineContext(r *http.Request) context.Context {
	ctx, err := newAppEngineFlexContext(r, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		return r.Context()
	}

	return ctx
}

type gcLogger struct {
	childLogger *gclog.Logger
	traceID     string
}

func (gcl *gcLogger) log(severity gclog.Severity, format string, params ...interface{}) {
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

// newAppEngineFlexContext creates a new Google App Engine Flex-based
// context, with a Google Cloud logger client bound to an http.Request.
func newAppEngineFlexContext(r *http.Request, project string) (ctx context.Context, err error) {
	if Clients.gclog == nil {
		return withLogger(r.Context(), NewNilLogger()), nil
	}

	// See https://cloud.google.com/appengine/docs/flexible/go/writing-application-logs
	traceID := strings.Split(r.Header.Get("X-Cloud-Trace-Context"), "/")[0]
	if traceID != "" {
		traceID = fmt.Sprintf("projects/%s/traces/%s", project, traceID)
	}
	ctx = withLogger(r.Context(), &gcLogger{
		childLogger: Clients.logger,
		traceID:     traceID,
	})
	return ctx, nil
}

// HandleWithGoogleCloudLogging handles the request with the given handler, setting the logger
// on the request's context to be a Google Cloud logging client for the given project.
//
// commonResource is an optional override to the monitored resource details appended to each log.
// e.g. in the Flex environment, it pays to override this value to type gae_app, to ensure finding
// logs is consistent between services.
func HandleWithGoogleCloudLogging(h http.HandlerFunc, project string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, err := newAppEngineFlexContext(r, project)
		if err != nil {
			h(w, r)
			return
		}
		h(w, r.WithContext(ctx))
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
