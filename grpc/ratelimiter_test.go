package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	testpb "github.com/tochemey/gopack/test/data/test/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type mockAuthorizedLimiter struct{}

func (*mockAuthorizedLimiter) Check(context.Context) bool {
	return false
}

type mockUnAuthorizedLimiter struct{}

func (*mockUnAuthorizedLimiter) Check(context.Context) bool {
	return true
}

// testServerStream is used for unit test.
// it implements the grpc.ServerStream interface
type testServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the context for this stream.
func (s *testServerStream) Context() context.Context {
	return s.ctx
}

// SendMsg sends a message
func (s *testServerStream) SendMsg(_ interface{}) error {
	return nil
}

// RecvMsg blocks until it receives a message into m or the stream is done
func (s *testServerStream) RecvMsg(_ interface{}) error {
	return nil
}

func TestNewRateLimitUnaryServerInterceptor(t *testing.T) {
	t.Run("authorized limiter", func(t *testing.T) {
		// create an instance of the interceptor
		interceptor := NewRateLimitUnaryServerInterceptor(&mockAuthorizedLimiter{})
		// fake some error prone rpc handler
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, status.Error(codes.InvalidArgument, "bad request")
		}
		info := &grpc.UnaryServerInfo{
			FullMethod: "GetAccount",
		}
		resp, err := interceptor(nil, nil, info, handler)
		assert.Nil(t, resp)
		assert.EqualError(t, err, "rpc error: code = InvalidArgument desc = bad request")
	})

	t.Run("unauthorized limiter", func(t *testing.T) {
		// create an instance of the interceptor
		interceptor := NewRateLimitUnaryServerInterceptor(&mockUnAuthorizedLimiter{})
		// fake some error prone rpc handler
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, status.Error(codes.InvalidArgument, "bad request")
		}
		info := &grpc.UnaryServerInfo{
			FullMethod: "GetAccount",
		}
		resp, err := interceptor(nil, nil, info, handler)
		assert.Nil(t, resp)
		assert.EqualError(t, err, "rpc error: code = ResourceExhausted desc = GetAccount have been rejected by rate limiting.")
	})
}

func TestNewRateLimitStreamServerInterceptor(t *testing.T) {
	t.Run("authorized limiter", func(t *testing.T) {
		// create a test stream
		testStream := &testServerStream{ctx: context.Background()}
		testService := struct{}{}
		streamInfo := &grpc.StreamServerInfo{
			FullMethod:     "TestService.StreamMethod",
			IsServerStream: true,
		}
		// create some mock handler
		handler := func(srv interface{}, stream grpc.ServerStream) error {
			return status.Error(codes.InvalidArgument, "bad request")
		}
		// create an instance of interceptor
		err := NewRateLimitStreamServerInterceptor(&mockAuthorizedLimiter{})(testService, testStream, streamInfo, handler)
		assert.EqualError(t, err, "rpc error: code = InvalidArgument desc = bad request")
	})
	t.Run("unauthorized limiter", func(t *testing.T) {
		// create a test stream
		testStream := &testServerStream{ctx: context.Background()}
		testService := struct{}{}
		streamInfo := &grpc.StreamServerInfo{
			FullMethod:     "GetAccountStream",
			IsServerStream: true,
		}
		// create some mock handler
		handler := func(srv interface{}, stream grpc.ServerStream) error {
			return status.Error(codes.InvalidArgument, "bad request")
		}
		// create an instance of interceptor
		err := NewRateLimitStreamServerInterceptor(&mockUnAuthorizedLimiter{})(testService, testStream, streamInfo, handler)
		assert.EqualError(t, err, "rpc error: code = ResourceExhausted desc = GetAccountStream have been rejected by rate limiting.")
	})
}

func TestNewRateLimitUnaryClientInterceptor(t *testing.T) {
	t.Run("authorized limiter", func(t *testing.T) {
		// create the go context
		ctx := context.Background()
		var err error
		// create a grpc server
		builder := NewInProcessServerBuilder()
		server := builder.Build()
		server.RegisterService(func(server *grpc.Server) {
			testpb.RegisterGreeterServer(server, &MockedService{})
		})
		err = server.Start()
		assert.NoError(t, err)

		// create an instance of the interceptor
		interceptor := NewRateLimitUnaryClientInterceptor(&mockAuthorizedLimiter{})
		// create the client connection
		clientConn, err := grpc.DialContext(ctx, "localhost:50051",
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
			grpc.WithUnaryInterceptor(interceptor),
			grpc.WithContextDialer(GetBufDialer(server.GetListener())))
		if err != nil {
			t.Fatalf("Failed to dial bufnet: %v", err)
		}

		// create the test client
		client := testpb.NewGreeterClient(clientConn)
		// create the request
		request := &testpb.HelloRequest{Name: "test"}
		// handle response and error
		resp, err := client.SayHello(ctx, request)
		if err != nil {
			t.Fatalf("SayHello failed: %v", err)
		}
		assert.Equal(t, resp.Message, "This is a mocked service test")
		server.Cleanup()
		err = clientConn.Close()
		assert.NoError(t, err)
	})

	t.Run("unauthorized limiter", func(t *testing.T) {
		// create the go context
		ctx := context.Background()
		var err error
		// create a grpc server
		builder := NewInProcessServerBuilder()
		server := builder.Build()
		server.RegisterService(func(server *grpc.Server) {
			testpb.RegisterGreeterServer(server, &MockedService{})
		})
		err = server.Start()
		assert.NoError(t, err)

		// create an instance of the interceptor
		interceptor := NewRateLimitUnaryClientInterceptor(&mockUnAuthorizedLimiter{})
		// create the client connection
		clientConn, err := grpc.DialContext(ctx, "localhost:50051",
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
			grpc.WithUnaryInterceptor(interceptor),
			grpc.WithContextDialer(GetBufDialer(server.GetListener())))
		if err != nil {
			t.Fatalf("Failed to dial bufnet: %v", err)
		}
		// create the test client
		client := testpb.NewGreeterClient(clientConn)
		// create the request
		request := &testpb.HelloRequest{Name: "test"}
		// handle response and error
		resp, err := client.SayHello(ctx, request)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.EqualError(t, err, "rpc error: code = ResourceExhausted desc = /test.v1.Greeter/SayHello have been rejected by rate limiting.")

		server.Cleanup()
		err = clientConn.Close()
		assert.NoError(t, err)
	})
}

func TestNewRateLimitStreamClientInterceptor(t *testing.T) {
	t.Run("authorized rate limiter", func(t *testing.T) {
		ctx := context.TODO()
		interceptor := NewRateLimitStreamClientInterceptor(&mockAuthorizedLimiter{})
		_, err := interceptor(ctx, &grpc.StreamDesc{StreamName: "test"}, nil, "test", func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			return nil, nil
		})
		assert.NoError(t, err)
	})
	t.Run("unauthorized rate limiter", func(t *testing.T) {
		ctx := context.TODO()
		interceptor := NewRateLimitStreamClientInterceptor(&mockUnAuthorizedLimiter{})
		_, err := interceptor(ctx, &grpc.StreamDesc{StreamName: "test"}, nil, "test", func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			return nil, nil
		})
		assert.Error(t, err)
		assert.EqualError(t, err, "rpc error: code = ResourceExhausted desc = test have been rejected by rate limiting.")
	})
}

func TestNewRateLimiter(t *testing.T) {
	// create a rate limiter of 2 request per seconds
	limiter := NewRateLimiter(1, 1*time.Second)
	assert.NotNil(t, limiter)
	assert.IsType(t, &RateLimiter{}, limiter)
	// assert that cl implements the interface schedulers.Job
	var iface interface{} = limiter
	_, ok := iface.(Limiter)
	assert.True(t, ok)
}
