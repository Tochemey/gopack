package metric

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

// Provider is a wrapper around the open telemetry  metric provider
type Provider struct {
	serviceName      string
	exporterEndpoint string
	exportFrequency  time.Duration

	metricProvider *metric.MeterProvider
}

// NewProvider creates a new instance of TraceProvider
func NewProvider(exporterEndPoint, serviceName string, exportFrequency time.Duration) *Provider {
	return &Provider{
		serviceName:      serviceName,
		exporterEndpoint: exporterEndPoint,
		exportFrequency:  exportFrequency,
	}
}

// Start initializes an OTLP exporter, and configures the corresponding metrics provider
func (p *Provider) Start(ctx context.Context) error {
	res, err := resource.New(ctx,
		resource.WithHost(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String(p.serviceName),
		),
	)
	if err != nil {
		return err
	}

	// Set up a trace exporter
	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(p.exporterEndpoint),
	)

	// set the metric provider
	p.metricProvider = metric.NewMeterProvider(
		metric.WithReader(
			// collects and exports metric data every 30 seconds.
			metric.NewPeriodicReader(metricExporter, metric.WithInterval(p.exportFrequency))),
		metric.WithResource(res),
	)

	otel.SetMeterProvider(p.metricProvider)
	return nil
}

// Stop will flush any remaining metrics and shut down the exporter.
func (p *Provider) Stop(ctx context.Context) error {
	return p.metricProvider.Shutdown(ctx)
}
