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

package grpc

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/tochemey/gopack/requestid"
)

// NewRequestIDUnaryServerInterceptor creates a new request ID interceptor.
// This interceptor adds a request ID to each grpc request
// nolint
func NewRequestIDUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// create the request ID
		requestID := getServerRequestID(ctx)
		// set the context with the newly created request ID
		ctx = context.WithValue(ctx, requestid.XRequestIDKey{}, requestID)
		return handler(ctx, req)
	}
}

// NewRequestIDStreamServerInterceptor creates a new request ID interceptor.
// This interceptor adds a request ID to each grpc request
// nolint
func NewRequestIDStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()
		// create the request ID
		requestID := getServerRequestID(ctx)
		// set the context with the newly created request ID
		ctx = context.WithValue(ctx, requestid.XRequestIDKey{}, requestID)
		stream := newServerStreamWithContext(ctx, ss)
		return handler(srv, stream)
	}
}

// NewRequestIDUnaryClientInterceptor creates a new request ID unary client interceptor.
// This interceptor adds a request ID to each outgoing context
func NewRequestIDUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// make a copy of the metadata
		requestMetadata, _ := metadata.FromOutgoingContext(ctx)
		metadataCopy := requestMetadata.Copy()
		// create the request ID
		requestID := getClientRequestID(ctx)
		// set the context with the newly created request ID
		ctx = context.WithValue(ctx, requestid.XRequestIDKey{}, requestID)
		// put back the metadata that originally came in
		newCtx := metadata.NewOutgoingContext(ctx, metadataCopy)
		return invoker(newCtx, method, req, reply, cc, opts...)
	}
}

// NewRequestIDStreamClientInterceptor  creates a new request ID stream client interceptor.
// This interceptor adds a request ID to each outgoing context
func NewRequestIDStreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		// make a copy of the metadata
		requestMetadata, _ := metadata.FromOutgoingContext(ctx)
		metadataCopy := requestMetadata.Copy()
		// create the request ID
		requestID := getClientRequestID(ctx)
		// set the context with the newly created request ID
		ctx = context.WithValue(ctx, requestid.XRequestIDKey{}, requestID)
		// put back the metadata that originally came in
		newCtx := metadata.NewOutgoingContext(ctx, metadataCopy)
		return streamer(newCtx, desc, cc, method, opts...)
	}
}

// getServerRequestID returns a request ID from gRPC metadata if available in the incoming ctx.
// If the request ID is not available then it is set
func getServerRequestID(ctx context.Context) string {
	// let us check whether the request id is set in the incoming context or not
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.NewString()
	}
	// the request is set in the incoming context
	// however the request id is empty then we create a new one
	header, ok := md[requestid.XRequestIDMetadataKey]
	if !ok || len(header) == 0 {
		return uuid.NewString()
	}
	// return the found request ID
	requestID := header[0]
	if requestID == "" {
		requestID = uuid.NewString()
	}
	return requestID
}

// getClientRequestID returns a request ID from gRPC metadata if available in outgoing ctx.
// If the request ID is not available then it is set
func getClientRequestID(ctx context.Context) string {
	// let us check whether the request id is set in the incoming context or not
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return uuid.NewString()
	}
	// the request is set in the incoming context
	// however the request id is empty then we create a new one
	header, ok := md[requestid.XRequestIDMetadataKey]
	if !ok || len(header) == 0 {
		return uuid.NewString()
	}
	// return the found request ID
	requestID := header[0]
	if requestID == "" {
		requestID = uuid.NewString()
	}
	return requestID
}

// create a serverStreamWithContext wrapper around the server stream
// to be able to pass in a context
type serverStreamWithContext struct {
	grpc.ServerStream
	ctx context.Context
}

// Context return the server steam context
func (ss serverStreamWithContext) Context() context.Context {
	return ss.ctx
}

// newServerStreamWithContext returns a grpc server stream with a given context
func newServerStreamWithContext(ctx context.Context, stream grpc.ServerStream) grpc.ServerStream {
	return serverStreamWithContext{
		ServerStream: stream,
		ctx:          ctx,
	}
}
