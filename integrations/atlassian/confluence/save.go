package confluence

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
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

	b := r.Form.Get("base_url")
	t := r.Form.Get("token")
	e := r.Form.Get("email")

	u, err := url.Parse(b)
	if err != nil {
		l.Warn("Failed to parse base URL", zap.Error(err))
		c.AbortBadRequest("base URL parsing error")
		return
	}

	// Ensure the base URL is formatted as we expect.
	b = fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	initData := sdktypes.NewVars().Set(baseURL, b, false).Set(token, t, true)
	if e != "" {
		initData = initData.Set(email, e, true).Set(authType, "apiToken", false)
	} else {
		initData = initData.Set(authType, "pat", false)
	}

	ctx := r.Context()
	for _, category := range eventCategories {
		// Register a new webhook to receive, parse, and dispatch
		// Confluence events, if there isn't one already.
		id, ok := getWebhook(ctx, l, b, e, t, category)
		if ok {
			// TODO: In the future, when we're sure which events we want to
			// subscribe to, uncomment this line and don't delete the webhook.
			// initData = initData.Set(webhookID, fmt.Sprintf("%d", id), false)

			if err := deleteWebhook(ctx, l, b, e, t, id); err != nil {
				l.Error("Failed to delete existing webhook", zap.Error(err))
				c.AbortServerError("failed to delete existing webhook")
				return
			}
		} // else {
		var secret string
		var err error
		id, secret, err = registerWebhook(ctx, l, b, e, t, category)
		if err != nil {
			l.Error("Failed to register webhook", zap.Error(err))
			c.AbortServerError("failed to register webhook")
			return
		}
		initData = initData.Set(webhookSecret(category), secret, true)
		// }

		initData = initData.Set(webhookID(category), strconv.Itoa(id), false)
	}

	c.Finalize(initData)
}
