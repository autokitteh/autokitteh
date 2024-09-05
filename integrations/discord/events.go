package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/discord/internal/vars"
)

func (h *handler) handleEvent(event any, eventType string) {
	akEvent, err := h.transformEvent(event, eventType)
	if err != nil {
		return
	}
	cids, err := h.vars.FindConnectionIDs(context.Background(), h.integrationID, vars.BotToken, "")
	if err != nil {
		h.logger.Error("Failed to find connection IDs", zap.Error(err))
		return
	}
	h.dispatchAsyncEventsToConnections(cids, akEvent)
}

func (h *handler) handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	h.handleEvent(m, "message_create")
}

func (h *handler) handleMessageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	h.handleEvent(m, "message_update")
}

func (h *handler) handleMessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	h.handleEvent(m, "message_delete")
}
