package telemetry

import "go.opentelemetry.io/otel/metric"

func NewUpDownCounter(name string, description string) (metric.Int64UpDownCounter, error) {
	return M().Int64UpDownCounter(name, metric.WithDescription(description))
}

func NewCounter(name string, description string) (metric.Int64Counter, error) {
	return M().Int64Counter(name, metric.WithDescription(description))
}

func NewHistogram(name string, description string) (metric.Int64Histogram, error) {
	return M().Int64Histogram(name, metric.WithDescription(description))
}

func NewGauge(name, description string) (metric.Float64Gauge, error) {
	return M().Float64Gauge(name, metric.WithDescription(description))
}
