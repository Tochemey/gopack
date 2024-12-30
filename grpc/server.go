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
	"context"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"

	"github.com/tochemey/gopack/errorschain"
	"github.com/tochemey/gopack/otel/metric"
	"github.com/tochemey/gopack/otel/trace"
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
	// wait for interruption/termination
	notifier := make(chan os.Signal, 1)
	done := make(chan struct{}, 1)
	signal.Notify(notifier, syscall.SIGINT, syscall.SIGTERM)
	// wait for a shutdown signal, and then shutdown
	go func() {
		<-notifier
		if err := errorschain.
			New(errorschain.ReturnFirst()).
			AddError(s.cleanup(ctx)).
			AddError(s.shutdownHook(ctx)).
			Error(); err != nil {
			panic(err)
		}

		signal.Stop(notifier)
		done <- struct{}{}
	}()
	<-done
	pid := os.Getpid()
	// make sure if it is unix init process to exit
	if pid == 1 {
		os.Exit(0)
	}

	process, _ := os.FindProcess(pid)
	switch {
	case runtime.GOOS == "windows":
		_ = process.Kill()
	default:
		_ = process.Signal(syscall.SIGTERM)
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
