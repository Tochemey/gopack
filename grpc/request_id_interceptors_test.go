/*
 * MIT License
 *
 * Copyright (c) 2022-2023 Tochemey
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

package grpc

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tochemey/gopack/requestid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	unaryInfo = &grpc.UnaryServerInfo{
		FullMethod: "TestService.UnaryMethod",
	}
	streamInfo = &grpc.StreamServerInfo{
		FullMethod:     "TestService.StreamMethod",
		IsServerStream: true,
	}
)

func TestNewUnaryServerInterceptor(t *testing.T) {
	t.Run("with request ID set", func(t *testing.T) {
		// create a request ID
		requestID := uuid.NewString()
		// create a unary handler
		unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
			if got, want := requestid.FromContext(ctx), requestID; got != want {
				t.Errorf("expect requestID to %q, but got %q", want, got)
			}

			if got, want := requestid.FromContext(ctx), requestID; got != want {
				t.Errorf("expect requestID to %q, but got %q", want, got)
			}
			return "output", nil
		}

		ctx := context.Background()
		md := metadata.Pairs(requestid.XRequestIDMetadataKey, requestID)
		ctx = metadata.NewIncomingContext(ctx, md)
		_, err := NewRequestIDUnaryServerInterceptor()(ctx, "xyz", unaryInfo, unaryHandler)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("with request ID is empty", func(t *testing.T) {
		// create a unary handler
		unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
			const output = "output"
			if got := requestid.FromContext(ctx); got == "" {
				t.Error("requestID must be generated by interceptor")
			}
			return output, nil
		}

		ctx := context.Background()
		md := metadata.Pairs(requestid.XRequestIDMetadataKey, "")
		ctx = metadata.NewIncomingContext(ctx, md)
		_, err := NewRequestIDUnaryServerInterceptor()(ctx, "xyz", unaryInfo, unaryHandler)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("without request ID", func(t *testing.T) {
		unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
			requestID := requestid.FromContext(ctx)
			if requestID == "" {
				t.Errorf("requestID must be generated by interceptor")
			}

			if got, want := requestid.FromContext(ctx), requestID; got != want {
				t.Errorf("expect requestID to %q, but got %q", want, got)
			}

			return "output", nil
		}

		ctx := context.Background()
		_, err := NewRequestIDUnaryServerInterceptor()(ctx, "xyz", unaryInfo, unaryHandler)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestNewStreamServerInterceptor(t *testing.T) {
	t.Run("without request ID", func(t *testing.T) {
		streamHandler := func(srv interface{}, stream grpc.ServerStream) error {
			requestID := requestid.FromContext(stream.Context())
			if requestID == "" {
				t.Errorf("requestID must be generated by interceptor")
			}

			if got, want := requestid.FromContext(stream.Context()), requestID; got != want {
				t.Errorf("expect requestID to %q, but got %q", want, got)
			}
			return nil
		}
		testService := struct{}{}
		testStream := &testServerStream{ctx: context.Background()}
		err := NewRequestIDStreamServerInterceptor()(testService, testStream, streamInfo, streamHandler)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("with request ID", func(t *testing.T) {
		// create the request ID
		requestID := uuid.NewString()
		// create the stream handler
		streamHandler := func(srv interface{}, stream grpc.ServerStream) error {
			requestID := requestid.FromContext(stream.Context())
			if requestID == "" {
				t.Errorf("requestID must be generated by interceptor")
			}

			if got, want := requestid.FromContext(stream.Context()), requestID; got != want {
				t.Errorf("expect requestID to %q, but got %q", want, got)
			}
			return nil
		}
		testService := struct{}{}
		ctx := context.Background()
		md := metadata.Pairs(requestid.XRequestIDMetadataKey, requestID)
		ctx = metadata.NewIncomingContext(ctx, md)
		testStream := &testServerStream{ctx: ctx}
		err := NewRequestIDStreamServerInterceptor()(testService, testStream, streamInfo, streamHandler)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	t.Run("with request ID is empty", func(t *testing.T) {
		// create the stream handler
		streamHandler := func(srv interface{}, stream grpc.ServerStream) error {
			requestID := requestid.FromContext(stream.Context())
			if requestID == "" {
				t.Error("requestID must be generated by interceptor")
			}
			return nil
		}
		testService := struct{}{}
		ctx := context.Background()
		md := metadata.Pairs(requestid.XRequestIDMetadataKey, "")
		ctx = metadata.NewIncomingContext(ctx, md)
		testStream := &testServerStream{ctx: ctx}
		err := NewRequestIDStreamServerInterceptor()(testService, testStream, streamInfo, streamHandler)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestFromContextWithNoKeySet(t *testing.T) {
	if rid := requestid.FromContext(context.Background()); rid != "" {
		t.Fatalf("got non-empty id from empty context")
	}
}

func TestNewUnaryClientInterceptor(t *testing.T) {
	t.Run("with request ID set", func(t *testing.T) {
		// get an instance of the interceptor
		interceptor := NewRequestIDUnaryClientInterceptor()
		// create a request ID
		requestID := uuid.NewString()
		// set the request ID
		ctx := context.Background()
		md := metadata.Pairs(requestid.XRequestIDMetadataKey, requestID)
		ctx = metadata.NewOutgoingContext(ctx, md)
		// call the interceptor
		err := interceptor(ctx, "test", "req", "reply", nil, func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			if got, want := requestid.FromContext(ctx), requestID; got != want {
				t.Errorf("expect requestID to %q, but got %q", want, got)
			}
			return nil
		})
		require.NoError(t, err)
	})
	t.Run("with empty request ID", func(t *testing.T) {
		// get an instance of the interceptor
		interceptor := NewRequestIDUnaryClientInterceptor()
		// set the request ID
		ctx := context.Background()
		md := metadata.Pairs(requestid.XRequestIDMetadataKey, "")
		ctx = metadata.NewOutgoingContext(ctx, md)
		// call the interceptor
		err := interceptor(ctx, "test", "req", "reply", nil, func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			if got := requestid.FromContext(ctx); got == "" {
				t.Error("expect requestID to be set")
			}
			return nil
		})
		require.NoError(t, err)
	})
	t.Run("with request ID not set", func(t *testing.T) {
		// get an instance of the interceptor
		interceptor := NewRequestIDUnaryClientInterceptor()
		// set the request ID
		ctx := context.Background()
		// call the interceptor
		err := interceptor(ctx, "test", "req", "reply", nil, func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			if got := requestid.FromContext(ctx); got == "" {
				t.Error("expect requestID to be set")
			}
			return nil
		})
		require.NoError(t, err)
	})
}

func TestNewStreamClientInterceptor(t *testing.T) {
	t.Run("with request ID set", func(t *testing.T) {
		// get an instance of the interceptor
		interceptor := NewRequestIDStreamClientInterceptor()
		// create a request ID
		requestID := uuid.NewString()
		// set the request ID
		ctx := context.Background()
		md := metadata.Pairs(requestid.XRequestIDMetadataKey, requestID)
		ctx = metadata.NewOutgoingContext(ctx, md)
		// call the interceptor
		_, err := interceptor(ctx, &grpc.StreamDesc{StreamName: "test"}, nil, "test", func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			if got, want := requestid.FromContext(ctx), requestID; got != want {
				t.Errorf("expect requestID to %q, but got %q", want, got)
			}
			return nil, nil
		})
		require.NoError(t, err)
	})
	t.Run("with empty request ID", func(t *testing.T) {
		// get an instance of the interceptor
		interceptor := NewRequestIDStreamClientInterceptor()
		// set the request ID
		ctx := context.Background()
		md := metadata.Pairs(requestid.XRequestIDMetadataKey, "")
		ctx = metadata.NewOutgoingContext(ctx, md)
		// call the interceptor
		_, err := interceptor(ctx, &grpc.StreamDesc{StreamName: "test"}, nil, "test", func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			if got := requestid.FromContext(ctx); got == "" {
				t.Error("expect requestID to be set")
			}
			return nil, nil
		})
		require.NoError(t, err)
	})
}
