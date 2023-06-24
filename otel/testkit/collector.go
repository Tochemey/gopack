package testkit

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/travisjeffery/go-dynaport"
	collectormetricpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// TestCollector is an interface that mock collectors should implement,
// so they can be used for the end-to-end testing.
// The code has been lifted from the https://github.com/open-telemetry/opentelemetry-go
// because we cannot import an internal package in go
type TestCollector interface {
	Stop() error
	GetMetrics() []*metricpb.Metric
	GetHeaders() metadata.MD
	GetEndPoint() string
	MetricService() *MetricService
	Listener() *Listener
	SetListener(listener *Listener)
	SetEndPoint(endpoint string)
	SetStopFn(fn func())
}

var errAlreadyStopped = fmt.Errorf("already stopped")

// TestCollectorConfig is the collector configuration
type TestCollectorConfig struct {
	Errors   []error
	Endpoint string
}

// collector is an opentelemetry collector suitable for tests
type collector struct {
	metricSvc *MetricService
	endpoint  string
	ln        *Listener
	stopFunc  func()
	stopOnce  sync.Once
}

var _ TestCollector = &collector{}

// NewTestCollector it has been lifted from the https://github.com/open-telemetry/opentelemetry-go with some little tweak
// This will be useful until opentelemetry go release a metrics test library
// TODO delete this file when opentelemetry-go release the metrics library and test framework
func NewTestCollector(config *TestCollectorConfig) TestCollector {
	return &collector{
		metricSvc: &MetricService{
			storage: NewMetricsStorage(),
			errors:  config.Errors,
		},
	}
}

// SetEndPoint sets the collector endpoint
func (mc *collector) SetEndPoint(endpoint string) {
	mc.endpoint = endpoint
}

// SetStopFn sets the collector stop function
func (mc *collector) SetStopFn(fn func()) {
	mc.stopFunc = fn
}

// SetListener the collector listener
func (mc *collector) SetListener(listener *Listener) {
	mc.ln = listener
}

// Listener returns the collector listener
func (mc *collector) Listener() *Listener {
	return mc.ln
}

// MetricService returns the collector metric service
func (mc *collector) MetricService() *MetricService {
	return mc.metricSvc
}

// GetMetrics returns the list of metrics
func (mc *collector) GetMetrics() []*metricpb.Metric {
	return mc.getMetrics()
}

// Stop the collector
func (mc *collector) Stop() error {
	return mc.stop()
}

// GetEndPoint returns the collector Endpoint
func (mc *collector) GetEndPoint() string {
	return mc.endpoint
}

func (mc *collector) GetHeaders() metadata.MD {
	return mc.metricSvc.GetHeaders()
}

// StartOtelCollector is a helper function to create a mock TestCollector
func StartOtelCollector() (TestCollector, error) {
	// create a dynamic port
	ports := dynaport.Get(1)
	return StartOtelCollectorWithEndpoint(fmt.Sprintf("localhost:%d", ports[0]))
}

// StartOtelCollectorWithEndpoint creates an instance of the collector and starts it
// at the given Endpoint
func StartOtelCollectorWithEndpoint(endpoint string) (TestCollector, error) {
	return StartOtelCollectorWithConfig(&TestCollectorConfig{Endpoint: endpoint})
}

// StartOtelCollectorWithConfig creates an instance of the collector and starts it given
// a mock config
func StartOtelCollectorWithConfig(mockConfig *TestCollectorConfig) (TestCollector, error) {
	ln, err := net.Listen("tcp", mockConfig.Endpoint)
	if err != nil {
		return nil, err
	}

	srv := grpc.NewServer()
	mc := NewTestCollector(mockConfig)
	collectormetricpb.RegisterMetricsServiceServer(srv, mc.MetricService())
	mc.SetListener(NewListener(ln))
	go func() {
		_ = srv.Serve((net.Listener)(mc.Listener()))
	}()

	mc.SetEndPoint(ln.Addr().String())
	// srv.Stop calls Disconnect on mc.ln.
	mc.SetStopFn(srv.Stop)
	return mc, nil
}

func (mc *collector) getMetrics() []*metricpb.Metric {
	return mc.metricSvc.GetMetrics()
}

func (mc *collector) stop() error {
	var err = errAlreadyStopped
	mc.stopOnce.Do(func() {
		err = nil
		if mc.stopFunc != nil {
			mc.stopFunc()
		}
	})
	// Give it sometime to shut down.
	<-time.After(160 * time.Millisecond)

	// Wait for services to finish reading/writing.
	// Getting the lock ensures the metricSvc is done flushing.
	mc.metricSvc.mu.Lock()
	defer mc.metricSvc.mu.Unlock()
	return err
}
