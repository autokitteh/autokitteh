package httpsvc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
)

type ctxKey string

var startTimeCtxKey = ctxKey("t0")

func GetStartTime(ctx context.Context) time.Time {
	if startTime, ok := ctx.Value(startTimeCtxKey).(time.Time); ok {
		return startTime
	}

	// No start time? Duration = 0, not start time = 0.
	return time.Now()
}

type responseInterceptor struct {
	http.ResponseWriter
	StatusCode int
	logger     *zap.Logger
}

func (r *responseInterceptor) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseInterceptor) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
		return
	}

	r.logger.Error(
		"responseInterceptor: underlying http.ResponseWriter does not implement http.Flusher",
		zap.String("type", fmt.Sprintf("%T", r.ResponseWriter)),
	)
}

type RequestLogExtractor func(*http.Request) []zap.Field

var metrics = struct {
	counters  map[string]metric.Int64Counter
	durations map[string]metric.Int64Histogram
}{make(map[string]metric.Int64Counter), make(map[string]metric.Int64Histogram)}

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

	cntName := fmt.Sprintf("api.%s.%s", service, api)
	histName := fmt.Sprintf("api.%s.%s.duration", service, api)

	counter, ok := metrics.counters[cntName]
	if !ok {
		if counter, err = t.NewCounter(cntName, fmt.Sprintf("GRPC request counter (%s)", cntName)); err != nil {
			return err
		}
		metrics.counters[cntName] = counter
	}

	histogram, ok := metrics.durations[histName]
	if !ok {
		if histogram, err = t.NewHistogram(histName, fmt.Sprintf("GRPC request duration (%s)", histName)); err != nil {
			return err
		}
		metrics.durations[histName] = histogram
	}

	counter.Add(ctx, 1, telemetry.WithLabels("status", string(statusCode)))
	histogram.Record(ctx, duration.Milliseconds(), telemetry.WithLabels("status", string(statusCode)))
	return nil
}

func intercept(z *zap.Logger, cfg *LoggerConfig, extractors []RequestLogExtractor, next http.Handler, telemetry *telemetry.Telemetry) (http.HandlerFunc, error) {
	// MustCompile is appropriate here because the patterns are static
	// and errors in them should be caught at startup. Furthermore, we
	// combine them into a single regular expression, because efficiency is
	// critical when we delay each and every incoming HTTP and gRPC request.
	unimportant := regexp.MustCompile(strings.Join(cfg.UnimportantRegexes, `|`))
	unlogged := regexp.MustCompile(strings.Join(cfg.UnloggedRegexes, `|`))

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-AutoKitteh-Process-ID", fixtures.ProcessID())

		l := z.With(zap.String("method", r.Method), zap.String("path", r.URL.Path))
		rwi := &responseInterceptor{ResponseWriter: w, StatusCode: http.StatusOK, logger: l}

		w.Header().Set("Trailer", "X-AutoKitteh-Duration")

		// Call the next handler in the chain, and calculate the duration.
		startTime := time.Now()
		r = r.WithContext(context.WithValue(r.Context(), startTimeCtxKey, startTime))

		next.ServeHTTP(rwi, r)

		duration := time.Since(startTime).Truncate(time.Microsecond)
		updateMetric(r.Context(), telemetry, r.URL.Path, rwi.StatusCode, duration)

		w.Header().Add("X-AutoKitteh-Duration", duration.String())

		// Don't log some requests, unless they result in an error.
		if unlogged.MatchString(r.URL.Path) && rwi.StatusCode < 400 {
			return
		}

		// Otherwise, determine the log level based on the request's
		// URL path and the response's HTTP status code.
		level := cfg.ImportantLevel.Level()
		if unimportant.MatchString(r.URL.Path) {
			level = cfg.UnimportantLevel.Level()
		}
		if rwi.StatusCode >= 400 {
			level = cfg.ErrorsLevel.Level()
		}

		l = l.With(zap.Int("statusCode", rwi.StatusCode), zap.Duration("duration", duration))
		msg := fmt.Sprintf("Response to incoming HTTP request: %s %s", r.Method, r.URL.Path)
		if ce := l.Check(level, msg); ce != nil {
			var fields []zap.Field
			for _, x := range extractors {
				fields = append(fields, x(r)...)
			}

			ce.Write(fields...)
		}
	}, nil
}
