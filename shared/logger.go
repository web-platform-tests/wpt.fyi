package shared

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

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

type gcLogger struct {
	logger      *gclog.Logger
	traceID     string
	maxSeverity gclog.Severity
}

func (gcl *gcLogger) log(severity gclog.Severity, format string, params ...interface{}) {
	// "Severity levels are ordered, with numerically smaller levels treated as less severe than numerically larger levels".
	// https://pkg.go.dev/cloud.google.com/go/logging#Severity.
	if int(severity) > int(gcl.maxSeverity) {
		gcl.maxSeverity = severity
	}

	gcl.logger.Log(gclog.Entry{
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

type responseWriter struct {
	status int
	w      http.ResponseWriter
}

func (w *responseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *responseWriter) Write(b []byte) (int, error) {
	// We must duplicate this behaviour of implicit WriteHeader here.
	// Otherwise, w.Write would call its own w.WriteHeader instead of our
	// own WriteHeader due to the lack of true polymorphism.
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.w.Write(b)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.w.WriteHeader(statusCode)
}

// HandleWithLogging handles the request with the given handler, setting the
// logger on the request's context to be either a logrus logger (when running
// locally) or a Google Cloud logger (when running on GCP).
func HandleWithLogging(h http.HandlerFunc) http.HandlerFunc {
	if isDevAppserver() {
		return func(w http.ResponseWriter, r *http.Request) {
			withLocalLogger(h, w, r)
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		withGCLogger(h, w, r, runtimeIdentity.AppID, Clients.childLogger, Clients.parentLogger)
	}
}

func withLocalLogger(h http.HandlerFunc, w http.ResponseWriter, r *http.Request) {
	rw := responseWriter{w: w}
	h(&rw, r.WithContext(withLogger(r.Context(), logrus.StandardLogger())))
	logrus.Infof("%d %s %s", rw.status, r.Method, r.URL)
}

func withGCLogger(h http.HandlerFunc, w http.ResponseWriter, r *http.Request, project string, childLogger, parentLogger *gclog.Logger) {
	// https://pkg.go.dev/cloud.google.com/go/logging#hdr-Grouping_Logs_by_Request
	start := time.Now()
	// See https://cloud.google.com/appengine/docs/flexible/go/writing-application-logs
	traceID := strings.Split(r.Header.Get("X-Cloud-Trace-Context"), "/")[0]
	if traceID != "" {
		traceID = fmt.Sprintf("projects/%s/traces/%s", project, traceID)
	}

	gcl := gcLogger{
		logger:      childLogger,
		traceID:     traceID,
		maxSeverity: gclog.Default,
	}
	rw := responseWriter{w: w}
	h(&rw, r.WithContext(withLogger(r.Context(), &gcl)))

	end := time.Now()
	e := gclog.Entry{
		Timestamp: end,
		Trace:     traceID,
		Severity:  gcl.maxSeverity,
		HTTPRequest: &gclog.HTTPRequest{
			Request: r,
			Status:  rw.status,
			Latency: end.Sub(start),
		},
	}
	parentLogger.Log(e)
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
