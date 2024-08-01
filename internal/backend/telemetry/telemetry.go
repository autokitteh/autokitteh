package telemetry

import (
	"context"
	"fmt"
	"reflect"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/zap"
)

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
	m.counter.Add(context.Background(), value, api.WithAttributes(toMetricAttrs(attrs)...))
}

func (m UpDownCounter) Update(value int64, attrs map[string]string) {
	m.counter.Add(context.Background(), value, api.WithAttributes(toMetricAttrs(attrs)...))
}

func Init(z *zap.Logger) {
	exporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
	if err != nil {
		z.Error("failed to create metric exporter: %v", zap.Error(err))
	}

	schemaURL := "https://opentelemetry.io/schemas/1.1.0"
	resourceAttrs := resource.NewWithAttributes(
		schemaURL,
		// REVIEW: do we want to report anything?
		//  attribute.String("deployment.environment", "igor"),
	)

	meterProvider := metric.NewMeterProvider(
		// FIXME: add koanf config for periodic reader?
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
		metricName := fieldType.Name

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
