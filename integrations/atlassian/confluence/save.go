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

// See [webhookEvents] in "webhooks.go" for more details.
var eventCategories = []string{
	"added",
	"archived",
	"copied",
	"created",
	"deleted",
	"moved",
	"removed",
	"trashed",
	"updated",
}

// handleAuth saves a new AutoKitteh connection with user-submitted data.
func (h handler) handleSave(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	// Check "Content-Type" header.
	contentType := r.Header.Get(headerContentType)
	if !strings.HasPrefix(contentType, contentTypeForm) {
		c.Abort("unexpected content type")
		return
	}

	// Read and parse POST request body.
	if err := r.ParseForm(); err != nil {
		l.Warn("Failed to parse incoming HTTP request", zap.Error(err))
		c.Abort("form parsing error")
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

	for _, category := range eventCategories {
		// Register a new webhook to receive, parse, and dispatch
		// Confluence events, if there isn't one already.
		id, ok := getWebhook(l, u, e, t, category)
		if ok {
			// TODO: In the future, when we're sure which events we want to
			// subscribe to, uncomment this line and don't delete the webhook.
			// initData = initData.Set(webhookID, fmt.Sprintf("%d", id), false)

			if err := deleteWebhook(l, u, e, t, id); err != nil {
				l.Warn("Failed to delete existing webhook", zap.Error(err))
				c.Abort("failed to delete existing webhook")
				return
			}
		} // else {
		var secret string
		var err error
		id, secret, err = registerWebhook(l, u, e, t, category)
		if err != nil {
			l.Warn("Failed to register webhook", zap.Error(err))
			c.Abort("failed to register webhook")
			return
		}
		initData = initData.Set(webhookSecret(category), secret, true)
		// }

		initData = initData.Set(webhookID(category), fmt.Sprintf("%d", id), false)
	}

	c.Finalize(initData)
}
