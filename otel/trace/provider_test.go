package trace

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tochemey/gopack/otel/testkit"
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
	p := NewProvider(s.collectorEndPoint, s.serviceName)
	s.Assert().NotNil(p)
}

func (s *ProviderTestSuite) TestStartAndStop() {
	ctx := context.TODO()
	p := NewProvider(s.collectorEndPoint, s.serviceName)
	s.Assert().NotNil(p)

	// let us register the metrics provider
	err := p.Start(ctx)
	s.Assert().NoError(err)

	// let us deregister
	err = p.Stop(ctx)
	s.Assert().NoError(err)
}
