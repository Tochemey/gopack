package logger

import (
	"context"
	"fmt"

	"github.com/tochemey/gopack/logger/internal/logutil"

	"go.uber.org/zap"
)

type Logger interface {
	Error(val ...interface{})
	Errorf(template string, args ...interface{})
	Errorw(val interface{}, keysAndValues ...interface{})
	Debug(val ...interface{})
	Debugf(template string, args ...interface{})
	Debugw(msg string, keysAndValues ...interface{})
	Info(val ...interface{})
	Infof(template string, args ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warn(val ...interface{})
	Warnf(template string, args ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	WithMap(m map[string]string) Logger
	// WithCtx extracts some common properties from ctx
	// and creates new logger with these fields
	WithCtx(context.Context) Logger
	Fatal(val ...interface{})
	Fatalf(template string, args ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})
	WithFields(keysAndValues ...interface{}) Logger

	// CoreLog returns pkg logger implementation
	CoreLog() interface{}
}

type loggerOpts struct {
	reqIDExtractor   RequestIDExtractor
	tracingExtractor TracingExtractor
	encoding         string
	level            string
	// nop logger can be used in unit tests to discard all output
	nopLogger bool
}

// RequestIDExtractor extracts request id property from ctx
// pass this as an option to the logger
// to avoid depending on specific request id implementations
type RequestIDExtractor func(ctx context.Context) string

// TracingExtractor extracts tracing fields from ctx
type TracingExtractor func(ctx context.Context) map[string]string

type LoggerOption func(opt *loggerOpts)

// WithReqIDExractor sets RequestIDExtractor to logger options
func WithReqIDExractor(extractor RequestIDExtractor) LoggerOption {
	return func(opt *loggerOpts) {
		opt.reqIDExtractor = extractor
	}
}

// WithTracingExtractor sets TracingExtractor to logger options
func WithTracingExtractor(extractor TracingExtractor) LoggerOption {
	return func(opt *loggerOpts) {
		opt.tracingExtractor = extractor
	}
}

func WithLevel(level string) LoggerOption {
	return func(opt *loggerOpts) {
		opt.level = level
	}
}

func WithEncoding(encoding string) LoggerOption {
	return func(opt *loggerOpts) {
		opt.encoding = encoding
	}
}

// nop logger can be used in unit tests to discard all output
func WithNop() LoggerOption {
	return func(opt *loggerOpts) {
		opt.nopLogger = true
	}
}

func NewLogger(opts ...LoggerOption) Logger {
	lo := &loggerOpts{}

	for _, opt := range opts {
		opt(lo)
	}

	var logger *zap.Logger
	if lo.nopLogger {
		logger = zap.NewNop()
	} else {
		logger = logutil.CreateLogger(lo.level, lo.encoding)
	}

	return &loggerImpl{
		opts:    lo,
		logcore: logger,
		log:     logger.Sugar(),
	}
}

type loggerImpl struct {
	opts    *loggerOpts
	logcore *zap.Logger
	log     *zap.SugaredLogger
}

func (l *loggerImpl) CoreLog() interface{} {
	return l.logcore
}

func (l *loggerImpl) WithMap(m map[string]string) Logger {
	if len(m) > 0 {
		fields := map2fields(m)
		return l.WithFields(fields...)
	}

	return l
}

func (l *loggerImpl) Error(val ...interface{}) {
	l.log.Error(val...)
}

func (l *loggerImpl) Errorf(template string, args ...interface{}) {
	l.log.Errorf(template, args...)
}

func (l *loggerImpl) Errorw(val interface{}, keysAndValues ...interface{}) {
	msg := ""
	switch v := val.(type) {
	case error:
		msg = fmt.Sprintf("%+v", v)
	case string:
		msg = v
	default:
		msg = fmt.Sprint(v)
	}

	l.log.Errorw(msg, keysAndValues...)
}

func (l *loggerImpl) Debug(val ...interface{}) {
	l.log.Debug(val...)
}

func (l *loggerImpl) Debugf(template string, args ...interface{}) {
	l.log.Debugf(template, args...)
}

func (l *loggerImpl) Debugw(msg string, keysAndValues ...interface{}) {
	l.log.Debugw(msg, keysAndValues...)
}

func (l *loggerImpl) Info(val ...interface{}) {
	l.log.Info(val...)
}

func (l *loggerImpl) Infof(template string, args ...interface{}) {
	l.log.Infof(template, args...)
}

func (l *loggerImpl) Infow(msg string, keysAndValues ...interface{}) {
	l.log.Infow(msg, keysAndValues...)
}

func (l *loggerImpl) Warn(val ...interface{}) {
	l.log.Warn(val...)
}

func (l *loggerImpl) Warnf(template string, args ...interface{}) {
	l.log.Warnf(template, args...)
}

func (l *loggerImpl) Warnw(msg string, keysAndValues ...interface{}) {
	l.log.Warnw(msg, keysAndValues...)
}

func (l *loggerImpl) Fatal(val ...interface{}) {
	l.log.Fatal(val...)
}

func (l *loggerImpl) Fatalf(template string, args ...interface{}) {
	l.log.Fatalf(template, args...)
}

func (l *loggerImpl) Fatalw(msg string, keysAndValues ...interface{}) {
	l.log.Fatalw(msg, keysAndValues...)
}

func (l *loggerImpl) WithCtx(ctx context.Context) Logger {
	var m map[string]string
	if l.opts.reqIDExtractor != nil {
		reqid := l.opts.reqIDExtractor(ctx)
		if reqid != "" {
			m = map[string]string{
				"req_id": reqid,
			}
		}
	}

	if l.opts.tracingExtractor != nil {
		fields := l.opts.tracingExtractor(ctx)
		if fields != nil {
			if m == nil {
				m = map[string]string{}
			}

			for k, v := range fields {
				m[k] = v
			}
		}
	}

	if m != nil {
		return l.WithMap(m)
	}

	return l
}

func (l *loggerImpl) WithFields(keysAndValues ...interface{}) Logger {
	return &loggerImpl{
		log:     l.log.With(keysAndValues...),
		opts:    l.opts,
		logcore: l.logcore,
	}
}

func map2fields(m map[string]string) []interface{} {
	fields := make([]interface{}, 0, len(m))
	for k, v := range m {
		fields = append(fields, zap.Any(k, v))
	}
	return fields
}
