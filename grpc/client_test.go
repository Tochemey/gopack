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
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"

	testpb "github.com/tochemey/gopack/test/data/test/v1"
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

		s.clientConn, err = clientBuilder.ClientConn("localhost:50051")

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
