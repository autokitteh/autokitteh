package discord

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/discord/internal/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (h *handler) handleEvent(event any, eventType string) {
	l := h.logger.With(zap.String("event_type", eventType))

	var initiatorID, dedupID string
	switch e := event.(type) {
	case *discordgo.MessageCreate:
		initiatorID = e.Author.ID
		dedupID = fmt.Sprintf("%s/create", e.ID)
	case *discordgo.MessageUpdate:
		initiatorID = e.Author.ID
		var ts string
		if e.EditedTimestamp != nil {
			ts = e.EditedTimestamp.String()
		}
		dedupID = fmt.Sprintf("%s/%s/update", e.ID, ts)
	case *discordgo.MessageDelete:
		initiatorID = "" // Deleted messages don't have an author
		dedupID = fmt.Sprintf("%s/delete", e.ID)
	case *discordgo.PresenceUpdate:
		initiatorID = e.User.ID
		dedupID = fmt.Sprintf("%s/%v/presence_update", e.User.ID, e.Since)
	case *discordgo.ThreadCreate:
		initiatorID = e.OwnerID
		dedupID = fmt.Sprintf("%s/create", e.ID)
	case *discordgo.ThreadDelete:
		initiatorID = e.OwnerID
		dedupID = fmt.Sprintf("%s/delete", e.ID)

	// NON-UNIQUE EVENTS for deduplication purposes.
	case *discordgo.ThreadUpdate:
		// No way to uniquely identify the event, no deduping.
		initiatorID = e.OwnerID
	case *discordgo.MessageReactionAdd:
		initiatorID = e.UserID
		dedupID = fmt.Sprintf("%s/%s/%s/add", e.UserID, e.MessageID, e.Emoji.ID)
	case *discordgo.MessageReactionRemove:
		initiatorID = e.UserID
		// No way to uniquely identify the event - identify on the first.
		dedupID = fmt.Sprintf("%s/%s/%s/remove", e.UserID, e.MessageID, e.Emoji.ID)

	default:
		l.Error("Unsupported event type")
		return
	}

	akEvent, err := h.transformEvent(dedupID, event, eventType)
	if err != nil {
		return
	}

	cids, err := h.vars.FindActiveConnectionIDs(context.Background(), h.integrationID, vars.BotToken, "")
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
