package trace

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/travisjeffery/go-dynaport"
	collectormetricpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// NewCollectorKit it has been lifted from the https://github.com/open-telemetry/opentelemetry-go with some little tweak
// This will be useful until opentelemetry go release a metrics test library
// TODO delete this file when opentelemetry-go release the metrics library and test framework
func NewCollectorKit(config *CollectorKitConfig) *CollectorKit {
	return &CollectorKit{
		metricSvc: &metricService{
			storage: NewMetricsStorage(),
			errors:  config.Errors,
		},
	}
}

type metricService struct {
	collectormetricpb.UnimplementedMetricsServiceServer

	requests int
	errors   []error

	headers metadata.MD
	mu      sync.RWMutex
	storage MetricsStorage
	delay   time.Duration
}

func (mms *metricService) getHeaders() metadata.MD {
	mms.mu.RLock()
	defer mms.mu.RUnlock()
	return mms.headers
}

func (mms *metricService) getMetrics() []*metricpb.Metric {
	mms.mu.RLock()
	defer mms.mu.RUnlock()
	return mms.storage.GetMetrics()
}

func (mms *metricService) Export(ctx context.Context, exp *collectormetricpb.ExportMetricsServiceRequest) (*collectormetricpb.ExportMetricsServiceResponse, error) {
	if mms.delay > 0 {
		time.Sleep(mms.delay)
	}

	mms.mu.Lock()
	defer func() {
		mms.requests++
		mms.mu.Unlock()
	}()

	reply := &collectormetricpb.ExportMetricsServiceResponse{}
	if mms.requests < len(mms.errors) {
		idx := mms.requests
		return reply, mms.errors[idx]
	}

	mms.headers, _ = metadata.FromIncomingContext(ctx)
	mms.storage.AddMetrics(exp)
	return reply, nil
}

// CollectorKit is an opentelemetry collector suitable for tests
type CollectorKit struct {
	metricSvc *metricService
	endpoint  string
	ln        *listener
	stopFunc  func()
	stopOnce  sync.Once
}

// GetMetrics returns the list of metrics
func (mc *CollectorKit) GetMetrics() []*metricpb.Metric {
	return mc.getMetrics()
}

// Stop the collector
func (mc *CollectorKit) Stop() error {
	return mc.stop()
}

// GetEndPoint returns the collector Endpoint
func (mc *CollectorKit) GetEndPoint() string {
	return mc.endpoint
}

// CollectorKitConfig is the collector configuration
type CollectorKitConfig struct {
	Errors   []error
	Endpoint string
}

var _ collectormetricpb.MetricsServiceServer = (*metricService)(nil)

var errAlreadyStopped = fmt.Errorf("already stopped")

func (mc *CollectorKit) stop() error {
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

func (mc *CollectorKit) GetHeaders() metadata.MD {
	return mc.metricSvc.getHeaders()
}

func (mc *CollectorKit) getMetrics() []*metricpb.Metric {
	return mc.metricSvc.getMetrics()
}

// StartCollectorKit is a helper function to create a mock Collector
func StartCollectorKit() (*CollectorKit, error) {
	// create a dynamic port
	ports := dynaport.Get(1)
	return StartCollectorKitWithEndpoint(fmt.Sprintf("localhost:%d", ports[0]))
}

// StartCollectorKitWithEndpoint creates an instance of the CollectorKit and starts it
// at the given Endpoint
func StartCollectorKitWithEndpoint(endpoint string) (*CollectorKit, error) {
	return StartCollectorKitWithConfig(&CollectorKitConfig{Endpoint: endpoint})
}

// StartCollectorKitWithConfig creates an instance of the CollectorKit and starts it given
// a mock config
func StartCollectorKitWithConfig(mockConfig *CollectorKitConfig) (*CollectorKit, error) {
	ln, err := net.Listen("tcp", mockConfig.Endpoint)
	if err != nil {
		return nil, err
	}

	srv := grpc.NewServer()
	mc := NewCollectorKit(mockConfig)
	collectormetricpb.RegisterMetricsServiceServer(srv, mc.metricSvc)
	mc.ln = newListener(ln)
	go func() {
		_ = srv.Serve((net.Listener)(mc.ln))
	}()

	mc.endpoint = ln.Addr().String()
	// srv.Stop calls Disconnect on mc.ln.
	mc.stopFunc = srv.Stop

	return mc, nil
}

type listener struct {
	closeOnce sync.Once
	wrapped   net.Listener
	C         chan struct{}
}

func newListener(wrapped net.Listener) *listener {
	return &listener{
		wrapped: wrapped,
		C:       make(chan struct{}, 1),
	}
}

func (l *listener) Close() error { return l.wrapped.Close() }

func (l *listener) Addr() net.Addr { return l.wrapped.Addr() }

// Accept waits for and returns the next connection to the listener. It will
// send a signal on l.C that a connection has been made before returning.
func (l *listener) Accept() (net.Conn, error) {
	conn, err := l.wrapped.Accept()
	if err != nil {
		// Go 1.16 exported net.ErrClosed that could clean up this check, but to
		// remain backwards compatible with previous versions of Go that we
		// support the following string evaluation is used instead to keep in line
		// with the previously recommended way to check this:
		// https://github.com/golang/go/issues/4373#issuecomment-353076799
		if strings.Contains(err.Error(), "use of closed network connection") {
			// If the listener has been closed, do not allow callers of
			// WaitForConn to wait for a connection that will never come.
			l.closeOnce.Do(func() { close(l.C) })
		}
		return conn, err
	}

	select {
	case l.C <- struct{}{}:
	default:
		// If C is full, assume nobody is listening and move on.
	}
	return conn, nil
}

// WaitForConn will wait indefinitely for a connection to be established with
// the listener before returning.
func (l *listener) WaitForConn() {
	for {
		select {
		case <-l.C:
			return
		default:
			runtime.Gosched()
		}
	}
}

// Collector is an interface that mock collectors should implement,
// so they can be used for the end-to-end testing.
// The code has been lifted from the https://github.com/open-telemetry/opentelemetry-go
// because we cannot import an internal package in go
type Collector interface {
	Stop() error
	GetMetrics() []*metricpb.Metric
	GetHeaders() metadata.MD
	GetEndPoint() string
}

// MetricsStorage stores the metrics. Mock collectors could use it to
// store metrics they have received.
type MetricsStorage struct {
	metrics []*metricpb.Metric
}

// NewMetricsStorage creates a new metrics storage.
func NewMetricsStorage() MetricsStorage {
	return MetricsStorage{}
}

// AddMetrics adds metrics to the metrics storage.
func (s *MetricsStorage) AddMetrics(request *collectormetricpb.ExportMetricsServiceRequest) {
	for _, rm := range request.GetResourceMetrics() {
		if len(rm.ScopeMetrics) > 0 {
			s.metrics = append(s.metrics, rm.ScopeMetrics[0].Metrics...)
		}
	}
}

// GetMetrics returns the stored metrics.
func (s *MetricsStorage) GetMetrics() []*metricpb.Metric {
	// copy in order to not change.
	m := make([]*metricpb.Metric, 0, len(s.metrics))
	return append(m, s.metrics...)
}
