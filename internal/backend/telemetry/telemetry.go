package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	noopMetric "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	sdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"
	noopTracer "go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

type Config struct {
	Enabled     bool   `koanf:"enabled"`
	ServiceName string `koanf:"service_name"`
	Endpoint    string `koanf:"endpoint"`
	Tracing     bool   `koanf:"tracing"`
}

var Configs = configset.Set[Config]{
	Default: &Config{Enabled: true, Tracing: true, ServiceName: "ak", Endpoint: "localhost:4318"},
	Dev:     &Config{Enabled: false, Tracing: true, ServiceName: "ak", Endpoint: "localhost:4318"},
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
	l   *zap.Logger
	cfg Config
	mp  *sdk.MeterProvider
	tp  *sdktrace.TracerProvider
	ra  *resource.Resource
	tr  trace.Tracer
}

func (t *Telemetry) Tracer() trace.Tracer {
	if t == nil {
		return noopTracer.NewTracerProvider().Tracer("gashash_balash")
	}
	return t.tr
}

func (t *Telemetry) Interceptor(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := baggage.ContextWithoutBaggage(r.Context())
		ctx, span := t.tr.Start(ctx, r.RequestURI)
		defer span.End()
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func New(z *zap.Logger, cfg *Config) (*Telemetry, error) {
	telemetry := &Telemetry{l: z, cfg: fixConfig(*cfg)} // just ensure that endpoint and service name are set

	const schemaURL = "https://opentelemetry.io/schemas/1.1.0"
	telemetry.ra = resource.NewWithAttributes(
		schemaURL,
		semconv.ServiceNameKey.String(cfg.ServiceName),
	)

	telemetry.setupMetrics()
	telemetry.setupTracing()

	return telemetry, nil
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	if t.cfg.Enabled {
		if err := t.mp.Shutdown(ctx); err != nil {
			t.l.Error("failed to shutdown metric provider: %v", zap.Error(err))
		}
	}
	if t.cfg.Tracing {
		if err := t.tp.Shutdown(ctx); err != nil {
			t.l.Error("failed to shutdown trace provider: %v", zap.Error(err))
		}
	}
	return nil
}

func (t *Telemetry) setupMetrics() {
	if !t.cfg.Enabled {
		otel.SetMeterProvider(noopMetric.NewMeterProvider()) // set global meter provider
		t.mp = sdk.NewMeterProvider()
		return
	}

	// TODO(ENG-1445): gRPC?
	exporter, err := otlpmetrichttp.New(
		context.Background(),
		otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithEndpoint(t.cfg.Endpoint),
		// metrics will be sent to ENDPOINT:/v1/Metrcis. Use WithURLPath to override
	)
	if err != nil {
		return
	}

	// NOTE: do we need a better control ober batching/sending. Should we use controller?
	meterProvider := sdk.NewMeterProvider(
		sdk.WithReader(sdk.NewPeriodicReader(exporter)),
		sdk.WithResource(t.ra),
	)

	otel.SetMeterProvider(meterProvider) // set global meter provider

	t.mp = meterProvider
}

func (t *Telemetry) setupTracing() {
	if !t.cfg.Tracing {
		t.tr = noopTracer.NewTracerProvider().Tracer("gashash_balash")
		return
	}

	traceExporter, err := otlptracehttp.New(context.Background(), otlptracehttp.WithInsecure(), otlptracehttp.WithEndpoint(t.cfg.Endpoint))
	if err != nil {
		t.l.Error("failed to create trace exporter: %v", zap.Error(err))
		t.cfg.Tracing = false
		return
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(t.ra),
		sdktrace.WithSpanProcessor(bsp),
	)

	otel.SetTracerProvider(tracerProvider) // set global tracer provider

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	t.tp = tracerProvider
	t.tr = tracerProvider.Tracer(t.cfg.ServiceName)
}

func (t *Telemetry) ensureServiceName(name string) string {
	if !strings.HasPrefix(name, t.cfg.ServiceName) {
		name = fmt.Sprintf("%s.%s", t.cfg.ServiceName, name)
	}
	return name
}

func (t *Telemetry) NewUpDownCounter(name string, description string) (metric.Int64UpDownCounter, error) {
	if t == nil {
		return noopMetric.Int64UpDownCounter{}, nil
	}
	meter := otel.GetMeterProvider().Meter(t.cfg.ServiceName)
	name = t.ensureServiceName(name)
	metric, err := meter.Int64UpDownCounter(name, metric.WithDescription(description))
	if err != nil {
		t.l.Error("failed to create metric", zap.String("name", name), zap.Error(err))
		return noopMetric.Int64UpDownCounter{}, err
	}
	return metric, nil
}

func (t *Telemetry) NewCounter(name string, description string) (metric.Int64Counter, error) {
	if t == nil {
		return noopMetric.Int64Counter{}, nil
	}
	meter := otel.GetMeterProvider().Meter(t.cfg.ServiceName)
	name = t.ensureServiceName(name)
	metric, err := meter.Int64Counter(name, metric.WithDescription(description))
	if err != nil {
		t.l.Error("failed to create metric", zap.String("name", name), zap.Error(err))
		return noopMetric.Int64Counter{}, err
	}
	return metric, nil
}

func (t *Telemetry) NewHistogram(name string, description string) (metric.Int64Histogram, error) {
	if t == nil {
		return noopMetric.Int64Histogram{}, nil
	}
	meter := otel.GetMeterProvider().Meter(t.cfg.ServiceName)
	name = t.ensureServiceName(name)
	metric, err := meter.Int64Histogram(name, metric.WithDescription(description))
	if err != nil {
		t.l.Error("failed to create metric", zap.String("name", name), zap.Error(err))
		return noopMetric.Int64Histogram{}, err
	}
	return metric, nil
}

func (t *Telemetry) NewSpan(ctx context.Context, name string) (context.Context, func()) {
	ctx, span := t.tr.Start(ctx, name)
	return ctx, func() {
		span.End()
	}
}
