package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/discord/internal/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (h *handler) handleEvent(event any, eventType string) {
	var authorID string

	// TODO(ENG-1546) - Add support for more event types
	switch e := event.(type) {
	case *discordgo.MessageCreate:
		authorID = e.Author.ID
	case *discordgo.MessageUpdate:
		authorID = e.Author.ID
	case *discordgo.MessageDelete:
		authorID = e.Author.ID
	default:
		h.logger.Error("Unsupported event type", zap.String("eventType", eventType))
		return
	}

	cids, err := h.vars.FindConnectionIDs(context.Background(), h.integrationID, vars.BotToken, "")
	if err != nil {
		h.logger.Error("Failed to find connection IDs", zap.Error(err))
		return
	}
	akEvent, err := h.transformEvent(event, eventType)
	if err != nil {
		return
	}

	// Filter out connections from the application itself
	var validCIDs []sdktypes.ConnectionID
	for _, cid := range cids {
		vs, err := h.vars.Get(context.Background(), sdktypes.NewVarScopeID(cid))
		if err != nil {
			h.logger.Error("Failed to get connection vars", zap.Error(err))
			return
		}
		if vs.Get(vars.BotID).Value() == authorID {
			h.logger.Debug("Skipping event for connection", zap.String("connectionID", cid.String()))
			continue
		}
		validCIDs = append(validCIDs, cid)
	}

	h.dispatchAsyncEventsToConnections(validCIDs, akEvent)
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
