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

package trace

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

var sc = trace.NewSpanContext(trace.SpanContextConfig{
	TraceID:    [16]byte{1},
	SpanID:     [8]byte{1},
	Remote:     true,
	TraceFlags: trace.FlagsSampled,
})

func TestPassThroughSpanFromGlobalTracer(t *testing.T) {
	var called bool
	router := chi.NewRouter()
	router.Use(Middleware("foobar"))
	// The default global TracerProvider provides "pass through" spans for any
	// span context in the incoming request context.
	router.HandleFunc("/user/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		got := trace.SpanFromContext(r.Context()).SpanContext()
		assert.Equal(t, sc, got)
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/user/123", nil)
	r = r.WithContext(trace.ContextWithRemoteSpanContext(context.Background(), sc))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)
	assert.True(t, called, "failed to run test")
}

func TestPropagationWithGlobalPropagators(t *testing.T) {
	defer func(p propagation.TextMapPropagator) {
		otel.SetTextMapPropagator(p)
	}(otel.GetTextMapPropagator())

	prop := propagation.TraceContext{}
	otel.SetTextMapPropagator(prop)

	r := httptest.NewRequest("GET", "/user/123", nil)
	w := httptest.NewRecorder()

	ctx := trace.ContextWithRemoteSpanContext(context.Background(), sc)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(r.Header))

	var called bool
	router := chi.NewRouter()
	router.Use(Middleware("foobar"))
	router.HandleFunc("/user/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		span := trace.SpanFromContext(r.Context())
		assert.Equal(t, sc, span.SpanContext())
		w.WriteHeader(http.StatusOK)
	}))

	router.ServeHTTP(w, r)
	assert.True(t, called, "failed to run test")
}

func TestPropagationWithCustomPropagators(t *testing.T) {
	prop := propagation.TraceContext{}

	r := httptest.NewRequest("GET", "/user/123", nil)
	w := httptest.NewRecorder()

	ctx := trace.ContextWithRemoteSpanContext(context.Background(), sc)
	prop.Inject(ctx, propagation.HeaderCarrier(r.Header))

	var called bool
	router := chi.NewRouter()
	router.Use(Middleware("foobar", WithPropagators(prop)))
	router.HandleFunc("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		called = true
		span := trace.SpanFromContext(r.Context())
		assert.Equal(t, sc, span.SpanContext())
		w.WriteHeader(http.StatusOK)
	})

	router.ServeHTTP(w, r)
	assert.True(t, called, "failed to run test")
}

type testResponseWriter struct {
	writer http.ResponseWriter
}

func (rw *testResponseWriter) Header() http.Header {
	return rw.writer.Header()
}
func (rw *testResponseWriter) Write(b []byte) (int, error) {
	return rw.writer.Write(b)
}
func (rw *testResponseWriter) WriteHeader(statusCode int) {
	rw.writer.WriteHeader(statusCode)
}

// implement Hijacker
func (rw *testResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, nil
}

// implement Pusher
func (rw *testResponseWriter) Push(string, *http.PushOptions) error {
	return nil
}

// implement Flusher
func (rw *testResponseWriter) Flush() {
}

// implement io.ReaderFrom
func (rw *testResponseWriter) ReadFrom(io.Reader) (n int64, err error) {
	return 0, nil
}

func TestResponseWriterInterfaces(t *testing.T) {
	// make sure the recordingResponseWriter preserves interfaces implemented by the wrapped writer
	router := chi.NewRouter()
	router.Use(Middleware("foobar"))
	router.HandleFunc("/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		assert.Implements(t, (*http.Hijacker)(nil), w)
		assert.Implements(t, (*http.Pusher)(nil), w)
		assert.Implements(t, (*http.Flusher)(nil), w)
		assert.Implements(t, (*io.ReaderFrom)(nil), w)
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest("GET", "/user/123", nil)
	w := &testResponseWriter{
		writer: httptest.NewRecorder(),
	}

	router.ServeHTTP(w, r)
}

func ok(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestSDKIntegration(t *testing.T) {
	sr := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider()
	provider.RegisterSpanProcessor(sr)

	router := chi.NewRouter()
	router.Use(Middleware("foobar", WithTracerProvider(provider)))
	router.HandleFunc("/user/{id:[0-9]+}", ok)
	router.HandleFunc("/book/{title}", ok)

	r0 := httptest.NewRequest("GET", "/user/123", nil)
	r1 := httptest.NewRequest("GET", "/book/foo", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r0)
	router.ServeHTTP(w, r1)

	require.Len(t, sr.Ended(), 2)
	assertSpan(t, sr.Ended()[0],
		"/user/123",
		trace.SpanKindServer,
		attribute.String("http.server_name", "foobar"),
		attribute.Int("http.status_code", http.StatusOK),
		attribute.String("http.method", "GET"),
		attribute.String("http.target", "/user/123"),
		attribute.String("http.route", "/user/123"),
	)
	assertSpan(t, sr.Ended()[1],
		"/book/foo",
		trace.SpanKindServer,
		attribute.String("http.server_name", "foobar"),
		attribute.Int("http.status_code", http.StatusOK),
		attribute.String("http.method", "GET"),
		attribute.String("http.target", "/book/foo"),
		attribute.String("http.route", "/book/foo"),
	)
}

func assertSpan(t *testing.T, span sdktrace.ReadOnlySpan, name string, kind trace.SpanKind, attrs ...attribute.KeyValue) { // nolint
	assert.Equal(t, name, span.Name())
	assert.Equal(t, trace.SpanKindServer, span.SpanKind())

	got := make(map[attribute.Key]attribute.Value, len(span.Attributes()))
	for _, a := range span.Attributes() {
		got[a.Key] = a.Value
	}
	for _, want := range attrs {
		if !assert.Contains(t, got, want.Key) {
			continue
		}
		assert.Equal(t, got[want.Key], want.Value)
	}
}
