package httpsvc

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
)

type apiMetrics struct {
	counter  metric.Int64Counter
	duration metric.Int64Histogram
}

var metrics sync.Map

// obtains service API metrics, creating them if not already present
func acquireServiceAPIMetrics(serviceAPI string, t *telemetry.Telemetry) (*apiMetrics, error) {
	if m, ok := metrics.Load(serviceAPI); ok {
		return m.(*apiMetrics), nil
	}

	cntName := fmt.Sprintf("api.%s", serviceAPI)
	histName := fmt.Sprintf("api.%s.duration", serviceAPI)

	counter, err := t.NewCounter(cntName, fmt.Sprintf("GRPC request counter (%s)", cntName))
	if err != nil {
		return nil, err
	}
	histogram, err := t.NewHistogram(histName, fmt.Sprintf("GRPC request duration (%s)", histName))
	if err != nil {
		return nil, err
	}
	m := apiMetrics{counter, histogram}
	metrics.LoadOrStore(serviceAPI, &m) // when called concurrently, this will not overwrite if already set
	return &m, nil
}

// See TestGetMetricNameFromPath for examples.
func getMetricNameFromPath(path string) string {
	// path will be like "/autokitteh.projects.v1.ProjectsService/Create"
	// 1. check this is an internal API path, e.g. starts with "/autokitteh."
	// 2. extract service (`projects`) and API name (`create`)

	path, found := strings.CutPrefix(path, "/autokitteh.")
	if !found {
		// only internal service APIs
		return ""
	}

	svc, method, found := strings.Cut(path, "/")
	if !found {
		return ""
	}

	// 0 - package, 1 - version, 2 - service name
	dotParts := strings.SplitN(svc, ".", 3)
	if len(dotParts) < 3 {
		return ""
	}

	service := strings.ToLower(strings.TrimSuffix(dotParts[2], "Service"))

	pkg, ver := dotParts[0], dotParts[1]

	if pkg != service {
		service = pkg + "_" + service
	}

	return fmt.Sprintf("%s.%s.%s", service, ver, strings.ToLower(method)) // e.g. "projects.v1.create"
}

func updateMetric(ctx context.Context, t *telemetry.Telemetry, path string, statusCode int, duration time.Duration) {
	name := getMetricNameFromPath(path)
	if name == "" {
		return
	}

	m, err := acquireServiceAPIMetrics(name, t)
	if err != nil {
		return
	}

	attrs := metric.WithAttributes(attribute.Int("status", statusCode))
	m.counter.Add(ctx, 1, attrs)
	m.duration.Record(ctx, duration.Milliseconds(), attrs)
}
