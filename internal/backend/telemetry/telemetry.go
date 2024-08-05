package telemetry

import (
	"context"
	"fmt"
	"reflect"
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

func toMetricAttrs(attrs map[string]string) []attribute.KeyValue {
	var result []attribute.KeyValue
	for k, v := range attrs {
		result = append(result, attribute.String(k, v))
	}
	return result
}

type Metric interface {
	Update(value int64, attrs map[string]string)
}

type Counter struct {
	counter api.Int64Counter
}

type UpDownCounter struct {
	counter api.Int64UpDownCounter
}

func (m Counter) Update(value int64, attrs map[string]string) {
	// REVIEW: should we check/report if we get a negative value
	if enabled {
		m.counter.Add(context.Background(), value, api.WithAttributes(toMetricAttrs(attrs)...))
	}
}

func (m UpDownCounter) Update(value int64, attrs map[string]string) {
	if enabled {
		m.counter.Add(context.Background(), value, api.WithAttributes(toMetricAttrs(attrs)...))
	}
}

var enabled = false

func metricName(field reflect.Value, fieldType reflect.StructField) string {
	metricName := fieldType.Name
	metricName = strings.ReplaceAll(metricName, "_", ".")
	metricName = "ak." + strings.ToLower(metricName)
	switch field.Interface().(type) {
	case Counter:
		metricName = metricName + ".counter"
	case UpDownCounter:
		metricName = metricName + ".gauge"
	}
	return metricName
}

func Init(z *zap.Logger, cfg *Config) {
	enabled = cfg.Enabled
	if !enabled {
		z.Warn("metrics are disabled")
	}

	// REVIEW: insecure? port? maybe over GRPC?
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
	}

	schemaURL := "https://opentelemetry.io/schemas/1.1.0"
	resourceAttrs := resource.NewWithAttributes(
		schemaURL,
		semconv.ServiceNameKey.String(cfg.ServiceName),
	)
	fmt.Println("with service name", cfg.ServiceName)

	// REVIEW: consider using controller?
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter)),
		metric.WithResource(resourceAttrs),
	)

	otel.SetMeterProvider(meterProvider) // set global meter provider
	meter := otel.GetMeterProvider().Meter("ak")

	metricsValue := reflect.ValueOf(&Metrics).Elem()
	metricsType := metricsValue.Type()

	for i := 0; i < metricsValue.NumField(); i++ {
		field := metricsValue.Field(i)
		fieldType := metricsType.Field(i)

		metricName := metricName(field, fieldType)
		description := api.WithDescription(metricName)

		var err error
		switch field.Interface().(type) {
		case Counter:
			int64Counter, err1 := meter.Int64Counter(metricName, description)
			if err = err1; err == nil {
				field.Set(reflect.ValueOf(Counter{counter: int64Counter}))
			}
		case UpDownCounter:
			int64UpDownCounter, err1 := meter.Int64UpDownCounter(metricName, description)
			if err = err1; err == nil {
				field.Set(reflect.ValueOf(UpDownCounter{counter: int64UpDownCounter}))
			}
		default:
			err = fmt.Errorf("not implemented metric type: %s", fieldType.Name)
		}
		if err != nil {
			z.Error("failed to create metric", zap.String("type", fieldType.Name), zap.Error(err))
		}
	}
}
