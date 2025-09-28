package telegram

import (
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
)

// HandleWebhook processes incoming Telegram webhook events
func (h handler) handleEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Check the request's headers and parse its body.
	telegramEvent := h.checkRequest(w, r)
	if telegramEvent == nil {
		return
	}
}

func (h handler) checkRequest(w http.ResponseWriter, r *http.Request) map[string]any {
	l := h.logger.With(
		zap.String("url_path", r.URL.Path),
		zap.String("event_type", r.Header.Get("Telegram-Event")),
	)

	// No need to check the HTTP method, as we only accept POST requests.

	// Check the request's HTTP headers.
	if common.PostWithoutJSONContentType(r) {
		ct := r.Header.Get(common.HeaderContentType)
		l.Warn("incoming event: unexpected content type", zap.String("content_type", ct))
		common.HTTPError(w, http.StatusBadRequest)
		return nil
	}

	return nil
}
