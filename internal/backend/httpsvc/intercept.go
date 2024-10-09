package httpsvc

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

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

func intercept(z *zap.Logger, cfg *LoggerConfig, extractors []RequestLogExtractor, next http.Handler, telemetry *telemetry.Telemetry) (http.HandlerFunc, error) {
	// MustCompile is appropriate here because the patterns are static
	// and errors in them should be caught at startup. Furthermore, we
	// combine them into a single regular expression, because efficiency is
	// critical when we delay each and every incoming HTTP and gRPC request.
	unimportant := regexp.MustCompile(strings.Join(cfg.UnimportantRegexes, `|`))
	unlogged := regexp.MustCompile(strings.Join(cfg.UnloggedRegexes, `|`))

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("AutoKitteh-Process-ID", fixtures.ProcessID())
		// https://pkg.go.dev/net/http#example-ResponseWriter-Trailers
		w.Header().Set("Trailer", "AutoKitteh-Duration")

		l := z.With(zap.String("method", r.Method), zap.String("path", r.URL.Path))
		rwi := &responseInterceptor{ResponseWriter: w, StatusCode: http.StatusOK, logger: l}

		// Call the next handler in the chain, and calculate the duration.
		startTime := time.Now()
		r = r.WithContext(context.WithValue(r.Context(), startTimeCtxKey, startTime))

		next.ServeHTTP(rwi, r)

		duration := time.Since(startTime)
		updateMetric(r.Context(), telemetry, r.URL.Path, rwi.StatusCode, duration)
		w.Header().Set("AutoKitteh-Duration", duration.String())

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
		msg := fmt.Sprintf("HTTP Response: %s %s", r.Method, r.URL.Path)
		if ce := l.Check(level, msg); ce != nil {
			var fields []zap.Field
			for _, x := range extractors {
				fields = append(fields, x(r)...)
			}

			ce.Write(fields...)
		}
	}, nil
}
