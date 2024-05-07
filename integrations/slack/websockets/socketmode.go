package websockets

import (
	"context"
	"fmt"
	"sync"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	eventsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1"
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

func NewHandler(l *zap.Logger, sec sdkservices.Vars, d sdkservices.Dispatcher, id sdktypes.IntegrationID) handler {
	return handler{
		logger:        l,
		vars:          sec,
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

func (h handler) dispatchAsyncEventsToConnections(cids []sdktypes.ConnectionID, event *eventsv1.Event) {
	ctx := extrazap.AttachLoggerToContext(h.logger, context.Background())
	for _, cid := range cids {
		l := h.logger.With(zap.String("cid", cid.String()))

		event.ConnectionId = cid.String()
		event, err := sdktypes.EventFromProto(event)
		if err != nil {
			l.Error("Failed to convert protocol buffer to SDK event",
				zap.Any("event", event),
				zap.Error(err),
			)
			return
		}

		eventID, err := h.dispatcher.Dispatch(ctx, event, nil)
		if err != nil {
			l.Error("Dispatch failed", zap.Error(err))
			return
		}

		l.Debug("Dispatched", zap.String("eventID", eventID.String()))
	}
}
