package telemetry

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	noop "go.opentelemetry.io/otel/metric/noop"
	sdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
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

type Telemetry struct {
	l     *zap.Logger
	cfg   Config
	attrs []attribute.KeyValue
}

func New(z *zap.Logger, cfg *Config, attrs ...string) (*Telemetry, error) {
	telemetry := &Telemetry{l: z, cfg: fixConfig(*cfg)} // just ensure that endpoint and service name are set

	if !telemetry.cfg.Enabled {
		z.Info("metrics are disabled")
		return telemetry, nil
	}

	var err error
	telemetry.attrs, err = parseAttributes(attrs)
	if err != nil {
		z.Error("parse attributes", zap.Error(err))
		return nil, err
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

// parseAttributes converts a list of key-value pairs (key1, value1, key2, value2, ...) to OpenTelemetry attributes
func parseAttributes(attrs []string) ([]attribute.KeyValue, error) {
	switch {
	case len(attrs) == 0:
		return nil, nil
	case len(attrs)%2 != 0:
		return nil, errors.New("attributes must be key-value pairs")
	}

	attributes := make([]attribute.KeyValue, 0, len(attrs)/2)

	for i := 0; i < len(attrs)-1; i += 2 {
		key := attrs[i]
		value := attrs[i+1]
		attributes = append(attributes, attribute.String(key, value))
	}

	return attributes, nil
}

// AddInt64Counter records a count measurement with optional attributes
func (t *Telemetry) AddInt64Counter(ctx context.Context, counter Counter, value int64) {
	if t == nil || !t.cfg.Enabled {
		return
	}

	counter.Add(ctx, value, metric.WithAttributes(t.attrs...))
}

// RecordInt64Histogram records a histogram measurement with optional attributes
func (t *Telemetry) RecordInt64Histogram(ctx context.Context, histogram metric.Int64Histogram, value int64) {
	if t == nil || !t.cfg.Enabled {
		return
	}

	histogram.Record(ctx, value, metric.WithAttributes(t.attrs...))
}

// Counter is shared interface for counter types.
type Counter interface {
	Add(ctx context.Context, incr int64, options ...metric.AddOption)
}

func (t *Telemetry) NewCounter(name string, description string, attrs ...string) (metric.Int64Counter, error) {
	if t == nil || !t.cfg.Enabled {
		return noop.Int64Counter{}, nil
	}
	meter := otel.GetMeterProvider().Meter(t.cfg.ServiceName)
	name = t.ensureServiceName(name)

	// Create options with description
	options := []metric.Int64CounterOption{metric.WithDescription(description)}

	// Create the counter
	metric, err := meter.Int64Counter(name, options...)
	if err != nil {
		t.l.Error("failed to create metric", zap.String("name", name), zap.Error(err))
		return noop.Int64Counter{}, err
	}
	return metric, nil
}

func (t *Telemetry) NewHistogram(name string, description string, attrs ...string) (metric.Int64Histogram, error) {
	if t == nil || !t.cfg.Enabled {
		return noop.Int64Histogram{}, nil
	}
	meter := otel.GetMeterProvider().Meter(t.cfg.ServiceName)
	name = t.ensureServiceName(name)

	// Create options with description
	options := []metric.Int64HistogramOption{metric.WithDescription(description)}

	// Create the histogram
	metric, err := meter.Int64Histogram(name, options...)
	if err != nil {
		t.l.Error("failed to create metric", zap.String("name", name), zap.Error(err))
		return noop.Int64Histogram{}, err
	}
	return metric, nil
}

func (t *Telemetry) NewUpDownCounter(name string, description string) (metric.Int64UpDownCounter, error) {
	if t == nil || !t.cfg.Enabled {
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
