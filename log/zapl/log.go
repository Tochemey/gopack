package zapl

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/tochemey/gopack/log"
	"github.com/tochemey/gopack/requestid"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// DefaultLogger represents the default Log to use
// This Log wraps zerolog under the hood
var DefaultLogger = New(log.DebugLevel, os.Stdout, os.Stderr)

// DiscardLogger is used not log anything
var DiscardLogger = New(log.InfoLevel, io.Discard)

// Info logs to INFO level.
func Info(v ...any) {
	DefaultLogger.Info(v...)
}

// Infof logs to INFO level
func Infof(format string, v ...any) {
	DefaultLogger.Infof(format, v...)
}

// Warning logs to the WARNING level.
func Warning(v ...any) {
	DefaultLogger.Warn(v...)
}

// Warningf logs to the WARNING level.
func Warningf(format string, v ...any) {
	DefaultLogger.Warnf(format, v...)
}

// Error logs to the ERROR level.
func Error(v ...any) {
	DefaultLogger.Error(v...)
}

// Errorf logs to the ERROR level.
func Errorf(format string, v ...any) {
	DefaultLogger.Errorf(format, v...)
}

// Fatal logs to the FATAL level followed by a call to os.Exit(1).
func Fatal(v ...any) {
	DefaultLogger.Fatal(v...)
}

// Fatalf logs to the FATAL level followed by a call to os.Exit(1).
func Fatalf(format string, v ...any) {
	DefaultLogger.Fatalf(format, v...)
}

// Panic logs to the PANIC level followed by a call to panic().
func Panic(v ...any) {
	DefaultLogger.Panic(v...)
}

// Panicf logs to the PANIC level followed by a call to panic().
func Panicf(format string, v ...any) {
	DefaultLogger.Panicf(format, v...)
}

// Log implements Logger interface with the underlying zap as
// the underlying logging library
type Log struct {
	*zap.Logger
}

// enforce compilation error
var _ log.Logger = &Log{}

// New creates an instance of Log
func New(level log.Level, writers ...io.Writer) *Log {
	// create the zap Log configuration
	cfg := zap.NewProductionConfig()
	// create the zap log core
	var core zapcore.Core

	// create the list of writers
	var syncWriters []zapcore.WriteSyncer
	for i, writer := range writers {
		syncWriters[i] = zapcore.AddSync(writer)
	}

	// set the log level
	switch level {
	case log.InfoLevel:
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(cfg.EncoderConfig),
			zap.CombineWriteSyncers(syncWriters...),
			zapcore.InfoLevel,
		)
	case log.DebugLevel:
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(cfg.EncoderConfig),
			zap.CombineWriteSyncers(syncWriters...),
			zapcore.DebugLevel,
		)
	case log.WarningLevel:
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(cfg.EncoderConfig),
			zap.CombineWriteSyncers(syncWriters...),
			zapcore.WarnLevel,
		)
	case log.ErrorLevel:
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(cfg.EncoderConfig),
			zap.CombineWriteSyncers(syncWriters...),
			zapcore.ErrorLevel,
		)
	case log.PanicLevel:
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(cfg.EncoderConfig),
			zap.CombineWriteSyncers(syncWriters...),
			zapcore.PanicLevel,
		)
	case log.FatalLevel:
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(cfg.EncoderConfig),
			zap.CombineWriteSyncers(syncWriters...),
			zapcore.FatalLevel,
		)
	default:
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(cfg.EncoderConfig),
			zap.CombineWriteSyncers(syncWriters...),
			zapcore.DebugLevel,
		)
	}
	// get the zap Log
	zapLogger := zap.New(core)
	// create the instance of Log and returns it
	return &Log{zapLogger}
}

// Debug starts a message with debug level
func (l *Log) Debug(v ...any) {
	defer l.Logger.Sync()
	l.Logger.Debug(fmt.Sprint(v...))
}

// Debugf starts a message with debug level
func (l *Log) Debugf(format string, v ...any) {
	defer l.Logger.Sync()
	l.Logger.Debug(fmt.Sprintf(format, v...))
}

// Panic starts a new message with panic level. The panic() function
// is called which stops the ordinary flow of a goroutine.
func (l *Log) Panic(v ...any) {
	defer l.Logger.Sync()
	l.Logger.Panic(fmt.Sprint(v...))
}

// Panicf starts a new message with panic level. The panic() function
// is called which stops the ordinary flow of a goroutine.
func (l *Log) Panicf(format string, v ...any) {
	defer l.Logger.Sync()
	l.Logger.Panic(fmt.Sprintf(format, v...))
}

// Fatal starts a new message with fatal level. The os.Exit(1) function
// is called which terminates the program immediately.
func (l *Log) Fatal(v ...any) {
	defer l.Logger.Sync()
	l.Logger.Fatal(fmt.Sprint(v...))
}

// Fatalf starts a new message with fatal level. The os.Exit(1) function
// is called which terminates the program immediately.
func (l *Log) Fatalf(format string, v ...any) {
	defer l.Logger.Sync()
	l.Logger.Fatal(fmt.Sprintf(format, v...))
}

// Error starts a new message with error level.
func (l *Log) Error(v ...any) {
	defer l.Logger.Sync()
	l.Logger.Error(fmt.Sprint(v...))
}

// Errorf starts a new message with error level.
func (l *Log) Errorf(format string, v ...any) {
	defer l.Logger.Sync()
	l.Logger.Error(fmt.Sprintf(format, v...))
}

// Warn starts a new message with warn level
func (l *Log) Warn(v ...any) {
	defer l.Logger.Sync()
	l.Logger.Warn(fmt.Sprint(v...))
}

// Warnf starts a new message with warn level
func (l *Log) Warnf(format string, v ...any) {
	defer l.Logger.Sync()
	l.Logger.Warn(fmt.Sprintf(format, v...))
}

// Info starts a message with info level
func (l *Log) Info(v ...any) {
	defer l.Logger.Sync()
	l.Logger.Info(fmt.Sprint(v...))
}

// Infof starts a message with info level
func (l *Log) Infof(format string, v ...any) {
	defer l.Logger.Sync()
	l.Logger.Info(fmt.Sprintf(format, v...))
}

// LogLevel returns the log level that is used
func (l *Log) LogLevel() log.Level {
	switch l.Level() {
	case zapcore.FatalLevel:
		return log.FatalLevel
	case zapcore.PanicLevel:
		return log.PanicLevel
	case zapcore.ErrorLevel:
		return log.ErrorLevel
	case zapcore.InfoLevel:
		return log.InfoLevel
	case zapcore.DebugLevel:
		return log.DebugLevel
	case zapcore.WarnLevel:
		return log.WarningLevel
	default:
		return log.InvalidLevel
	}
}

// WithContext returns the Logger associated with the ctx.
// This will set the traceid, requestid and spanid in case there are
// in the context
func (l *Log) WithContext(ctx context.Context) log.Logger {
	// define the zap core fields
	var fields []zap.Field
	// grab the request id from the context
	requestID := requestid.FromContext(ctx)
	// set the request id when it is defined
	if requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}
	// set the span and trace id when defined
	if otSpan := trace.SpanFromContext(ctx); otSpan != nil {
		// get the trace id
		traceID := otSpan.SpanContext().TraceID().String()
		// grab the span id
		spanID := otSpan.SpanContext().SpanID().String()
		fields = append(fields,
			zap.String("trace_id", traceID),
			zap.String("span_id", spanID),
		)
	}

	// set the fields when set
	if len(fields) > 0 {
		l.Logger.With(fields...)
	}
	return l
}
