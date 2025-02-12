/*
 * MIT License
 *
 * Copyright (c) 2022-2025 Arsene Tochemey Gandote
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package zapl

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/tochemey/gopack/log"
	"github.com/tochemey/gopack/requestid"
)

// DefaultLogger represents the default Log to use
// This Log wraps zap under the hood
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

// WithContext returns the Logger associated with the ctx.
// This will set the traceid, requestid and spanid in case there are
// in the context
func WithContext(ctx context.Context) log.Logger {
	return DefaultLogger.WithContext(ctx)
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
	cfg := zap.Config{
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		// copied from "zap.NewProductionEncoderConfig" with some updates
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:       "ts",
			LevelKey:      "level",
			NameKey:       "logger",
			CallerKey:     "caller",
			MessageKey:    "msg",
			StacktraceKey: "stacktrace",
			LineEnding:    zapcore.DefaultLineEnding,
			EncodeLevel:   zapcore.LowercaseLevelEncoder,

			// Custom EncodeTime function to ensure we match format and precision of historic capnslog timestamps
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Format("2006-01-02T15:04:05.000000Z0700"))
			},

			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
	// create the zap log core
	var core zapcore.Core

	// create the list of writers
	syncWriters := make([]zapcore.WriteSyncer, len(writers))
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
	zapLogger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.PanicLevel),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.AddStacktrace(zapcore.FatalLevel))

	// set the global logger
	zap.ReplaceGlobals(zapLogger)
	// create the instance of Log and returns it
	return &Log{zapLogger}
}

// Debug starts a message with debug level
func (l *Log) Debug(v ...any) {
	l.Logger.Sugar().Debug(fmt.Sprint(v...))
}

// Debugf starts a message with debug level
func (l *Log) Debugf(format string, v ...any) {
	l.Logger.Sugar().Debug(fmt.Sprintf(format, v...))
}

// Panic starts a new message with panic level. The panic() function
// is called which stops the ordinary flow of a goroutine.
func (l *Log) Panic(v ...any) {
	l.Logger.Sugar().Panic(fmt.Sprint(v...))
}

// Panicf starts a new message with panic level. The panic() function
// is called which stops the ordinary flow of a goroutine.
func (l *Log) Panicf(format string, v ...any) {
	l.Logger.Sugar().Panic(fmt.Sprintf(format, v...))
}

// Fatal starts a new message with fatal level. The os.Exit(1) function
// is called which terminates the program immediately.
func (l *Log) Fatal(v ...any) {
	l.Logger.Sugar().Fatal(fmt.Sprint(v...))
}

// Fatalf starts a new message with fatal level. The os.Exit(1) function
// is called which terminates the program immediately.
func (l *Log) Fatalf(format string, v ...any) {
	l.Logger.Sugar().Fatal(fmt.Sprintf(format, v...))
}

// Error starts a new message with error level.
func (l *Log) Error(v ...any) {
	l.Logger.Sugar().Error(fmt.Sprint(v...))
}

// Errorf starts a new message with error level.
func (l *Log) Errorf(format string, v ...any) {
	l.Logger.Sugar().Error(fmt.Sprintf(format, v...))
}

// Warn starts a new message with warn level
func (l *Log) Warn(v ...any) {
	l.Logger.Sugar().Warn(fmt.Sprint(v...))
}

// Warnf starts a new message with warn level
func (l *Log) Warnf(format string, v ...any) {
	l.Logger.Sugar().Warn(fmt.Sprintf(format, v...))
}

// Info starts a message with info level
func (l *Log) Info(v ...any) {
	l.Logger.Sugar().Info(fmt.Sprint(v...))
}

// Infof starts a message with info level
func (l *Log) Infof(format string, v ...any) {
	l.Logger.Sugar().Info(fmt.Sprintf(format, v...))
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
