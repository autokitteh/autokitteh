package chatgpt

import (
	"net/http"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	headerContentType = "Content-Type"
	contentTypeForm   = "application/x-www-form-urlencoded"
)

var apiKeyVar = sdktypes.NewSymbol("api_key")

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
	// Test the OpenAI authentication details.
	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	api := OpenAIAPI{
		APIKey: r.Form.Get("key"),
	}
	err := api.Test(ctx)
	if err != nil {
		l.Warn("OpenAI authentication test failed", zap.Error(err))
		c.AbortBadRequest("OpenAI authentication test failed: " + err.Error())
		return
	}

	c.Finalize(sdktypes.NewVars().Set(apiKeyVar, api.APIKey, true))
}
