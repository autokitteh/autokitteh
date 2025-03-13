package jira

import (
	"fmt"
	"net/http"
	"net/url"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

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

	c.Finalize(initData)
}
