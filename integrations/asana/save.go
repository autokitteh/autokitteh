package asana

import (
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	headerContentType = "Content-Type"
	contentTypeForm   = "application/x-www-form-urlencoded"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to save data from web form submissions as connections.
type handler struct {
	logger *zap.Logger
}

func NewHTTPHandler(l *zap.Logger) http.Handler {
	return handler{logger: l}
}

// ServeHTTP saves a new autokitteh connection with user-submitted data.
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	// Check "Content-Type" header.
	contentType := r.Header.Get(headerContentType)
	if !strings.HasPrefix(contentType, contentTypeForm) {
		c.AbortBadRequest("unexpected content type")
		return
	}

	// Read and parse POST request body.
	if err := r.ParseForm(); err != nil {
		l.Warn("Failed to parse incoming HTTP request", zap.Error(err))
		c.AbortBadRequest("form parsing error")
		return
	}

	// Test the PAT's usability
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, "https://app.asana.com/api/1.0/users/me", nil)
	if err != nil {
		l.Error("Failed to create HTTP request", zap.Error(err))
		c.AbortServerError("request creation error")
		return
	}

	ps := r.Form.Get("pat")
	if ps == "" {
		l.Warn("PAT not provided in request form")
		c.AbortBadRequest("PAT is missing")
		return
	}

	req.Header.Add("Authorization", "Bearer "+ps)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		l.Error("Failed to execute HTTP request", zap.Error(err))
		c.AbortBadRequest("execution error")
		return
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		l.Error("Failed to read response body", zap.Error(err))
		c.AbortServerError("response reading error")
		return
	}

	if resp.StatusCode != http.StatusOK {
		l.Warn("Token is invalid or an error occurred", zap.Int("status_code", resp.StatusCode))
		c.AbortBadRequest("invalid token or error occurred")
		return
	}

	// Finalize the connection with the valid PAT
	c.Finalize(sdktypes.NewVars().
		Set(patVar, ps, true).
		Set(authTypeVar, integrations.PAT, false))
}
