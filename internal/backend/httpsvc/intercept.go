package httpsvc

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

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
	res, err := kittehs.TransformError(cfg.NonimportantRegexes, regexp.Compile)
	if err != nil {
		return nil, fmt.Errorf("compiling important regexes: %w", err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-AutoKitteh-ID", fixtures.ProcessID())

		z := z.With(zap.String("method", r.Method), zap.String("path", r.URL.Path))

		level := cfg.ImportantLevel.Level()

		for _, re := range res {
			if re.MatchString(r.URL.Path) {
				level = cfg.NonimportantLevel.Level()
				break
			}
		}

		if ce := z.Check(level, "HTTP Request"); ce != nil {
			var fields []zap.Field
			for _, x := range extractors {
				fields = append(fields, x(r)...)
			}

			ce.Write(fields...)
		}

		rwi := &responseInterceptor{ResponseWriter: w, StatusCode: http.StatusOK, logger: z}

		t0 := time.Now()
		next.ServeHTTP(rwi, r)
		d := time.Since(t0)

		if rwi.StatusCode >= 400 {
			level = cfg.ErrorsLevel.Level()
		}

		if ce := z.Check(level, "HTTP Response"); ce != nil {
			ce.Write(zap.Int("status_code", rwi.StatusCode), zap.Duration("duration", d))
		}
	}, nil
}
