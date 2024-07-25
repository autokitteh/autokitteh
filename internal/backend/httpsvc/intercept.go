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

func intercept(z *zap.Logger, cfg *LoggerConfig, extractors []RequestLogExtractor, next http.Handler) (http.HandlerFunc, error) {
	// MustCompile is appropriate here because the patterns are static
	// and errors in them should be caught at startup. Furthermore, we
	// combine them into a single regular expression, because efficiency is
	// critical when we delay each and every incoming HTTP and gRPC request.
	unimportant := regexp.MustCompile(strings.Join(cfg.UnimportantRegexes, `|`))
	unlogged := regexp.MustCompile(strings.Join(cfg.UnloggedRegexes, `|`))

	return func(w http.ResponseWriter, r *http.Request) {
		if unlogged.MatchString(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("X-AutoKitteh-Process-ID", fixtures.ProcessID())

		l := z.With(zap.String("method", r.Method), zap.String("path", r.URL.Path))
		rwi := &responseInterceptor{ResponseWriter: w, StatusCode: http.StatusOK, logger: l}

		w.Header().Set("Trailer", "X-AutoKitteh-Duration")

		startTime := time.Now()
		r = r.WithContext(context.WithValue(r.Context(), startTimeCtxKey, startTime))

		next.ServeHTTP(rwi, r)
		d := time.Since(startTime)

		w.Header().Add("X-AutoKitteh-Duration", d.Truncate(time.Microsecond).String())

		level := cfg.ImportantLevel.Level()
		if unimportant.MatchString(r.URL.Path) {
			level = cfg.UnimportantLevel.Level()
		}
		if rwi.StatusCode >= 400 {
			level = cfg.ErrorsLevel.Level()
		}

		l = l.With(zap.Int("statusCode", rwi.StatusCode), zap.Duration("duration", d))
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
