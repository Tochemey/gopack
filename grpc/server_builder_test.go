package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/travisjeffery/go-dynaport"
)

type builderTestSuite struct {
	suite.Suite
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestBuildTestSuite(t *testing.T) {
	suite.Run(t, new(builderTestSuite))
}

func (s *builderTestSuite) TestNewServerBuilder() {
	ports := dynaport.Get(1)
	builder := NewServerBuilder().
		WithReflection(true).
		WithDefaultKeepAlive().
		WithHealthCheck(false).
		WithPort(ports[0]).
		WithServiceName("hello").
		WithService(&MockedService{}).
		WithTraceURL("").
		WithDefaultStreamInterceptors().
		WithDefaultUnaryInterceptors().
		WithTracingEnabled(false).
		WithMetricsEnabled(false).
		WithShutdownHook(func(ctx context.Context) error {
			s.T().Log("closing...")
			return nil
		})

	s.Assert().NotNil(builder)
	srv, err := builder.Build()
	s.Assert().NoError(err)
	s.Assert().NotNil(srv)
}

func (s *builderTestSuite) TestBuild() {
	s.Run("Build should be called once", func() {
		ports := dynaport.Get(1)
		builder := NewServerBuilder().
			WithReflection(true).
			WithDefaultKeepAlive().
			WithHealthCheck(false).
			WithPort(ports[0]).
			WithServiceName("hello").
			WithService(&MockedService{}).
			WithTraceURL("").
			WithDefaultStreamInterceptors().
			WithDefaultUnaryInterceptors().
			WithTracingEnabled(false).
			WithMetricsEnabled(false).
			WithShutdownHook(func(ctx context.Context) error {
				s.T().Log("closing...")
				return nil
			})

		s.Assert().NotNil(builder)

		// call build the first time no error
		srv, err := builder.Build()
		s.Assert().NoError(err)
		s.Assert().NotNil(srv)

		// let us call build the second time
		srv, err = builder.Build()
		s.Assert().Error(err)
		s.Assert().EqualError(err, "cannot use the same builder to build more than once")
		s.Assert().Nil(srv)
	})
}
