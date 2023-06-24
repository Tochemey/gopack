package grpc

import (
	"context"
	"net"

	"google.golang.org/grpc"
)

// Server will be implemented by the grpcServer
type Server interface {
	Start(ctx context.Context)
	Stop(ctx context.Context)
	Run(ctx context.Context)
	GetListener() net.Listener
	GetServer() *grpc.Server
}
