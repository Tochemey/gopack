package trace

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// SpanContext helps create custom spans
func SpanContext(ctx context.Context, methodName string) (context.Context, trace.Span) {
	// Create a span
	tracer := otel.GetTracerProvider()
	spanCtx, span := tracer.Tracer("").Start(ctx, methodName)
	return spanCtx, span
}
