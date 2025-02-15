package websockets

import (
	"context"
	"sync"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handler implements WebSockets to receive and dispatch
// third-party asynchronous event notifications.
type Handler struct {
	logger        *zap.Logger
	vars          sdkservices.Vars
	dispatch      sdkservices.DispatchFunc
	integrationID sdktypes.IntegrationID
}

func NewHandler(l *zap.Logger, v sdkservices.Vars, d sdkservices.DispatchFunc, i sdktypes.IntegrationID) Handler {
	return Handler{logger: l, vars: v, dispatch: d, integrationID: i}
}

var (
	// Key = Slack app ID (to ensure one WebSocket per app).
	// Writes happen during server startup, and when a user
	// creates a new Socket Mode connection in the UI.
	webSocketClients = make(map[string]*socketmode.Client)

	mu = &sync.Mutex{}
)

func (h Handler) OpenWebSocketConnection(appID, appToken, botToken string) {
	// Ensure multiple users don't reference the same app at the same time.
	mu.Lock()
	defer mu.Unlock()

	// No need to open multiple connections for the same app - yet.
	// See: https://docs.autokitteh.com/integrations/slack/connection
	if _, ok := webSocketClients[appID]; ok {
		return
	}

	client := slack.New(botToken, slack.OptionAppLevelToken(appToken))
	webSocketClients[appID] = socketmode.New(client)

	go func() {
		for {
			smh := socketmode.NewSocketmodeHandler(webSocketClients[appID])
			smh.HandleDefault(h.socketModeHandler)

			err := smh.RunEventLoop()
			if err == nil {
				return // Normal process termination.
			}
			h.logger.Error("Slack Socket Mode (re)connection error", zap.Error(err))
		}
	}()
}

func (h Handler) socketModeHandler(e *socketmode.Event, c *socketmode.Client) {
	msg := "Slack Socket Mode event: " + string(e.Type)
	switch string(e.Type) {
	// WebSocket connection flow.
	case "connecting", "connected":
		h.logger.Debug(msg)
	case "hello":
		h.logger.Debug(msg)

	// TODO(ENG-549): slack-go handles "disconnect" events, but not robustly
	// at all (https://api.slack.com/apis/connections/socket#disconnect).

	// Events.
	case "events_api":
		h.logger.Debug(msg)
		h.handleBotEvent(e, c)
	case "interactive":
		h.logger.Debug(msg)
		h.handleInteractiveEvent(e, c)
	case "slash_commands":
		h.logger.Debug(msg)
		h.handleSlashCommand(e, c)

	// Errors.
	case "connection_error", "incoming_error":
		h.logger.Error(msg, zap.Any("data", e.Data))

	default:
		h.logger.Warn("Unhandled Slack Socket Mode event",
			zap.String("type", string(e.Type)),
			zap.Any("data", e.Data),
			zap.Any("request", e.Request),
		)
	}
}

// Transform the received Slack event into an AutoKitteh event.
func transformEvent(l *zap.Logger, slackEvent any, eventType string) (sdktypes.Event, error) {
	l = l.With(
		zap.String("eventType", eventType),
		zap.Any("event", slackEvent),
	)

	wrapped, err := sdktypes.WrapValue(slackEvent)
	if err != nil {
		l.Error("Failed to wrap Slack event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	data, err := wrapped.ToStringValuesMap()
	if err != nil {
		l.Error("Failed to convert wrapped Slack event", zap.Error(err))
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

func (h Handler) dispatchAsyncEventsToConnections(cids []sdktypes.ConnectionID, e sdktypes.Event) {
	ctx := extrazap.AttachLoggerToContext(h.logger, context.Background())
	for _, cid := range cids {
		eid, err := h.dispatch(ctx, e.WithConnectionDestinationID(cid), nil)
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
