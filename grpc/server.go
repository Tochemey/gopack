package grpc

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/tochemey/gopack/otel/metric"
	"github.com/tochemey/gopack/otel/trace"
	"google.golang.org/grpc"
)

// ShutdownHook is used to perform some cleaning before stopping
// the long-running grpcServer
type ShutdownHook func(ctx context.Context) error

// Server will be implemented by the grpcServer
type Server interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	AwaitTermination(ctx context.Context)
	GetListener() net.Listener
	GetServer() *grpc.Server
}

// serviceRegistry.RegisterService will be implemented by any grpc service
type serviceRegistry interface {
	RegisterService(*grpc.Server)
}

type grpcServer struct {
	addr     string
	server   *grpc.Server
	listener net.Listener

	traceProvider  *trace.Provider
	metricProvider *metric.Provider

	shutdownHook ShutdownHook
}

var _ Server = (*grpcServer)(nil)

// GetServer returns the underlying grpc.Server
// This is useful when one want to use the underlying grpc.Server
// for some registration like metrics, traces and so one
func (s *grpcServer) GetServer() *grpc.Server {
	return s.server
}

// GetListener returns the underlying tcp listener
func (s *grpcServer) GetListener() net.Listener {
	return s.listener
}

// Start the GRPC server and listen to incoming connections.
func (s *grpcServer) Start(ctx context.Context) error {
	// start the metrics
	if s.metricProvider != nil {
		// let us register the metrics
		grpcPrometheus.Register(s.GetServer())
		// starts the metrics exporter
		if err := s.metricProvider.Start(ctx); err != nil {
			return err
		}
	}

	// let us register the tracer
	if s.traceProvider != nil {
		err := s.traceProvider.Start(ctx)
		if err != nil {
			return err
		}
	}

	var err error
	s.listener, err = net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	go s.serv()
	return nil
}

// Stop will shut down gracefully the running service.
// This is very useful when one wants to control the shutdown
// without waiting for an OS signal. For a long-running process, kindly use
// AwaitTermination after Start
func (s *grpcServer) Stop(ctx context.Context) error {
	if err := s.cleanup(ctx); err != nil {
		return err
	}
	if s.shutdownHook != nil {
		err := s.shutdownHook(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// AwaitTermination makes the program wait for the signal termination
// Valid signal termination (SIGINT, SIGTERM). This function should succeed Start.
func (s *grpcServer) AwaitTermination(ctx context.Context) {
	interruptSignal := make(chan os.Signal, 1)
	signal.Notify(interruptSignal, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-interruptSignal
	if err := s.cleanup(ctx); err != nil {
		panic(err)
	}

	if s.shutdownHook != nil {
		err := s.shutdownHook(ctx)
		if err != nil {
			panic(err)
		}
	}
}

// serv makes the grpc listener ready to accept connections
func (s *grpcServer) serv() {
	if err := s.server.Serve(s.listener); err != nil {
		panic(err)
	}
}

// cleanup stops the OTLP tracer and the metrics server and gracefully shutdowns the grpc server
// It stops the server from accepting new connections and RPCs and blocks until all the pending RPCs are
// finished and closes the underlying listener.
func (s *grpcServer) cleanup(ctx context.Context) error {
	// stop the metrics grpcServer
	if s.metricProvider != nil {
		if err := s.metricProvider.Stop(ctx); err != nil {
			return err
		}
	}

	// stop the tracing service
	if s.traceProvider != nil {
		err := s.traceProvider.Stop(ctx)
		if err != nil {
			return err
		}
	}
	s.server.GracefulStop()
	return nil
}
