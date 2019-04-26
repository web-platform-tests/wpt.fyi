package shared

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	gclog "cloud.google.com/go/logging"
	log "github.com/Hexcles/logrus"
	"google.golang.org/appengine"
	gaelog "google.golang.org/appengine/log"
	mrpb "google.golang.org/genproto/googleapis/api/monitoredres"
)

// Logger is an abstract logging interface that contains an intersection of
// logrus and GAE logging functionality.
type Logger interface {
	Debugf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warningf(format string, args ...interface{})
}

// NewRequestContext creates a new  context bound to an *http.Request.
func NewRequestContext(r *http.Request) context.Context {
	ctx := appengine.NewContext(r)
	return WithLogger(ctx, log.WithFields(log.Fields{
		"request": r,
	}))
}

// SplitLogger is a logger that sends logging operations to both A and B.
type loggerMux struct {
	delegates []Logger
}

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

// NewAppEngineContext creates a new Google App Engine Standard-based
// context bound to an http.Request.
func NewAppEngineContext(r *http.Request) context.Context {
	ctx := appengine.NewContext(r)
	return WithLogger(ctx, NewGAELogger(ctx))
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

type nilLogger struct{}

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

// HandleWithGoogleCloudLogging handles the request with the given handler, setting the logger
// on the request's context to be a Google Cloud logging client for the given project.
//
// commonResource is an optional override to the monitored resource details appended to each log.
// e.g. in the Flex environment, it pays to override this value to type gae_app, to ensure finding
// logs is consistent between services.
func HandleWithGoogleCloudLogging(h http.HandlerFunc, project string, commonResource *mrpb.MonitoredResource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, err := NewAppEngineFlexContext(r, project, commonResource)
		if err != nil {
			h(w, r)
			return
		}
		h(w, r.WithContext(ctx))
	}
}

// NewAppEngineFlexContext creates a new Google App Engine Flex-based
// context, with a Google Cloud logger client bound to an http.Request.
func NewAppEngineFlexContext(r *http.Request, project string, commonResource *mrpb.MonitoredResource) (ctx context.Context, err error) {
	ctx = r.Context()
	client, err := gclog.NewClient(ctx, project)
	if err != nil {
		return nil, err
	}
	// See https://cloud.google.com/appengine/docs/flexible/go/writing-application-logs
	traceID := strings.Split(r.Header.Get("X-Cloud-Trace-Context"), "/")[0]
	if traceID != "" {
		traceID = fmt.Sprintf("projects/%s/traces/%s", project, traceID)
	}
	childLogger := client.Logger("request_log_entries", gclog.CommonResource(commonResource))
	ctx = WithLogger(ctx, &gcLogger{
		childLogger: childLogger,
		traceID:     traceID,
	})
	return ctx, nil
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
