package httpsvc

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
	"go.opentelemetry.io/otel/metric"
)

type apiMetrics struct {
	counter  metric.Int64Counter
	duration metric.Int64Histogram
}

var metrics sync.Map

// metrics    map[string]apiMetrics
// metricsMux sync.Mutex

func getServiceAPIMetrics(serviceAPI string, t *telemetry.Telemetry) (*apiMetrics, error) {
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
	metrics.LoadOrStore(serviceAPI, &m)
	return &m, nil
}

func updateMetric(ctx context.Context, t *telemetry.Telemetry, path string, statusCode int, duration time.Duration) (err error) {
	// path will be like "/autokitteh.projects.v1.ProjectsService/Create"
	// 1. check this is an internal API path, e.g. starts with "/autokitteh."
	// 2. extract service (`projects`) and API name (`create`)

	if !strings.HasPrefix(path, "/autokitteh.") {
		return nil // only internal service APIs
	}

	slashParts := strings.Split(path, "/")
	if len(slashParts) < 3 {
		return errors.New("invalid API path")
	}
	api := strings.ToLower(slashParts[2])

	var service string
	dotParts := strings.Split(slashParts[1], ".")
	if len(dotParts) < 2 {
		return errors.New("invalid service path")
	}
	service = dotParts[1]

	serviceAPI := fmt.Sprintf("%s.%s", service, api) // e.g. "projects.create"
	m, err := getServiceAPIMetrics(serviceAPI, t)
	if err != nil {
		return err
	}

	m.counter.Add(ctx, 1, telemetry.WithLabels("status", strconv.Itoa(statusCode)))
	m.duration.Record(ctx, duration.Milliseconds(), telemetry.WithLabels("status", strconv.Itoa(statusCode)))
	return nil
}
