package websockets

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
	integration   sdktypes.Integration
	integrationID sdktypes.IntegrationID
}

func NewHandler(l *zap.Logger, v sdkservices.Vars, d sdkservices.Dispatcher, i sdktypes.Integration) handler {
	return handler{
		logger:        l,
		vars:          v,
		dispatcher:    d,
		integration:   i,
		integrationID: i.ID(),
	}
}

var (
	webSocketClients = make(map[string]*discordgo.Session)

	mu sync.RWMutex
)

func (h handler) OpenSocketModeConnection(botID, botToken string) {
	mu.Lock()
	defer mu.Unlock()

	// No need to open multiple connections for the same app - yet.
	if _, ok := webSocketClients[botID]; ok {
		return
	}

	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		h.logger.Error("Error creating Discord session", zap.Error(err))
		return
	}

	addHandlers(dg, h)

	// Open a WebSocket connection to Discord.
	err = dg.Open()

	if err == nil {
		webSocketClients[botID] = dg
		return // Normal process termination.
	}
	h.logger.Error("Failed to open Discord WebSocket connection", zap.Error(err))
}

func (h handler) dispatchAsyncEventsToConnections(cids []sdktypes.ConnectionID, e sdktypes.Event) {
	ctx := extrazap.AttachLoggerToContext(h.logger, context.Background())
	for _, cid := range cids {
		eid, err := h.dispatcher.Dispatch(ctx, e.WithConnectionID(cid), nil)
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

// Transform the received Discord event into an AutoKitteh event.
func transformEvent(l *zap.Logger, discordEvent any, eventType string) (sdktypes.Event, error) {
	l = l.With(
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

func addHandlers(dg *discordgo.Session, h handler) {
	// messages
	dg.AddHandler(h.HandleMessageCreate)
	dg.AddHandler(h.HandleMessageDelete)
	dg.AddHandler(h.HandleMessageUpdate)
}
