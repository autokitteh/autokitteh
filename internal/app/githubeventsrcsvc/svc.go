// This should probably be split into multiple parts: event source and general
// integration as it does several stuff.
package githubeventsrcsvc

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v42/github"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/autokitteh/autokitteh/api/gen/stubs/go/githubeventsrc"

	"github.com/autokitteh/autokitteh/sdk/api/apievent"
	"github.com/autokitteh/autokitteh/sdk/api/apieventsrc"
	"github.com/autokitteh/autokitteh/sdk/api/apiproject"
	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
	"github.com/autokitteh/autokitteh/internal/pkg/events"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsrcsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/githubinstalls"

	H "github.com/autokitteh/autokitteh/pkg/h"
	L "github.com/autokitteh/L"
)

// TODO: filter.
var EventTypes = []string{
	"check_run",
	"check_suite",
	"commit_comment",
	"create",
	"delete",
	"deployment",
	"deployment_review",
	"deployment_status",
	"discussion",
	"discussion_comment",
	"fork",
	"gollum",
	"issues",
	"issue_comment",
	"label",
	"member",
	"membership",
	"merge_queue_entry",
	"meta",
	"milestone",
	"organization",
	"org_block",
	"project",
	"project_card",
	"project_column",
	"public",
	"pull_request",
	"pull_request_review",
	"pull_request_review_comment",
	"pull_request_review_thread",
	"push",
	"release",
	"repository",
	"repository_dispatch",
	"star",
	"status",
	"team",
	"team_add",
	"watch",
	"workflow_dispatch",
	"workflow_job",
	"workflow_run",
}

type Config struct {
	EventSourceID apieventsrc.EventSourceID `envconfig:"EVENT_SOURCE_ID" json:"event_source_id"`
	WebhookSecret string                    `envconfig:"WEBHOOK_SECRET" json:"webhook_secret"`
	AppID         int64                     `envconfig:"APP_ID" json:"app_id"`
	AppPrivateKey string                    `envconfig:"APP_PRIVATE_KEY" json:"app_private_key"`
}

type Svc struct {
	pb.UnimplementedGithubEventSourceServer

	Config       Config
	Events       *events.Events
	EventSources eventsrcsstore.Store
	Installs     *githubinstalls.Installs
	L            L.Nullable
}

func (s *Svc) Register(ctx context.Context, srv *grpc.Server, gw *runtime.ServeMux, r *mux.Router) {
	if r != nil {
		r.HandleFunc("/githubsrc/event", s.httpEvent).Methods("POST")
		r.HandleFunc("/githubsrc/setup", s.httpSetup).Methods("GET")
	}

	if srv != nil {
		pb.RegisterGithubEventSourceServer(srv, s)
	}

	if gw != nil {
		if err := pb.RegisterGithubEventSourceHandlerServer(ctx, gw, s); err != nil {
			panic(err)
		}
	}
}

func (s *Svc) httpSetup(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	iid, action := q.Get("installation_id"), q.Get("action")

	l := s.L.With("installation_id", iid, "action", action)

	l.Debug("got setup")

	fmt.Fprintf(w, "Github app setup for installation %s\n", iid)
}

func (s *Svc) httpEvent(w http.ResponseWriter, r *http.Request) {
	webhookType, deliveryID := github.WebHookType(r), github.DeliveryID(r)

	l := s.L.With("webhook_type", webhookType, "delivery_id", deliveryID)

	l.Debug("got event")

	defer r.Body.Close()

	payload, err := github.ValidatePayload(r, []byte(s.Config.WebhookSecret))
	if err != nil {
		H.Error(l, w, http.StatusBadRequest, "validate payload error", "err", err)
		return
	}

	l.Debug("valid payload")

	event, err := github.ParseWebHook(webhookType, payload)
	if err != nil {
		H.Error(l, w, http.StatusBadRequest, "validate payload error", "err", err)
		return
	}

	l.Debug("valid event", "event", event)

	if event, ok := event.(*github.InstallationEvent); ok {
		if err := s.onInstallation(r.Context(), event); err != nil {
			H.Respond(l, w, err)
			return
		}

		H.WriteJSON(w, http.StatusOK, "=^.^=")
		return
	}

	pev, err := parseEvent(event)
	if err != nil {
		H.Respond(l, w, err)
		return
	}

	pev.Data["delivery_id"] = apivalues.String(deliveryID)

	l = l.With("installation_id", pev.Installation.ID, "owner", pev.Owner, "repo", pev.Repo)

	l.Debug("parsed event")

	eids := make(map[string]apievent.EventID, 2)

	if eids[pev.Owner], err = s.Events.IngestEvent(r.Context(), s.Config.EventSourceID, pev.Owner, deliveryID, webhookType, pev.Data, nil); err != nil {
		l.Error("owner ingest event error", "err", err)
	}

	if pev.Repo != "" {
		assoc := fmt.Sprintf("%s/%s", pev.Owner, pev.Repo)
		if eids[assoc], err = s.Events.IngestEvent(r.Context(), s.Config.EventSourceID, assoc, deliveryID, webhookType, pev.Data, nil); err != nil {
			l.Error("repo ingest event error", "err", err)
		}
	}

	l.Debug("sent event", "ids", eids)

	H.WriteJSON(w, http.StatusCreated, eids)
}

func (s *Svc) Add(ctx context.Context, pid apiproject.ProjectID, name, org, repo string) error {
	s.L.Debug("adding binding", "name", name, "project_id", pid, "org", org, "repo", repo)

	if repo != "" {
		org += "/"
	}

	// TODO: this is obviously insecture as one can listen to any repo.
	if err := s.EventSources.AddProjectBinding(
		ctx,
		s.Config.EventSourceID,
		pid,
		name,
		fmt.Sprintf("%s%s", org, repo), // assoc,
		"",                             // config
		true,
		(&apieventsrc.EventSourceProjectBindingSettings{}).SetEnabled(true),
	); err != nil {
		return fmt.Errorf("bind: %w", err)
	}

	return nil
}

func (s *Svc) Bind(ctx context.Context, req *pb.BindRequest) (*pb.BindResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	if err := s.Add(ctx, apiproject.ProjectID(req.ProjectId), req.Name, req.Org, req.Repo); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "add: %v", err)
	}

	return &pb.BindResponse{}, nil
}

func (s *Svc) Unbind(ctx context.Context, req *pb.UnbindRequest) (*pb.UnbindResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validate: %v", err)
	}

	// TODO

	return &pb.UnbindResponse{}, nil
}

func (s *Svc) onInstallation(ctx context.Context, event *github.InstallationEvent) error {
	if event.Installation == nil || event.Action == nil {
		return H.NewError(http.StatusBadRequest, "missing fields")
	}

	if *event.Action != "created" {
		// TODO: handle others. See doc at github.InstallationEvent.
		return nil
	}

	for _, repo := range event.Repositories {
		// for some reason repo.Owner is not populated, by repo.FullName is.
		var parts []string

		if fullName := repo.FullName; fullName != nil {
			parts = strings.SplitN(*fullName, "/", 2)
		}

		if len(parts) != 2 {
			return H.NewError(http.StatusBadRequest, "cannot deduce repo owner and name", "full_name", repo.FullName)
		}

		owner, name := parts[0], parts[1]

		l := s.L.With("owner", owner, "name", name)
		l.Debug("new installation")

		if err := s.Installs.Add(ctx, owner, name, event.Installation); err != nil {
			return H.NewError(http.StatusInternalServerError, "installation store error: %v", err)
		}
	}

	return nil
}
