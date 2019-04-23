package shared

import (
	"context"
	"fmt"
	"net/http"
	"time"

	gclog "cloud.google.com/go/logging"
	log "github.com/Hexcles/logrus"
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

// WithLogger is a strongly-typed setter for the logger value on the context.
func WithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, DefaultLoggerCtxKey(), logger)
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

type cloudLoggerClientKey struct{}
type traceIDKey struct{}

var (
	clck  = cloudLoggerClientKey{}
	trace = traceIDKey{}
)

// HandleWithGoogleCloudLogging handles the request with the given handler, setting the logger
// on the request's context to be a Google Cloud logging client for the given project.
func HandleWithGoogleCloudLogging(h http.HandlerFunc, project string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, err := NewAppEngineFlexContext(r, project)
		if err != nil {
			http.Error(w, "Failed to create flex context: "+err.Error(), http.StatusInternalServerError)
			// h(w, r)
			return
		}
		withLogger := r.WithContext(ctx)
		began := time.Now()

		h(w, withLogger)

		if logger, ok := ctx.Value(DefaultLoggerCtxKey()).(*gcLogger); ok {
			logger.childLogger.Flush()
		}
		// "Parent" log event that spans the child logger's timestamps.
		if client, ok := ctx.Value(clck).(*gclog.Client); ok {
			parentLogger := client.Logger("request_log")
			entry := gclog.Entry{
				HTTPRequest: &gclog.HTTPRequest{
					Request: withLogger,
					Latency: time.Now().Sub(began),
				},
				Trace: ctx.Value(trace).(string),
			}
			parentLogger.Log(entry)
			parentLogger.Flush()
		}
	}
}

// NewAppEngineFlexContext creates a new Google App Engine Flex-based
// context, with a Google Cloud logger client bound to an http.Request.
func NewAppEngineFlexContext(r *http.Request, project string) (ctx context.Context, err error) {
	ctx = r.Context()
	client, err := gclog.NewClient(ctx, project)
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, clck, client)

	traceID := r.Header.Get("X-Cloud-Trace-Context")
	ctx = context.WithValue(ctx, trace, traceID)

	childLogger := client.Logger("request_log_entries")
	ctx = WithLogger(ctx, &gcLogger{
		childLogger: childLogger,
		traceID:     traceID,
	})
	return ctx, nil
}

type gcLogger struct {
	childLogger *gclog.Logger
	maxSeverity gclog.Severity
	traceID     string
}

func (gcl *gcLogger) log(severity gclog.Severity, format string, params ...interface{}) {
	if severity > gcl.maxSeverity {
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
