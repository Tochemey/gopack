//go:build go1.21

package logger

import (
	"context"
	"log/slog"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestSlogHandler(t *testing.T) {
	logMock, logs := observer.New(zapcore.DebugLevel)
	sl := slog.New(NewHandler(logMock, &HandlerOptions{}))
	sl.Info("msg")

	require.Len(t, logs.AllUntimed(), 1, "Expected exactly one entry to be logged")
	entry := logs.AllUntimed()[0]
	require.Equal(t, "msg", entry.Message, "Unexpected message")
}

func TestSlogTraceAttrs(t *testing.T) {
	exporter := tracetest.NewNoopExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
	)
	tracer := tp.Tracer("test")
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	logMock, logs := observer.New(zapcore.DebugLevel)
	sl := slog.New(NewHandler(logMock, &HandlerOptions{
		traceExtractor: OtelTraceIdExtractor,
	}))
	sl.InfoContext(ctx, "msg")

	require.Len(t, logs.AllUntimed(), 1, "Expected exactly one entry to be logged")
	entry := logs.AllUntimed()[0]
	require.Equal(t, "msg", entry.Message, "Unexpected message")

	traceFieldId := slices.IndexFunc(entry.Context, func(f zapcore.Field) bool {
		return f.Key == "traceId"
	})
	require.GreaterOrEqual(t, traceFieldId, 0, "traceId not found")
	require.Equal(t, span.SpanContext().TraceID().String(), entry.Context[traceFieldId].String, "invalid traceId value")

	spanFieldId := slices.IndexFunc(entry.Context, func(f zapcore.Field) bool {
		return f.Key == "spanId"
	})
	require.GreaterOrEqual(t, spanFieldId, 0, "spanId not found")
	require.Equal(t, span.SpanContext().SpanID().String(), entry.Context[spanFieldId].String, "invalid spanId value")
}
