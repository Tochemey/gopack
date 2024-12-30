/*
 * MIT License
 *
 * Copyright (c) 2022-2024 Arsene Tochemey Gandote
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
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

// NewTracingUnaryInterceptor helps gather traces and metrics from any grpc unary server
// request. Make sure to start the TracerProvider to connect to an OLTP connector
func NewTracingUnaryInterceptor(opts ...otelgrpc.Option) grpc.UnaryServerInterceptor {
	return otelgrpc.UnaryServerInterceptor(opts...)
}

// NewTracingStreamInterceptor helps gather traces and metrics from any grpc stream server
// request. Make sure to start the TracerProvider to connect to an OLTP connector
func NewTracingStreamInterceptor(opts ...otelgrpc.Option) grpc.StreamServerInterceptor {
	return otelgrpc.StreamServerInterceptor(opts...)
}

// NewTracingClientUnaryInterceptor helps gather traces and metrics from any grpc unary client
// request. Make sure to start the TracerProvider to connect to an OLTP connector
func NewTracingClientUnaryInterceptor(opts ...otelgrpc.Option) grpc.UnaryClientInterceptor {
	return otelgrpc.UnaryClientInterceptor(opts...)
}

// NewTracingClientStreamInterceptor helps gather traces and metrics from any grpc stream client
// request. Make sure to start the TracerProvider to connect to an OLTP connector
func NewTracingClientStreamInterceptor(opts ...otelgrpc.Option) grpc.StreamClientInterceptor {
	return otelgrpc.StreamClientInterceptor(opts...)
}
