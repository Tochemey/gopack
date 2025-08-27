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
	"crypto/tls"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ConnectionBuilder is grpc client builder
type ConnectionBuilder struct {
	options              []grpc.DialOption
	transportCredentials credentials.TransportCredentials
}

// NewConnectionBuilder creates an instance of ConnectionBuilder
func NewConnectionBuilder() *ConnectionBuilder {
	return &ConnectionBuilder{}
}

// WithOptions set dial options
func (b *ConnectionBuilder) WithOptions(opts ...grpc.DialOption) *ConnectionBuilder {
	b.options = append(b.options, opts...)
	return b
}

// WithInsecure set the connection as insecure
func (b *ConnectionBuilder) WithInsecure() *ConnectionBuilder {
	b.options = append(b.options, grpc.WithTransportCredentials(insecure.NewCredentials()))
	return b
}

// WithKeepAliveParams set the keep alive params
// ClientParameters is used to set keepalive parameters on the client-side.
// These configure how the client will actively probe to notice when a
// connection is broken and send pings so intermediaries will be aware of the
// liveness of the connection. Make sure these parameters are set in
// coordination with the keepalive policy on the server, as incompatible
// settings can result in closing of connection.
func (b *ConnectionBuilder) WithKeepAliveParams(params keepalive.ClientParameters) *ConnectionBuilder {
	keepAlive := grpc.WithKeepaliveParams(params)
	b.options = append(b.options, keepAlive)
	return b
}

// WithUnaryInterceptors set a list of interceptors to the Grpc client for unary connection
// By default, gRPC doesn't allow one to have more than one interceptor either on the client nor on the server side.
// By using `grpc_middleware` we are able to provides convenient method to add a list of interceptors
func (b *ConnectionBuilder) WithUnaryInterceptors(interceptors ...grpc.UnaryClientInterceptor) *ConnectionBuilder {
	b.options = append(b.options, grpc.WithChainUnaryInterceptor(interceptors...))
	return b
}

// WithStreamInterceptors set a list of interceptors to the Grpc client for stream connection
// By default, gRPC doesn't allow one to have more than one interceptor either on the client nor on the server side.
// By using `grpc_middleware` we are able to provides convenient method to add a list of interceptors
func (b *ConnectionBuilder) WithStreamInterceptors(interceptors ...grpc.StreamClientInterceptor) *ConnectionBuilder {
	b.options = append(b.options, grpc.WithChainStreamInterceptor(interceptors...))
	return b
}

// WithTLS sets the client TLS configuration
func (b *ConnectionBuilder) WithTLS(config *tls.Config) *ConnectionBuilder {
	b.transportCredentials = credentials.NewTLS(config)
	return b
}

// WithDefaultUnaryInterceptors sets the default unary interceptors for the grpc server
func (b *ConnectionBuilder) WithDefaultUnaryInterceptors() *ConnectionBuilder {
	return b.WithUnaryInterceptors(
		NewRequestIDUnaryClientInterceptor(),
		NewClientMetricUnaryInterceptor(),
	).WithOptions(grpc.WithStatsHandler(NewClientTracingHandler()))
}

// WithDefaultStreamInterceptors sets the default stream interceptors for the grpc server
func (b *ConnectionBuilder) WithDefaultStreamInterceptors() *ConnectionBuilder {
	return b.WithStreamInterceptors(
		NewRequestIDStreamClientInterceptor(),
		NewClientMetricStreamInterceptor(),
	).WithOptions(grpc.WithStatsHandler(NewClientTracingHandler()))
}

// Conn returns the client connection to the server
func (b *ConnectionBuilder) Conn(addr string) (*grpc.ClientConn, error) {
	if addr == "" {
		return nil, fmt.Errorf("target connection parameter missing. address = %s", addr)
	}
	cc, err := grpc.NewClient(addr, b.options...)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to client. address = %s. error = %+v", addr, err)
	}
	return cc, nil
}

// TLSConn returns client connection to the server
func (b *ConnectionBuilder) TLSConn(addr string) (*grpc.ClientConn, error) {
	b.options = append(b.options, grpc.WithTransportCredentials(b.transportCredentials))
	if addr == "" {
		return nil, fmt.Errorf("target connection parameter missing. address = %s", addr)
	}
	cc, err := grpc.NewClient(addr, b.options...)
	if err != nil {
		return nil, fmt.Errorf("failed to get tls conn. Unable to connect to client. address = %s: %w", addr, err)
	}
	return cc, nil
}

// DefaultConn return a grpc client connection
func DefaultConn(addr string) (*grpc.ClientConn, error) {
	// create the client builder
	clientBuilder := NewConnectionBuilder().
		WithDefaultUnaryInterceptors().
		WithDefaultStreamInterceptors().
		WithInsecure().
		WithKeepAliveParams(keepalive.ClientParameters{
			Time:                1200 * time.Second,
			PermitWithoutStream: true,
		})
	// get the gRPC client connection
	conn, err := clientBuilder.Conn(addr)
	// handle the connection error
	if err != nil {
		return nil, fmt.Errorf("failed to connect to client. address = %s: %w", addr, err)
	}
	// return the client connection created
	return conn, nil
}
