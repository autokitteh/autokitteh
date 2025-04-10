package telemetry

import (
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/embedded"
)

type namedMeter struct {
	embedded.Meter

	underlying metric.Meter
	svcName    string
}

func (m *namedMeter) canonizeName(name string) string {
	if !strings.HasPrefix(name, m.svcName) {
		name = fmt.Sprintf("%s.%s", name, name)
	}
	return name
}

func (m *namedMeter) Int64Counter(name string, options ...metric.Int64CounterOption) (metric.Int64Counter, error) {
	return m.underlying.Int64Counter(m.canonizeName(name), options...)
}

func (m *namedMeter) Int64UpDownCounter(name string, options ...metric.Int64UpDownCounterOption) (metric.Int64UpDownCounter, error) {
	return m.underlying.Int64UpDownCounter(m.canonizeName(name), options...)
}

func (m *namedMeter) Int64Histogram(name string, options ...metric.Int64HistogramOption) (metric.Int64Histogram, error) {
	return m.underlying.Int64Histogram(m.canonizeName(name), options...)
}

func (m *namedMeter) Int64Gauge(name string, options ...metric.Int64GaugeOption) (metric.Int64Gauge, error) {
	return m.underlying.Int64Gauge(m.canonizeName(name), options...)
}

func (m *namedMeter) Int64ObservableCounter(name string, options ...metric.Int64ObservableCounterOption) (metric.Int64ObservableCounter, error) {
	return m.underlying.Int64ObservableCounter(m.canonizeName(name), options...)
}

func (m *namedMeter) Int64ObservableUpDownCounter(name string, options ...metric.Int64ObservableUpDownCounterOption) (metric.Int64ObservableUpDownCounter, error) {
	return m.underlying.Int64ObservableUpDownCounter(m.canonizeName(name), options...)
}

func (m *namedMeter) Int64ObservableGauge(name string, options ...metric.Int64ObservableGaugeOption) (metric.Int64ObservableGauge, error) {
	return m.underlying.Int64ObservableGauge(m.canonizeName(name), options...)
}

func (m *namedMeter) Float64Counter(name string, options ...metric.Float64CounterOption) (metric.Float64Counter, error) {
	return m.underlying.Float64Counter(m.canonizeName(name), options...)
}

func (m *namedMeter) Float64UpDownCounter(name string, options ...metric.Float64UpDownCounterOption) (metric.Float64UpDownCounter, error) {
	return m.underlying.Float64UpDownCounter(m.canonizeName(name), options...)
}

func (m *namedMeter) Float64Histogram(name string, options ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
	return m.underlying.Float64Histogram(m.canonizeName(name), options...)
}

func (m *namedMeter) Float64Gauge(name string, options ...metric.Float64GaugeOption) (metric.Float64Gauge, error) {
	return m.underlying.Float64Gauge(m.canonizeName(name), options...)
}

func (m *namedMeter) Float64ObservableCounter(name string, options ...metric.Float64ObservableCounterOption) (metric.Float64ObservableCounter, error) {
	return m.underlying.Float64ObservableCounter(m.canonizeName(name), options...)
}

func (m *namedMeter) Float64ObservableUpDownCounter(name string, options ...metric.Float64ObservableUpDownCounterOption) (metric.Float64ObservableUpDownCounter, error) {
	return m.underlying.Float64ObservableUpDownCounter(m.canonizeName(name), options...)
}

func (m *namedMeter) Float64ObservableGauge(name string, options ...metric.Float64ObservableGaugeOption) (metric.Float64ObservableGauge, error) {
	return m.underlying.Float64ObservableGauge(m.canonizeName(name), options...)
}

func (m *namedMeter) RegisterCallback(f metric.Callback, instruments ...metric.Observable) (metric.Registration, error) {
	return m.underlying.RegisterCallback(f, instruments...)
}
