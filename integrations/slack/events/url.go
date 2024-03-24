package events

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/slack/api"
)

type urlVerificationContainer struct {
	Challenge string `json:"challenge"`
}

// https://api.slack.com/events/url_verification
func URLVerificationHandler(l *zap.Logger, w http.ResponseWriter, body []byte, _ *Callback) any {
	// Parse the inner event details.
	j := &urlVerificationContainer{}
	if err := json.Unmarshal(body, j); err != nil {
		invalidEventError(l, w, body, err)
		return nil
	}

	// Respond to Slack immediately.
	l.Debug("Echoing the challenge string")
	if w != nil {
		// WebSockets never use this event by definition, so we
		// don't have to check that w != nil, but it's still smart.
		w.Header().Add(api.HeaderContentType, api.ContentTypeJSONCharsetUTF8)
		fmt.Fprintf(w, `{"challenge":"%s"}`, j.Challenge)
	}

	// No need to dispatch this event to autokitteh.
	return nil
}
