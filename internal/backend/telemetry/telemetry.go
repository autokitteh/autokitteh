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

func verifyConfig(cfg *Config) *Config {
	if cfg == nil {
		return Configs.Default
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = Configs.Default.ServiceName
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = Configs.Default.Endpoint
	}
	return cfg
}

type Labels map[string]string

func toOTELAttrs(attrs Labels) (otelAttrs []attribute.KeyValue) {
	for k, v := range attrs {
		otelAttrs = append(otelAttrs, attribute.String(k, v))
	}
	return otelAttrs
}

func WithLabels(labels Labels) api.MeasurementOption {
	var attrs []attribute.KeyValue
	for k, v := range labels {
		attrs = append(attrs, attribute.String(k, v))
	}
	return api.WithAttributes(attrs...)
}

type Telemetry struct {
	z           *zap.Logger
	enabled     bool
	serviceName string
}

func New(z *zap.Logger, cfg *Config) *Telemetry {
	cfg = verifyConfig(cfg)

	telemetry := &Telemetry{z: z, enabled: cfg.Enabled, serviceName: cfg.ServiceName}

	if !telemetry.enabled {
		z.Info("metrics are disabled")
		return telemetry
	}

	// REVIEW: encryption? GRPC?
	// REVIEW: should we use AK env or maybe standard OTEL onses?
	//         - OTEL_EXPORTER_OTLP_ENDPOINT or OTEL_EXPORTER_OTLP_METRICS_ENDPOINT
	exporter, err := otlpmetrichttp.New(
		context.Background(),
		otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithEndpoint(cfg.Endpoint),
		// otlpmetrichttp.WithURLPath("/"), // /v1/metrics by default, unless set
	)
	if err != nil {
		z.Error("failed to create metric exporter: %v", zap.Error(err))
		telemetry.enabled = false
		return telemetry
	}

	schemaURL := "https://opentelemetry.io/schemas/1.1.0"
	resourceAttrs := resource.NewWithAttributes(
		schemaURL,
		semconv.ServiceNameKey.String(telemetry.serviceName),
	)

	// REVIEW: consider using controller?
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter)),
		metric.WithResource(resourceAttrs),
	)

	otel.SetMeterProvider(meterProvider) // set global meter provider
	return telemetry
}

type NoOpCounter struct {
	api.Int64UpDownCounter
}

func (NoOpCounter) Add(context.Context, int64, ...api.AddOption) {}
func (NoOpCounter) int64UpDownCounter()                          {}

func (t *Telemetry) NewOtelUpDownCounter(name string, description string) api.Int64UpDownCounter {
	if !t.enabled {
		return NoOpCounter{}
	}
	meter := otel.GetMeterProvider().Meter(t.serviceName)
	if !strings.HasPrefix(name, t.serviceName) {
		name = fmt.Sprintf("%s.%s", t.serviceName, name)
	}
	metric, err := meter.Int64UpDownCounter(name, api.WithDescription(description))
	if err != nil {
		t.z.Error("failed to create metric", zap.String("name", name), zap.Error(err))
		// REVIEW: should we panic? kittehs.Must?
	}
	return metric
}
