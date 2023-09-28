//go:build go1.21

package logger

import (
	"context"
	"log/slog"

	"github.com/tochemey/gopack/logger/internal/logutil"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const LevelFatal slog.Level = 9

// Handler implements the slog.Handler by writing to a zap Core.
type Handler struct {
	core           zapcore.Core
	traceExtractor TraceExtractor
	name           string
}

// HandlerOptions are options for a Zap-based [slog.Handler].
type HandlerOptions struct {
	traceExtractor TraceExtractor
	LoggerName     string
}

type TraceExtractor func(context.Context) map[string]string

type slogOpts struct {
	traceExtractor TraceExtractor
	encoding       string
	level          string
}

type SlogOption func(*slogOpts)

func WithLogLevel(level string) SlogOption {
	return func(opt *slogOpts) {
		opt.level = level
	}
}

func WithLogEncoding(encoding string) SlogOption {
	return func(opt *slogOpts) {
		opt.encoding = encoding
	}
}

func WithTraceExtractor(e func(context.Context) map[string]string) SlogOption {
	return func(opt *slogOpts) {
		opt.traceExtractor = e
	}
}

func WithOtelTraceExtractor() SlogOption {
	return WithTraceExtractor(OtelTraceIdExtractor)
}

func NewSlog(opts ...SlogOption) *slog.Logger {
	lo := &slogOpts{}

	for _, opt := range opts {
		opt(lo)
	}

	ll := logutil.CreateLogger(lo.level, lo.encoding)
	return slog.New(NewHandler(ll.Core(), &HandlerOptions{
		traceExtractor: lo.traceExtractor,
	}))
}

// NewHandler builds a [Handler] that writes to the supplied [zapcore.Core]
// with the default options.
func NewHandler(core zapcore.Core, opts *HandlerOptions) *Handler {
	if opts == nil {
		opts = &HandlerOptions{}
	}
	return &Handler{
		core:           core,
		name:           opts.LoggerName,
		traceExtractor: opts.traceExtractor,
	}
}

var _ slog.Handler = (*Handler)(nil)

// groupObject holds all the Attrs saved in a slog.GroupValue.
type groupObject []slog.Attr

func (gs groupObject) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for _, attr := range gs {
		convertAttrToField(attr).AddTo(enc)
	}
	return nil
}

func convertAttrToField(attr slog.Attr) zapcore.Field {
	switch attr.Value.Kind() {
	case slog.KindBool:
		return zap.Bool(attr.Key, attr.Value.Bool())
	case slog.KindDuration:
		return zap.Duration(attr.Key, attr.Value.Duration())
	case slog.KindFloat64:
		return zap.Float64(attr.Key, attr.Value.Float64())
	case slog.KindInt64:
		return zap.Int64(attr.Key, attr.Value.Int64())
	case slog.KindString:
		return zap.String(attr.Key, attr.Value.String())
	case slog.KindTime:
		return zap.Time(attr.Key, attr.Value.Time())
	case slog.KindUint64:
		return zap.Uint64(attr.Key, attr.Value.Uint64())
	case slog.KindGroup:
		return zap.Object(attr.Key, groupObject(attr.Value.Group()))
	case slog.KindLogValuer:
		return convertAttrToField(slog.Attr{
			Key: attr.Key,
			// TODO: resolve the value in a lazy way.
			// This probably needs a new Zap field type
			// that can be resolved lazily.
			Value: attr.Value.Resolve(),
		})
	default:
		return zap.Any(attr.Key, attr.Value.Any())
	}
}

// convertSlogLevel maps slog Levels to zap Levels.
// Note that there is some room between slog levels while zap levels are continuous, so we can't 1:1 map them.
// See also https://go.googlesource.com/proposal/+/master/design/56345-structured-logging.md?pli=1#levels
func convertSlogLevel(l slog.Level) zapcore.Level {
	switch {
	case l >= LevelFatal:
		return zapcore.FatalLevel
	case l >= slog.LevelError:
		return zapcore.ErrorLevel
	case l >= slog.LevelWarn:
		return zapcore.WarnLevel
	case l >= slog.LevelInfo:
		return zapcore.InfoLevel
	default:
		return zapcore.DebugLevel
	}
}

// Enabled reports whether the handler handles records at the given level.
func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.core.Enabled(convertSlogLevel(level))
}

// Handle handles the Record.
func (h *Handler) Handle(ctx context.Context, record slog.Record) error {
	ent := zapcore.Entry{
		Level:      convertSlogLevel(record.Level),
		Time:       record.Time,
		Message:    record.Message,
		LoggerName: h.name,
	}
	ce := h.core.Check(ent, nil)
	if ce == nil {
		return nil
	}

	traceFields := h.traceFields(ctx, record)
	fields := make([]zapcore.Field, 0, record.NumAttrs()+len(traceFields))
	fields = append(fields, traceFields...)

	record.Attrs(func(attr slog.Attr) bool {
		fields = append(fields, convertAttrToField(attr))
		return true
	})

	ce.Write(fields...)
	return nil
}

func (h *Handler) traceFields(ctx context.Context, record slog.Record) []zapcore.Field {
	if h.traceExtractor == nil {
		return nil
	}
	traceFields := h.traceExtractor(ctx)
	if len(traceFields) == 0 {
		return nil
	}
	fields := make([]zapcore.Field, 0, len(traceFields))
	for key, val := range traceFields {
		fields = append(fields, zap.String(key, val))
	}
	return fields
}

// WithAttrs returns a new Handler whose attributes consist of
// both the receiver's attributes and the arguments.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	fields := make([]zapcore.Field, len(attrs))
	for i, attr := range attrs {
		fields[i] = convertAttrToField(attr)
	}
	return h.withFields(fields...)
}

// WithGroup returns a new Handler with the given group appended to
// the receiver's existing groups.
func (h *Handler) WithGroup(group string) slog.Handler {
	return h.withFields(zap.Namespace(group))
}

// withFields returns a cloned Handler with the given fields.
func (h *Handler) withFields(fields ...zapcore.Field) *Handler {
	cloned := *h
	cloned.core = h.core.With(fields)
	return &cloned
}
