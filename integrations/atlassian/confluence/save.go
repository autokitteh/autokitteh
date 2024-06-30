package confluence

import (
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	contentTypeForm = "application/x-www-form-urlencoded"
)

// handleAuth saves a new AutoKitteh connection with user-submitted data.
func (h handler) handleSave(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", r.URL.Path))

	// Check the "Content-Type" header.
	contentType := r.Header.Get(headerContentType)
	if !strings.HasPrefix(contentType, contentTypeForm) {
		// Probably an attack, so no need for user-friendliness.
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Read and parse POST request body.
	if err := r.ParseForm(); err != nil {
		l.Warn("Failed to parse inbound HTTP request", zap.Error(err))
		redirectToErrorPage(w, r, "form parsing error: "+err.Error())
		return
	}

	u := r.Form.Get("base_url")
	t := r.Form.Get("token")
	e := r.Form.Get("email")

	initData := sdktypes.NewVars().
		Set(baseURL, u, false).
		Set(token, t, true)

	if e != "" {
		initData = initData.Set(email, e, true)
	}

	// Register a new webhook to receive, parse, and dispatch
	// Confluence events, if there isn't one already.
	id, ok := getWebhook(l, u, e, t)
	if ok {
		// TODO: In the future, when we're sure which events we want to
		// subscribe to, uncomment this line and don't delete the webhook.
		// initData = initData.Set(webhookID, fmt.Sprintf("%d", id), false)

		if err := deleteWebhook(l, u, e, t, id); err != nil {
			redirectToErrorPage(w, r, "failed to delete existing webhook: "+err.Error())
			return
		}
	} // else {
	var secret string
	var err error
	id, secret, err = registerWebhook(l, u, e, t)
	if err != nil {
		redirectToErrorPage(w, r, "failed to register webhook: "+err.Error())
		return
	}
	initData = initData.Set(webhookSecret, secret, true)
	// }

	initData = initData.Set(webhookID, fmt.Sprintf("%d", id), false)

	sdkintegrations.FinalizeConnectionInit(w, r, integrationID, initData)
}
