package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	testpb "github.com/tochemey/gopack/test/data/test/v1"
	"google.golang.org/grpc"
)

type ClientTestSuite struct {
	suite.Suite

	server     InProcessServer
	clientConn *grpc.ClientConn
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

// SetupTest will run before each test in the suite.
func (s *ClientTestSuite) SetupTest() {
	builder := NewInProcessServerBuilder()
	s.server = builder.Build()
	s.server.RegisterService(func(server *grpc.Server) {
		testpb.RegisterGreeterServer(server, &MockedService{})
	})
	err := s.server.Start()
	s.Assert().NoError(err)
}

// TearDownTest will run after each test in the suite
func (s *ClientTestSuite) TearDownTest() {
	s.server.Cleanup()
	err := s.clientConn.Close()
	s.Assert().NoError(err)
}

func (s *ClientTestSuite) TestSayHello() {
	s.Run("with context", func() {
		ctx := context.Background()
		var err error
		clientBuilder := NewClientBuilder().
			WithInsecure().
			WithDefaultStreamInterceptors().
			WithDefaultUnaryInterceptors().
			WithBlock().
			WithOptions(grpc.WithContextDialer(GetBufDialer(s.server.GetListener())))

		s.clientConn, err = clientBuilder.ClientConn(ctx, "localhost:50051")

		if err != nil {
			s.T().Fatalf("Failed to dial bufnet: %v", err)
		}
		client := testpb.NewGreeterClient(s.clientConn)
		request := &testpb.HelloRequest{Name: "test"}
		resp, err := client.SayHello(ctx, request)
		if err != nil {
			s.T().Fatalf("SayHello failed: %v", err)
		}
		s.Assert().Equal(resp.Message, "This is a mocked service test")
	})
}
