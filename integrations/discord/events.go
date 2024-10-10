package discord

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/discord/internal/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (h *handler) handleEvent(event any, eventType string) {
	l := h.logger.With(zap.String("eventType", eventType))

	var initiatorID string
	switch e := event.(type) {
	case *discordgo.MessageCreate:
		initiatorID = e.Author.ID
	case *discordgo.MessageUpdate:
		initiatorID = e.Author.ID
	case *discordgo.MessageDelete:
		initiatorID = "" // Deleted messages don't have an author
	case *discordgo.MessageReactionAdd:
		initiatorID = e.UserID
	case *discordgo.MessageReactionRemove:
		initiatorID = e.UserID
	case *discordgo.PresenceUpdate:
		initiatorID = e.User.ID
	case *discordgo.ThreadCreate:
		initiatorID = e.OwnerID
	case *discordgo.ThreadUpdate:
		initiatorID = e.OwnerID
	case *discordgo.ThreadDelete:
		initiatorID = e.OwnerID
	default:
		l.Error("Unsupported event type")
		return
	}

	akEvent, err := h.transformEvent(event, eventType)
	if err != nil {
		return
	}

	cids, err := h.vars.FindConnectionIDs(context.Background(), h.integrationID, vars.BotToken, "")
	if err != nil {
		l.Error("Failed to find connection IDs", zap.Error(err))
		return
	}

	// Don't send the event to connections that use the same bot that initiated it.
	var validCIDs []sdktypes.ConnectionID
	for _, cid := range cids {
		vs, err := h.vars.Get(context.Background(), sdktypes.NewVarScopeID(cid))
		if err != nil {
			l.Error("Failed to get connection vars", zap.String("connectionID", cid.String()), zap.Error(err))
			continue
		}
		if vs.Get(vars.BotID).Value() == initiatorID {
			l.Debug("Skipping event for connection", zap.String("connectionID", cid.String()))
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

func (h *handler) handleMessageReactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	h.handleEvent(m, "message_reaction_add")
}

func (h *handler) handleMessageReactionRemove(s *discordgo.Session, m *discordgo.MessageReactionRemove) {
	h.handleEvent(m, "message_reaction_remove")
}

func (h *handler) handlePresenceUpdate(s *discordgo.Session, p *discordgo.PresenceUpdate) {
	h.handleEvent(p, "presence_update")
}

func (h *handler) handleThreadCreate(s *discordgo.Session, t *discordgo.ThreadCreate) {
	h.handleEvent(t, "thread_create")
}

func (h *handler) handleThreadUpdate(s *discordgo.Session, t *discordgo.ThreadUpdate) {
	h.handleEvent(t, "thread_update")
}

func (h *handler) handleThreadDelete(s *discordgo.Session, t *discordgo.ThreadDelete) {
	h.handleEvent(t, "thread_delete")
}
