package telemetry

import (
	"context"
	"fmt"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	noop "go.opentelemetry.io/otel/metric/noop"
	sdk "go.opentelemetry.io/otel/sdk/metric"
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

func fixConfig(cfg Config) Config {
	if cfg.ServiceName == "" {
		cfg.ServiceName = Configs.Default.ServiceName
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = Configs.Default.Endpoint
	}
	return cfg
}

func WithLabels(args ...string) metric.MeasurementOption {
	var attrs []attribute.KeyValue
	if len(args)%2 != 0 {
		sdklogger.DPanic("invalid telemetry labels")
	}
	for i := 0; i < len(args); i += 2 {
		attrs = append(attrs, attribute.String(args[i], args[i+1]))
	}
	return metric.WithAttributes(attrs...)
}

type Telemetry struct {
	l   *zap.Logger
	cfg Config
}

func New(z *zap.Logger, cfg *Config) (*Telemetry, error) {
	telemetry := &Telemetry{l: z, cfg: fixConfig(*cfg)} // just ensure that endpoint and service name are set

	if !telemetry.cfg.Enabled {
		z.Info("metrics are disabled")
		return telemetry, nil
	}

	// TODO(ENG-1445): gRPC?
	exporter, err := otlpmetrichttp.New(
		context.Background(),
		otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithEndpoint(cfg.Endpoint),
		// metrics will be sent to ENDPOINT:/v1/Metrcis. Use WithURLPath to override
	)
	if err != nil {
		z.Error("failed to create metric exporter: %v", zap.Error(err))
		telemetry.cfg.Enabled = false
		return telemetry, err
	}

	const schemaURL = "https://opentelemetry.io/schemas/1.1.0"
	resourceAttrs := resource.NewWithAttributes(
		schemaURL,
		semconv.ServiceNameKey.String(telemetry.cfg.ServiceName),
	)

	// NOTE: do we need a better control ober batching/sending. Should we use controller?
	meterProvider := sdk.NewMeterProvider(
		sdk.WithReader(sdk.NewPeriodicReader(exporter)),
		sdk.WithResource(resourceAttrs),
	)

	otel.SetMeterProvider(meterProvider) // set global meter provider
	return telemetry, nil
}

func (t *Telemetry) ensureServiceName(name string) string {
	if !strings.HasPrefix(name, t.cfg.ServiceName) {
		name = fmt.Sprintf("%s.%s", t.cfg.ServiceName, name)
	}
	return name
}

func (t *Telemetry) NewUpDownCounter(name string, description string) (metric.Int64UpDownCounter, error) {
	if !t.cfg.Enabled {
		return noop.Int64UpDownCounter{}, nil
	}
	meter := otel.GetMeterProvider().Meter(t.cfg.ServiceName)
	name = t.ensureServiceName(name)
	metric, err := meter.Int64UpDownCounter(name, metric.WithDescription(description))
	if err != nil {
		t.l.Error("failed to create metric", zap.String("name", name), zap.Error(err))
		return noop.Int64UpDownCounter{}, err
	}
	return metric, nil
}

func (t *Telemetry) NewCounter(name string, description string) (metric.Int64Counter, error) {
	if !t.cfg.Enabled {
		return noop.Int64Counter{}, nil
	}
	meter := otel.GetMeterProvider().Meter(t.cfg.ServiceName)
	name = t.ensureServiceName(name)
	metric, err := meter.Int64Counter(name, metric.WithDescription(description))
	if err != nil {
		t.l.Error("failed to create metric", zap.String("name", name), zap.Error(err))
		return noop.Int64Counter{}, err
	}
	return metric, nil
}
