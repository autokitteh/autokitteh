package discord

import (
	"context"
	"sync"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type handler struct {
	logger        *zap.Logger
	vars          sdkservices.Vars
	dispatcher    sdkservices.Dispatcher
	integrationID sdktypes.IntegrationID
}

func NewHandler(l *zap.Logger, v sdkservices.Vars, d sdkservices.Dispatcher, i sdktypes.Integration) handler {
	return handler{
		logger:        l,
		vars:          v,
		dispatcher:    d,
		integrationID: i.ID(),
	}
}

var (
	// Key = botToken (ensures one WebSocket per bot).
	discordSessions = make(map[string]*discordgo.Session)

	mu = &sync.Mutex{}
)

func (h handler) OpenWebSocketConnection(botToken string) {
	// Ensure multiple users don't reference the same bot at the same time.
	mu.Lock()
	defer mu.Unlock()

	// Check if a session already exists for this bot token.
	if _, ok := discordSessions[botToken]; ok {
		return
	}

	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		h.logger.Error("Error creating Discord session", zap.Error(err))
		return
	}

	h.addHandlers(dg)

	// Open a WebSocket connection to Discord.
	if err := dg.Open(); err != nil {
		h.logger.Error("Failed to open Discord WebSocket connection", zap.Error(err))
		return
	}

	discordSessions[botToken] = dg
}

func (h handler) dispatchAsyncEventsToConnections(cids []sdktypes.ConnectionID, e sdktypes.Event) {
	ctx := extrazap.AttachLoggerToContext(h.logger, context.Background())
	for _, cid := range cids {
		eid, err := h.dispatcher.Dispatch(ctx, e.WithConnectionDestinationID(cid), nil)
		l := h.logger.With(
			zap.String("connectionID", cid.String()),
			zap.String("eventID", eid.String()),
		)
		if err != nil {
			l.Error("Event dispatch failed", zap.Error(err))
			return
		}
		l.Debug("Event dispatched")
	}
}

// transformEvent transforms the received Discord event into an AutoKitteh event.
func (h handler) transformEvent(discordEvent any, eventType string) (sdktypes.Event, error) {
	l := h.logger.With(
		zap.String("eventType", eventType),
		zap.Any("event", discordEvent),
	)

	wrapped, err := sdktypes.WrapValue(discordEvent)
	if err != nil {
		l.Error("Failed to wrap Discord event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	data, err := wrapped.ToStringValuesMap()
	if err != nil {
		l.Error("Failed to convert wrapped Discord event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	akEvent, err := sdktypes.EventFromProto(&sdktypes.EventPB{
		EventType: eventType,
		Data:      kittehs.TransformMapValues(data, sdktypes.ToProto),
	})
	if err != nil {
		l.Error("Failed to convert protocol buffer to SDK event",
			zap.Any("data", data),
			zap.Error(err),
		)
		return sdktypes.InvalidEvent, err
	}

	return akEvent, nil
}

func (h handler) addHandlers(dg *discordgo.Session) {
	dg.AddHandler(h.handleMessageCreate)
	dg.AddHandler(h.handleMessageDelete)
	dg.AddHandler(h.handleMessageUpdate)
}
