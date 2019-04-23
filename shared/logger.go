package shared

import (
	"context"
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

var (
	clck = cloudLoggerClientKey{}
)

// HandleWithGoogleCloudLogging handles the request with the given handler, setting the logger
// on the request's context to be a Google Cloud logging client for the given project.
func HandleWithGoogleCloudLogging(h http.HandlerFunc, project string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, err := NewAppEngineFlexContext(r, project)
		if err != nil {
			h(w, r)
			return
		}
		withLogger := r.WithContext(ctx)
		began := time.Now()

		h(w, withLogger)

		// "Parent" log event that spans the child logger's timestamps.
		if client, ok := ctx.Value(clck).(*gclog.Client); ok {
			parentLogger := client.Logger("request")
			entry := gclog.Entry{
				HTTPRequest: &gclog.HTTPRequest{
					Request: r,
					Latency: time.Now().Sub(began),
				},
			}
			if childLogger, ok := ctx.Value(DefaultLoggerCtxKey()).(*gcLogger); ok {
				entry.Severity = childLogger.maxSeverity
			}
			parentLogger.Log(entry)
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

	childLogger := client.Logger("request-events")
	ctx = context.WithValue(ctx, DefaultLoggerCtxKey(), &gcLogger{
		childLogger: childLogger,
	})
	return ctx, nil
}

type gcLogger struct {
	childLogger *gclog.Logger
	maxSeverity gclog.Severity
}

func (gcl *gcLogger) log(severity gclog.Severity, format string, params ...interface{}) {
	if severity > gcl.maxSeverity {
		gcl.maxSeverity = severity
	}
	gcl.childLogger.StandardLogger(severity).Printf(format, params...)
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
