package events

import (
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
)

func invalidEventError(l *zap.Logger, w http.ResponseWriter, body []byte, err error) {
	l.Error("Failed to parse JSON payload",
		zap.ByteString("json", body),
		zap.Error(err),
	)

	if w == nil {
		// If the event parsing function was called by a WebSocket,
		// there's no way/need to write an HTTP error response.
		return
	}
	common.HTTPError(w, http.StatusBadRequest)
}
