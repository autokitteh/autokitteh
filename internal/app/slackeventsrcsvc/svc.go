package slackeventsrcsvc

import (
	"context"
	"fmt"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/slack-go/slack"
	"google.golang.org/grpc"

	pb "github.com/autokitteh/autokitteh/api/gen/stubs/go/slackeventsrc"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apieventsrc"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiproject"
	"github.com/autokitteh/autokitteh/internal/pkg/credsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/events"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsrcsstore"
	L "github.com/autokitteh/autokitteh/pkg/l"
)

// TODO
var EventTypes = []string{}

type Config struct {
	EventSourceID apieventsrc.EventSourceID `envconfig:"ID" json:"event_source_id"`
	SocketMode    bool                      `envconfig:"SOCKET_MODE" json:"socket_mode"`
	Debug         bool                      `envconfig:"DEBUG" json:"debug"`
	OAuthEnabled  bool                      `envconfig:"OAUTH_ENABLED" json:"oauth_enabled"`
	ClientID      string                    `envconfig:"CLIENT_ID" json:"client_id"`
	ClientSecret  string                    `envconfig:"CLIENT_SECRET" json:"client_secret"`
	SigningSecret string                    `envconfig:"SIGNING_SECRET" json:"signing_secret"`
	AppToken      string                    `envconfig:"APP_TOKEN" json:"app_token"`
	BotToken      string                    `envconfig:"BOT_TOKEN" json:"bot_token"`
}

type Svc struct {
	pb.UnimplementedSlackEventSourceServer

	Config Config

	Events       *events.Events
	EventSources eventsrcsstore.Store
	CredsStore   *credsstore.Store

	L L.Nullable

	client *slack.Client
}

func (s *Svc) Start(ctx context.Context, srv *grpc.Server, gw *runtime.ServeMux, r *mux.Router) error {
	if s.Config.EventSourceID.Empty() {
		return fmt.Errorf("event source id not configured")
	}

	s.L.Debug("started")

	s.register(ctx, srv, gw, r)

	s.client = slack.New(
		s.Config.BotToken,
		slack.OptionAppLevelToken(s.Config.AppToken),
		slack.OptionDebug(s.Config.Debug),
		slack.OptionLog(wrapLogger(s.L)),
	)

	if s.Config.SocketMode {
		s.startSocketMode()
		return nil
	}

	return nil
}

func (s *Svc) register(ctx context.Context, srv *grpc.Server, gw *runtime.ServeMux, r *mux.Router) {
	if srv != nil {
		pb.RegisterSlackEventSourceServer(srv, s)
	}

	if gw != nil {
		if err := pb.RegisterSlackEventSourceHandlerServer(ctx, gw, s); err != nil {
			panic(err)
		}
	}

	s.initHTTP(r)
}

func (s *Svc) Remove(context.Context, apiproject.ProjectID, string) error {
	// nothing to do essentially.
	return nil
}

func (s *Svc) Add(ctx context.Context, pid apiproject.ProjectID, name, tid string) error {
	if err := s.EventSources.AddProjectBinding(
		ctx,
		s.Config.EventSourceID,
		pid,
		name,
		string(tid),
		"",
		true,
		(&apieventsrc.EventSourceProjectBindingSettings{}).SetEnabled(true),
	); err != nil {
		return fmt.Errorf("add binding: %w", err)
	}

	return nil
}
