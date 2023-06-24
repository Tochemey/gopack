package grpc

import (
	"context"

	testv1 "github.com/tochemey/gopack/test/data/test/v1"

	"google.golang.org/grpc"
)

// MockedService is only used in grpc unit tests
type MockedService struct{}

// SayHello will handle the HelloRequest and return the appropriate response
func (s *MockedService) SayHello(_ context.Context, in *testv1.HelloRequest) (*testv1.HelloReply, error) {
	return &testv1.HelloReply{Message: "This is a mocked service " + in.Name}, nil
}

func (s *MockedService) RegisterService(server *grpc.Server) {
	testv1.RegisterGreeterServer(server, s)
}
