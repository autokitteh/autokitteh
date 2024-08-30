package websockets

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/discord/internal/vars"
)

func (h *handler) handleMessage(_ *discordgo.Session, m any, eventType string) {
	akEvent, err := transformEvent(h.logger, m, eventType)
	if err != nil {
		return
	}
	cids, err := h.vars.FindConnectionIDs(context.Background(), h.integrationID, vars.BotTokenName, "")
	if err != nil {
		h.logger.Error("Failed to find connection IDs", zap.Error(err))
		return
	}
	h.dispatchAsyncEventsToConnections(cids, akEvent)
}

func (h *handler) HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	h.handleMessage(s, m, "message_create")
}

func (h *handler) HandleMessageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	h.handleMessage(s, m, "message_update")
}

func (h *handler) HandleMessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	h.handleMessage(s, m, "message_delete")
}
