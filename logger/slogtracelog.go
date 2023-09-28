//go:build go1.21

package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

func OtelTraceIdExtractor(ctx context.Context) map[string]string {
	span := trace.SpanFromContext(ctx).SpanContext()
	traceID := ""
	if span.TraceID().IsValid() {
		traceID = span.TraceID().String()
	}
	spanID := ""
	if span.SpanID().IsValid() {
		spanID = span.SpanID().String()
	}

	if traceID == "" {
		return nil
	}

	return map[string]string{
		"traceId": traceID,
		"spanId":  spanID,
	}
}
