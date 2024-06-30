package websockets

import (
	"context"
	"fmt"
	"sync"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to receive and dispatch asynchronous event notifications.
type handler struct {
	logger        *zap.Logger
	vars          sdkservices.Vars
	dispatcher    sdkservices.Dispatcher
	integrationID sdktypes.IntegrationID
}

func NewHandler(l *zap.Logger, v sdkservices.Vars, d sdkservices.Dispatcher, id sdktypes.IntegrationID) handler {
	return handler{
		logger:        l,
		vars:          v,
		dispatcher:    d,
		integrationID: id,
	}
}

var (
	// Key = Slack app ID (to ensure one WebSocket per app).
	// Writes happen during server startup, and when a user
	// creates a new Socket Mode connection in the UI.
	webSocketClients = make(map[string]*socketmode.Client)

	mu = &sync.Mutex{}
)

func (h handler) OpenSocketModeConnection(appID, botToken, appToken string) {
	// Ensure multiple users don't reference the same app at the same time.
	mu.Lock()
	defer mu.Unlock()

	// No need to open multiple connections for the same app - yet.
	// See: https://docs.autokitteh.com/tutorials/new_connections/slack.
	if _, ok := webSocketClients[appID]; ok {
		return
	}

	api := slack.New(botToken, slack.OptionAppLevelToken(appToken))
	webSocketClients[appID] = socketmode.New(api)

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

func (h handler) socketModeHandler(e *socketmode.Event, c *socketmode.Client) {
	msg := fmt.Sprintf("Slack Socket Mode event: %s", string(e.Type))
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

	wrapped, err := sdktypes.DefaultValueWrapper.Wrap(slackEvent)
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
