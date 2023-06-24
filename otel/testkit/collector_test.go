package testkit

import (
	"net"
	"testing"

	"github.com/stretchr/testify/suite"
	v1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
)

type OtelCollectorSuite struct {
	suite.Suite
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestOtelCollector(t *testing.T) {
	suite.Run(t, new(OtelCollectorSuite))
}

func (s *OtelCollectorSuite) TestNewOtelCollector() {
	collectorKit := NewTestCollector(&TestCollectorConfig{Endpoint: "localhost:0"})
	s.Assert().NotNil(collectorKit)
}

func (s *OtelCollectorSuite) TestGetMetrics() {
	collectorKit := NewTestCollector(&TestCollectorConfig{Endpoint: "localhost:0"})
	s.Assert().NotNil(collectorKit)
	metrics := collectorKit.GetMetrics()
	s.Assert().NotNil(metrics)
	s.Assert().Empty(metrics)
}

func (s *OtelCollectorSuite) TestGetEndPoint() {
	collectorKit := NewTestCollector(&TestCollectorConfig{Endpoint: "localhost:4774"})
	s.Assert().NotNil(collectorKit)
	endpoint := collectorKit.GetEndPoint()
	s.Assert().Empty(endpoint)
}

func (s *OtelCollectorSuite) TestStartCollectorKit() {
	collectorKit, err := StartOtelCollector()
	s.Assert().NoError(err)
	s.Assert().NotNil(collectorKit)
	endpoint := collectorKit.GetEndPoint()
	s.Assert().NotEmpty(endpoint)
	s.Assert().Contains(endpoint, "127.0.0.1")
	err = collectorKit.Stop()
	s.Assert().NoError(err)
}

func (s *OtelCollectorSuite) TestStartCollectorKitWithEndpoint() {
	collectorKit, err := StartOtelCollectorWithEndpoint("127.0.0.1:4447")
	s.Assert().NoError(err)
	s.Assert().NotNil(collectorKit)
	endpoint := collectorKit.GetEndPoint()
	s.Assert().NotEmpty(endpoint)
	s.Assert().Equal("127.0.0.1:4447", endpoint)
	err = collectorKit.Stop()
	s.Assert().NoError(err)
}

func (s *OtelCollectorSuite) TestStartCollectorKitWithConfig() {
	s.Run("valid endpoint", func() {
		collectorKit, err := StartOtelCollectorWithConfig(&TestCollectorConfig{
			Endpoint: "127.0.0.1:4447",
		})
		s.Assert().NoError(err)
		s.Assert().NotNil(collectorKit)
		endpoint := collectorKit.GetEndPoint()
		s.Assert().NotEmpty(endpoint)
		s.Assert().Equal("127.0.0.1:4447", endpoint)
		err = collectorKit.Stop()
		s.Assert().NoError(err)
	})

	s.Run("invalid endpoint", func() {
		collectorKit, err := StartOtelCollectorWithConfig(&TestCollectorConfig{
			Endpoint: "some-point",
		})
		s.Assert().Error(err)
		s.Assert().Nil(collectorKit)
	})
}

func (s *OtelCollectorSuite) TestAddMetrics() {
	s.Run("when there are some metrics", func() {
		metricStorage := NewMetricsStorage()
		metricStorage.AddMetrics(&v1.ExportMetricsServiceRequest{
			ResourceMetrics: []*metricpb.ResourceMetrics{
				{
					ScopeMetrics: []*metricpb.ScopeMetrics{
						{
							Metrics: []*metricpb.Metric{
								{
									Name:        "metric-1",
									Description: "metric-1",
									Unit:        "unit-1",
									Data:        nil,
								},
							},
						},
					},
				},
			},
		})

		s.Assert().NotEmpty(metricStorage.metrics)
		s.Assert().Equal(1, len(metricStorage.metrics))
	})

	s.Run("when there are no metrics", func() {
		metricStorage := NewMetricsStorage()
		metricStorage.AddMetrics(&v1.ExportMetricsServiceRequest{
			ResourceMetrics: []*metricpb.ResourceMetrics{},
		})

		s.Assert().Empty(metricStorage.metrics)
	})
}

func (s *OtelCollectorSuite) TestStorageGetMetrics() {
	s.Run("when there no metrics", func() {
		ms := NewMetricsStorage()
		metrics := ms.GetMetrics()
		s.Assert().Empty(metrics)
	})

	s.Run("when there some metrics", func() {
		ms := NewMetricsStorage()
		ms.metrics = []*metricpb.Metric{
			{
				Name:        "metric-1",
				Description: "metric-1",
				Unit:        "unit-1",
				Data:        nil,
			},
			{
				Name:        "metric-2",
				Description: "metric-2",
				Unit:        "unit-2",
				Data:        nil,
			},
		}

		metrics := ms.GetMetrics()
		s.Assert().NotEmpty(metrics)
		s.Assert().Equal(2, len(metrics))
	})
}

func (s *OtelCollectorSuite) TestListener() {
	ln, err := net.Listen("tcp", "localhost:50051")
	s.Assert().NoError(err)

	lnr := NewListener(ln)
	s.Assert().NotNil(lnr)

	addr := lnr.Addr()
	s.Assert().NotNil(addr)

	err = lnr.Close()
	s.Assert().NoError(err)
}
