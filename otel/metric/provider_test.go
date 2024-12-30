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

package metric

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/tochemey/gopack/otel/testkit"

	"github.com/stretchr/testify/suite"
	"github.com/travisjeffery/go-dynaport"
)

type ProviderTestSuite struct {
	suite.Suite

	collectorEndPoint string
	serviceName       string
	collector         testkit.TestCollector
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestProvider(t *testing.T) {
	suite.Run(t, new(ProviderTestSuite))
}

// SetupTest will run before each test in the suite.
func (s *ProviderTestSuite) SetupSuite() {
	var err error
	ports := dynaport.Get(1)
	s.collectorEndPoint = fmt.Sprintf(":%d", ports[0])
	s.serviceName = "metrics-test"
	s.collector, err = testkit.StartOtelCollectorWithEndpoint(s.collectorEndPoint)
	s.Assert().NoError(err)
}

func (s *ProviderTestSuite) TearDownSuite() {
	err := s.collector.Stop()
	s.Assert().NoError(err)
}

func (s *ProviderTestSuite) TestNewTraceProvider() {
	p := NewProvider(s.collectorEndPoint, s.serviceName, time.Second)
	s.Assert().NotNil(p)
}

func (s *ProviderTestSuite) TestStartAndStop() {
	ctx := context.TODO()
	p := NewProvider(s.collectorEndPoint, s.serviceName, time.Second)
	s.Assert().NotNil(p)

	// let us register the metrics provider
	err := p.Start(ctx)
	s.Assert().NoError(err)

	// let us deregister
	err = p.Stop(ctx)
	s.Assert().NoError(err)
}
