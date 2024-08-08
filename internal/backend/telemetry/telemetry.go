package telemetry

import (
	"context"
	"fmt"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.uber.org/zap"
)

type Config struct {
	Enabled     bool   `koanf:"enabled"`
	ServiceName string `koanf:"service_name"`
	Endpoint    string `koanf:"endpoint"`
}

var Configs = configset.Set[Config]{
	Default: &Config{Enabled: true, ServiceName: "ak", Endpoint: "localhost:4318"},
	Dev:     &Config{Enabled: false, ServiceName: "ak", Endpoint: "localhost:4318"},
}

func (cfg *Config) fixConfig() {
	if cfg.ServiceName == "" {
		cfg.ServiceName = Configs.Default.ServiceName
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = Configs.Default.Endpoint
	}
}

func WithLabels(args ...string) api.MeasurementOption {
	var attrs []attribute.KeyValue
	if len(args)%2 != 0 {
		args = args[:len(args)-1] // strip the last one. TODO: log?
	}
	for i := 0; i < len(args); i += 2 {
		attrs = append(attrs, attribute.String(args[i], args[i+1]))
	}
	return api.WithAttributes(attrs...)
}

type Telemetry struct {
	l           *zap.Logger
	enabled     bool
	serviceName string
}

func New(z *zap.Logger, cfg *Config) *Telemetry {
	cfg.fixConfig() // just ensure that endpoint and service name are set

	telemetry := &Telemetry{l: z, enabled: cfg.Enabled, serviceName: cfg.ServiceName}

	if !telemetry.enabled {
		z.Info("metrics are disabled")
		return telemetry
	}

	// TODO: [ENG-1445] GRPC?
	exporter, err := otlpmetrichttp.New(
		context.Background(),
		otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithEndpoint(cfg.Endpoint),
		// metrics will be sent to ENDPOINT:/v1/Metrcis. Use WithURLPath to override
	)
	if err != nil {
		z.Error("failed to create metric exporter: %v", zap.Error(err))
		telemetry.enabled = false
		return telemetry
	}

	const schemaURL = "https://opentelemetry.io/schemas/1.1.0"
	resourceAttrs := resource.NewWithAttributes(
		schemaURL,
		semconv.ServiceNameKey.String(telemetry.serviceName),
	)

	// NOTE: do we need a better control ober batching/sending. Should we use controller?
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter)),
		metric.WithResource(resourceAttrs),
	)

	otel.SetMeterProvider(meterProvider) // set global meter provider
	return telemetry
}

type NoOpMetric struct {
	api.Int64UpDownCounter
	api.Int64Counter
}

func (NoOpMetric) Add(context.Context, int64, ...api.AddOption) {}

func newMetric[T any](t *Telemetry, name string, description string,
	createFunc func(meter api.Meter, name string, description string) (T, error),
) T {
	meter := otel.GetMeterProvider().Meter(t.serviceName)
	if !strings.HasPrefix(name, t.serviceName) {
		name = fmt.Sprintf("%s.%s", t.serviceName, name)
	}
	metric, err := createFunc(meter, name, description)
	if err != nil {
		t.l.Error("failed to create metric", zap.String("name", name), zap.Error(err))
		// REVIEW: should we panic? kittehs.Must?
	}
	return metric
}

func (t *Telemetry) NewUpDownCounter(name string, description string) api.Int64UpDownCounter {
	if !t.enabled {
		return NoOpMetric{}
	}
	createFunc := func(meter api.Meter, name string, description string) (api.Int64UpDownCounter, error) {
		return meter.Int64UpDownCounter(name, api.WithDescription(description))
	}
	return newMetric(t, name, description, createFunc)
}

func (t *Telemetry) NewCounter(name string, description string) api.Int64Counter {
	if !t.enabled {
		return NoOpMetric{}
	}
	createFunc := func(meter api.Meter, name string, description string) (api.Int64Counter, error) {
		return meter.Int64Counter(name, api.WithDescription(description))
	}
	return newMetric(t, name, description, createFunc)
}
