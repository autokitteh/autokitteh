package salesforce

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
)

func (h handler) dispatchEvent(payload map[string]any, eventType string) {
	akEvent, err := common.TransformEvent(h.logger, payload, eventType)
	if err != nil {
		h.logger.Error("failed to transform Salesforce event", zap.Error(err))
		return
	}

	ctx := context.Background()
	cids, err := h.vars.FindActiveConnectionIDs(ctx, desc.ID(), instanceURLVar, "")
	if err != nil {
		h.logger.Error("failed to find connection IDs", zap.Error(err))
		return
	}

	common.DispatchEvent(ctx, h.logger, h.dispatch, akEvent, cids)
}
