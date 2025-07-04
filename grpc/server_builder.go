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
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"github.com/tochemey/gopack/otel/trace"
)

const (
	// MaxConnectionAge is the duration a connection may exist before shutdown
	MaxConnectionAge = 600 * time.Second
	// MaxConnectionAgeGrace is the maximum duration a
	// connection will be kept alive for outstanding RPCs to complete
	MaxConnectionAgeGrace = 60 * time.Second
	// KeepAliveTime is the period after which a keepalive ping is sent on the
	// transport
	KeepAliveTime = 1200 * time.Second

	// default grpc port
	defaultGrpcPort = 50051
)

var (
	ErrMissingTraceURL         = errors.New("trace URL is not defined")
	ErrMissingServiceName      = errors.New("service name is not defined")
	ErrMsgCannotUseSameBuilder = errors.New("cannot use the same builder to build more than once")
)

// ServerBuilder helps build a grpc grpcServer
type ServerBuilder struct {
	sync.Mutex
	options           []grpc.ServerOption
	services          []serviceRegistry
	enableReflection  bool
	enableHealthCheck bool
	metricsEnabled    bool
	tracingEnabled    bool
	serviceName       string
	grpcPort          int
	grpcHost          string
	traceURL          string
	logger            log.Logger

	shutdownHook ShutdownHook
	isBuilt      bool
}

// NewServerBuilder creates an instance of ServerBuilder
func NewServerBuilder() *ServerBuilder {
	return &ServerBuilder{
		grpcPort: defaultGrpcPort,
		isBuilt:  false,
	}
}

// NewServerBuilderFromConfig returns a grpcserver.ServerBuilder given a grpc config
func NewServerBuilderFromConfig(cfg *Config) *ServerBuilder {
	// build the grpc server
	return NewServerBuilder().
		WithReflection(cfg.EnableReflection).
		WithDefaultUnaryInterceptors().
		WithDefaultStreamInterceptors().
		WithTracingEnabled(cfg.TraceEnabled).
		WithTraceURL(cfg.TraceURL).
		WithServiceName(cfg.ServiceName).
		WithPort(int(cfg.GrpcPort)).
		WithHost(cfg.GrpcHost)
}

// WithShutdownHook sets the shutdown hook
func (sb *ServerBuilder) WithShutdownHook(fn ShutdownHook) *ServerBuilder {
	sb.shutdownHook = fn
	return sb
}

// WithPort sets the grpc service port
func (sb *ServerBuilder) WithPort(port int) *ServerBuilder {
	sb.grpcPort = port
	return sb
}

// WithHost sets the grpc service host
func (sb *ServerBuilder) WithHost(host string) *ServerBuilder {
	sb.grpcHost = host
	return sb
}

// WithMetricsEnabled enable grpc metrics
func (sb *ServerBuilder) WithMetricsEnabled(enabled bool) *ServerBuilder {
	sb.metricsEnabled = enabled
	return sb
}

// WithTracingEnabled enables tracing
func (sb *ServerBuilder) WithTracingEnabled(enabled bool) *ServerBuilder {
	sb.tracingEnabled = enabled
	return sb
}

// WithTraceURL sets the tracing URL
func (sb *ServerBuilder) WithTraceURL(traceURL string) *ServerBuilder {
	sb.traceURL = traceURL
	return sb
}

// WithOption adds a grpc service option
func (sb *ServerBuilder) WithOption(o grpc.ServerOption) *ServerBuilder {
	sb.options = append(sb.options, o)
	return sb
}

// WithService registers service with gRPC grpcServer
func (sb *ServerBuilder) WithService(service serviceRegistry) *ServerBuilder {
	sb.services = append(sb.services, service)
	return sb
}

// WithServiceName sets the service name
func (sb *ServerBuilder) WithServiceName(serviceName string) *ServerBuilder {
	sb.serviceName = serviceName
	return sb
}

// WithReflection enables the reflection
// gRPC RunnableService Reflection provides information about publicly-accessible gRPC services on a grpcServer,
// and assists clients at runtime to construct RPC requests and responses without precompiled service information.
// It is used by gRPC CLI, which can be used to introspect grpcServer protos and send/receive test RPCs.
// Warning! We should not have this enabled in production
func (sb *ServerBuilder) WithReflection(enabled bool) *ServerBuilder {
	sb.enableReflection = enabled
	return sb
}

// WithHealthCheck enables the default health check service
func (sb *ServerBuilder) WithHealthCheck(enabled bool) *ServerBuilder {
	sb.enableHealthCheck = enabled
	return sb
}

// WithKeepAlive is used to set keepalive and max-age parameters on the grpcServer-side.
func (sb *ServerBuilder) WithKeepAlive(serverParams keepalive.ServerParameters) *ServerBuilder {
	keepAlive := grpc.KeepaliveParams(serverParams)
	sb.WithOption(keepAlive)
	return sb
}

// WithDefaultKeepAlive is used to set the default keep alive parameters on the grpcServer-side
func (sb *ServerBuilder) WithDefaultKeepAlive() *ServerBuilder {
	return sb.WithKeepAlive(keepalive.ServerParameters{
		MaxConnectionIdle:     0,
		MaxConnectionAge:      MaxConnectionAge,
		MaxConnectionAgeGrace: MaxConnectionAgeGrace,
		Time:                  KeepAliveTime,
		Timeout:               0,
	})
}

// WithStreamInterceptors set a list of interceptors to the Grpc grpcServer for stream connection
// By default, gRPC doesn't allow one to have more than one interceptor either on the client nor on the grpcServer side.
// By using `grpcMiddleware` we are able to provides convenient method to add a list of interceptors
func (sb *ServerBuilder) WithStreamInterceptors(interceptors ...grpc.StreamServerInterceptor) *ServerBuilder {
	chain := grpc.ChainStreamInterceptor(interceptors...)
	sb.WithOption(chain)
	return sb
}

// WithUnaryInterceptors set a list of interceptors to the Grpc grpcServer for unary connection
// By default, gRPC doesn't allow one to have more than one interceptor either on the client nor on the grpcServer side.
// By using `grpc_middleware` we are able to provides convenient method to add a list of interceptors
func (sb *ServerBuilder) WithUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) *ServerBuilder {
	chain := grpc.ChainUnaryInterceptor(interceptors...)
	sb.WithOption(chain)
	return sb
}

// WithTLSCert sets credentials for grpcServer connections
func (sb *ServerBuilder) WithTLSCert(cert *tls.Certificate) *ServerBuilder {
	sb.WithOption(grpc.Creds(credentials.NewServerTLSFromCert(cert)))
	return sb
}

// WithDefaultUnaryInterceptors sets the default unary interceptors for the grpc grpcServer
func (sb *ServerBuilder) WithDefaultUnaryInterceptors() *ServerBuilder {
	return sb.WithUnaryInterceptors(
		NewRequestIDUnaryServerInterceptor(),
		NewMetricUnaryInterceptor(),
		NewRecoveryUnaryInterceptor(),
	).WithOption(grpc.StatsHandler(NewServerTracingHandler()))
}

// WithDefaultStreamInterceptors sets the default stream interceptors for the grpc grpcServer
func (sb *ServerBuilder) WithDefaultStreamInterceptors() *ServerBuilder {
	return sb.WithStreamInterceptors(
		NewRequestIDStreamServerInterceptor(),
		NewMetricStreamInterceptor(),
		NewRecoveryStreamInterceptor(),
	).WithOption(grpc.StatsHandler(NewServerTracingHandler()))
}

// Build is responsible for building a GRPC grpcServer
func (sb *ServerBuilder) Build() (Server, error) {
	// check whether the builder has already been used
	sb.Lock()
	defer sb.Unlock()

	if sb.isBuilt {
		return nil, ErrMsgCannotUseSameBuilder
	}

	// create the grpc server
	srv := grpc.NewServer(sb.options...)

	// create the grpc server
	addr := fmt.Sprintf("%s:%d", sb.grpcHost, sb.grpcPort)
	grpcServer := &grpcServer{
		addr:         addr,
		server:       srv,
		shutdownHook: sb.shutdownHook,
	}

	// register services
	for _, service := range sb.services {
		service.RegisterService(srv)
	}

	// set reflection when enable
	if sb.enableReflection {
		reflection.Register(srv)
	}

	// register health check if enabled
	if sb.enableHealthCheck {
		grpc_health_v1.RegisterHealthServer(srv, health.NewServer())
	}

	// register tracing if enabled
	if sb.tracingEnabled {
		if sb.traceURL == "" {
			return nil, ErrMissingTraceURL
		}

		if sb.serviceName == "" {
			return nil, ErrMissingServiceName
		}
		grpcServer.traceProvider = trace.NewProvider(sb.traceURL, sb.serviceName)
	}

	// set isBuild
	sb.isBuilt = true

	return grpcServer, nil
}
