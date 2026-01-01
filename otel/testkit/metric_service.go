// MIT License
//
// Copyright (c) 2022-2026 Arsene Tochemey Gandote
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package testkit

import (
	"context"
	"sync"
	"time"

	collectormetricpb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	metricpb "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/grpc/metadata"
)

// MetricService implements the open-telemetry  collector metrics gRPC interface
type MetricService struct {
	collectormetricpb.UnimplementedMetricsServiceServer

	requests int
	errors   []error

	headers metadata.MD
	mu      sync.RWMutex
	storage MetricsStorage
	delay   time.Duration
}

var _ collectormetricpb.MetricsServiceServer = (*MetricService)(nil)

// GetHeaders returns the metadata
func (mms *MetricService) GetHeaders() metadata.MD {
	mms.mu.RLock()
	defer mms.mu.RUnlock()
	return mms.headers
}

// GetMetrics returns the list of metric
func (mms *MetricService) GetMetrics() []*metricpb.Metric {
	mms.mu.RLock()
	defer mms.mu.RUnlock()
	return mms.storage.GetMetrics()
}

// Export exports the metrics
func (mms *MetricService) Export(ctx context.Context, exp *collectormetricpb.ExportMetricsServiceRequest) (*collectormetricpb.ExportMetricsServiceResponse, error) {
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
