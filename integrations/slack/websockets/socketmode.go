package websockets

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	eventsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to receive and dispatch asynchronous event notifications.
type handler struct {
	logger        *zap.Logger
	secrets       sdkservices.Secrets
	dispatcher    sdkservices.Dispatcher
	integrationID sdktypes.IntegrationID
	scope         string
}

func NewHandler(l *zap.Logger, sec sdkservices.Secrets, d sdkservices.Dispatcher, scope string, id sdktypes.IntegrationID) handler {
	return handler{
		logger:        l,
		secrets:       sec,
		dispatcher:    d,
		scope:         scope,
		integrationID: id,
	}
}

// Key = Slack app ID (to ensure one WebSocket per app).
var webSocketClients = make(map[string]*socketmode.Client)

func (h handler) OpenSocketModeConnection(appID, botToken, appToken string) {
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
			if err != nil {
				h.logger.Error("Slack Socket Mode (re)connection error", zap.Error(err))
			}
			return
		}
	}()
}

func (h handler) socketModeHandler(e *socketmode.Event, c *socketmode.Client) {
	msg := fmt.Sprintf("Slack Socket Mode event: %s", string(e.Type))
	switch string(e.Type) {
	// WebSocket connection flow.
	case "connecting", "connected":
		h.logger.Debug(msg, zap.Any("data", e.Data))
	case "hello":
		h.logger.Debug(msg, zap.Any("request", e.Request))

	// TODO(ENG-549): slack-go handles "disconnect" events, but not robustly
	// at all (https://api.slack.com/apis/connections/socket#disconnect).

	// Events.
	case "events_api":
		h.logger.Debug(msg, zap.Any("data", e.Data), zap.Any("request", e.Request))
		h.handleBotEvent(e, c)
	case "slash_commands":
		h.logger.Debug(msg, zap.Any("data", e.Data), zap.Any("request", e.Request))
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

func (h handler) dispatchAsyncEventsToConnections(tokens []string, event *eventsv1.Event) {
	ctx := extrazap.AttachLoggerToContext(h.logger, context.Background())
	for _, connToken := range tokens {
		event.IntegrationToken = connToken
		event := kittehs.Must1(sdktypes.EventFromProto(event))
		eventID, err := h.dispatcher.Dispatch(ctx, event, nil)
		if err != nil {
			h.logger.Error("Dispatch failed",
				zap.String("connectionToken", connToken),
				zap.Error(err),
			)
			return
		}
		h.logger.Debug("Dispatched",
			zap.String("connectionToken", connToken),
			zap.String("eventID", eventID.String()),
		)
	}
}
