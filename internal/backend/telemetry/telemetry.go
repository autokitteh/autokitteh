package telemetry

import (
	"context"
	"fmt"
	"net/http"

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
)

var (
	tracerProvider trace.TracerProvider = noopTracer.NewTracerProvider()
	tracer         trace.Tracer         = tracerProvider.Tracer("ak")

	meterProvider metric.MeterProvider = noopMetric.NewMeterProvider()
	meter         metric.Meter         = meterProvider.Meter("ak")
)

func init() {
	otel.SetMeterProvider(meterProvider)
	otel.SetTracerProvider(tracerProvider)
}

func TraceProvider() trace.TracerProvider  { return tracerProvider }
func MetricProvider() metric.MeterProvider { return meterProvider }

func T() trace.Tracer { return tracer }
func M() metric.Meter { return meter }

func HTTPInterceptor(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := baggage.ContextWithoutBaggage(r.Context())
		ctx, span := tracer.Start(ctx, r.RequestURI)
		defer span.End()
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func Init(l *zap.Logger, cfg *Config) error {
	rsc := resource.NewWithAttributes(
		"https://opentelemetry.io/schemas/1.1.0",
		semconv.ServiceNameKey.String(cfg.ServiceName),
	)

	if err := setupMetrics(cfg, rsc); err != nil {
		return fmt.Errorf("metrics: %w", err)
	}

	if err := setupTracing(cfg, rsc); err != nil {
		return fmt.Errorf("tracing: %w", err)
	}

	return nil
}

func setupMetrics(cfg *Config, ra *resource.Resource) error {
	if !cfg.Enabled {
		return nil
	}

	// TODO(ENG-1445): gRPC?
	exporter, err := otlpmetrichttp.New(
		context.Background(),
		otlpmetrichttp.WithInsecure(),
		otlpmetrichttp.WithEndpoint(cfg.Endpoint),
		// metrics will be sent to ENDPOINT:/v1/Metrcis. Use WithURLPath to override
	)
	if err != nil {
		return fmt.Errorf("failed to create metric exporter: %w", err)
	}

	// NOTE: do we need a better control ober batching/sending. Should we use controller?
	meterProvider = sdk.NewMeterProvider(
		sdk.WithReader(sdk.NewPeriodicReader(exporter)),
		sdk.WithResource(ra),
	)

	otel.SetMeterProvider(meterProvider)

	meter = &namedMeter{underlying: meterProvider.Meter(cfg.ServiceName), svcName: cfg.ServiceName}

	return nil
}

func setupTracing(cfg *Config, ra *resource.Resource) error {
	if !cfg.Tracing {
		return nil
	}

	traceExporter, err := otlptracehttp.New(context.Background(), otlptracehttp.WithInsecure(), otlptracehttp.WithEndpoint(cfg.Endpoint))
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)

	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.TracingFraction)),
		sdktrace.WithResource(ra),
		sdktrace.WithSpanProcessor(bsp),
	)

	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	tracer = tracerProvider.Tracer(cfg.ServiceName)

	return nil
}
