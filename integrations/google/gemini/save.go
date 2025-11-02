package gemini

import (
	"context"
	"net/http"

	"github.com/google/generative-ai-go/genai"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
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

	// Check the "Content-Type" header.
	if common.PostWithoutFormContentType(r) {
		ct := r.Header.Get(common.HeaderContentType)
		l.Warn("save connection: unexpected POST content type", zap.String("content_type", ct))
		c.AbortBadRequest("unexpected content type")
		return
	}

	// Read and parse POST request body.
	if err := r.ParseForm(); err != nil {
		l.Warn("Failed to parse incoming HTTP request", zap.Error(err))
		c.AbortBadRequest("form parsing error")
		return
	}

	// Sanity check: the connection ID is valid.
	cid, err := sdktypes.StrictParseConnectionID(c.ConnectionID)
	if err != nil {
		l.Info("save connection: invalid connection ID", zap.Error(err))
		c.AbortBadRequest("invalid connection ID")
		return
	}

	api_key := r.Form.Get("key")
	if api_key == "" {
		l.Debug("API key is missing from form submission for connection "+cid.String(),
			zap.String("connection_id", cid.String()),
			zap.String("form_field", "key"))
		c.AbortBadRequest("API key is required")
		return
	}

	err = validateGeminiAPIKey(r.Context(), api_key)
	if err != nil {
		l.Debug("Failed to create Google Gemini client for connection "+cid.String()+": "+err.Error(), zap.Error(err))
		c.AbortBadRequest("failed to validate API key, please check your credentials and try again")
		return
	}

	c.Finalize(sdktypes.NewVars().Set(apiKeyVar, api_key, true))
}

// validateGeminiAPIKey makes a test request to validate the provided API key.
func validateGeminiAPIKey(ctx context.Context, apiKey string) error {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return err
	}
	defer client.Close()

	// Get the first model to validate the key works.
	iter := client.ListModels(ctx)
	_, err = iter.Next()
	if err == iterator.Done {
		// No models available, but key is valid.
		return nil
	}
	return err
}
