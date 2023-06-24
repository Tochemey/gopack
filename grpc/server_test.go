package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/tochemey/gopack/otel/testkit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type serverTestSuite struct {
	suite.Suite
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestServerTestSuite(t *testing.T) {
	suite.Run(t, new(serverTestSuite))
}

func (s *serverTestSuite) TestStartAndStop() {
	// create the test context to use
	ctx := context.TODO()
	// create an oltp collector
	mockCollector, err := testkit.StartOtelCollectorWithEndpoint("127.0.0.1:4448")
	s.Assert().NoError(err)
	s.Assert().NotNil(mockCollector)

	// create a server instance
	srv, err := NewServerBuilder().
		WithPort(3001).
		WithHealthCheck(false).
		WithDefaultKeepAlive().
		WithService(new(MockedService)).
		WithDefaultUnaryInterceptors().
		WithTracingEnabled(true).
		WithMetricsEnabled(true).
		WithDefaultStreamInterceptors().
		WithServiceName("test").
		WithTraceURL("127.0.0.1:4448").
		Build()
	s.Assert().NoError(err)
	s.Assert().NotNil(srv)

	// start the service
	err = srv.Start(ctx)
	s.Require().NoError(err)

	// Dial the service
	_, err = grpc.Dial("localhost:3001",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock()) // nolint
	s.Assert().NoError(err)

	// assert that collector is up
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", "4448"), time.Second)
	s.Assert().NotNil(conn)
	s.Assert().NoError(err)

	// stop the service
	s.Assert().NoError(srv.Stop(ctx))

	// let us try to connect back to the server
	conn, err = net.DialTimeout("tcp", net.JoinHostPort("localhost", "3000"), time.Second)
	s.Assert().Error(err)
	s.Assert().Nil(conn)
}
