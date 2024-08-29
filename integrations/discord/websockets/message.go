package websockets

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/discord/internal/vars"
)

const (
	MessagePath = "/discord/message"
)

func (h *handler) HandleDiscordMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Transform the received Slack event into an AutoKitteh event.
	akEvent, err := transformEvent(h.logger, m, "message")
	if err != nil {
		return
	}
	// Retrieve all the relevant connections for this event.
	cids, err := h.vars.FindConnectionIDs(context.Background(), h.integrationID, vars.BotTokenName, "")
	if err != nil {
		h.logger.Error("Failed to find connection IDs", zap.Error(err))
		return
	}
	// Dispatch the event to all of them, for asynchronous handling.
	h.dispatchAsyncEventsToConnections(cids, akEvent)
}
