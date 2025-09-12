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

// nolint
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
