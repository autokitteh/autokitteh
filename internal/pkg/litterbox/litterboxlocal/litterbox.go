package litterboxlocal

import (
	"context"
	"errors"
	"fmt"

	"go.autokitteh.dev/sdk/api/apiaccount"
	"go.autokitteh.dev/sdk/api/apievent"
	"go.autokitteh.dev/sdk/api/apieventsrc"
	"go.autokitteh.dev/sdk/api/apiprogram"
	"go.autokitteh.dev/sdk/api/apiproject"

	"github.com/autokitteh/L"
	"github.com/autokitteh/autokitteh/internal/pkg/events"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsrcsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/litterbox"
	"github.com/autokitteh/autokitteh/internal/pkg/projectsstore"
	"github.com/autokitteh/pubsub"
	"github.com/autokitteh/stores/kvstore"
)

type Config struct {
	AccountName string `envconfig:"ACCOUNT_NAME" default:"litterbox" json:"account_name"`
}

type LitterBox struct {
	Config        Config
	Projects      projectsstore.Store
	Events        *events.Events
	EventSrcs     eventsrcsstore.Store
	PubSub        pubsub.PubSub
	ProgramsStore kvstore.Store
	L             L.Nullable
}

var _ litterbox.LitterBox = &LitterBox{}

func (lb *LitterBox) EventSourceID() apieventsrc.EventSourceID {
	return apieventsrc.EventSourceID(fmt.Sprintf("%s.litterbox", lb.Config.AccountName))
}

func (lb *LitterBox) Loader(ctx context.Context, path *apiprogram.Path) ([]byte, error) {
	return lb.ProgramsStore.Get(ctx, path.String())
}

func (lb *LitterBox) projectID(id litterbox.LitterBoxID) apiproject.ProjectID {
	return apiproject.NewProjectID(
		apiaccount.AccountName(lb.Config.AccountName),
		string(id),
	)
}

func (lb *LitterBox) Setup(
	ctx context.Context,
	id litterbox.LitterBoxID,
	sources map[string][]byte,
	main string,
) (litterbox.LitterBoxID, error) {
	if len(sources) == 0 {
		return "", litterbox.ErrNoSources
	}

	if main == "" {
		if len(sources) > 1 {
			return "", litterbox.ErrMainNotSpecified
		}

		for k := range sources {
			main = k
		}
	}

	pid := lb.projectID(id)

	id = litterbox.LitterBoxID(pid.Unique())

	root := apiprogram.MustNewPath("litterbox", string(id), "")

	mainPath_, err := apiprogram.NewPath("", main, "")
	if err != nil {
		return "", fmt.Errorf("invalid main path %q: %w", main, err)
	}

	mainPath, err := apiprogram.JoinPaths(root, mainPath_)
	if err != nil {
		return "", fmt.Errorf("invalid main path %q: %w", main, err)
	}

	for k, v := range sources {
		path_, err := apiprogram.NewPath("", k, "")
		if err != nil {
			return "", fmt.Errorf("invalid path %q: %w", k, err)
		}

		path, err := apiprogram.JoinPaths(root, path_)
		if err != nil {
			return "", fmt.Errorf("invalid path %q: %w", k, err)
		}

		if err := lb.ProgramsStore.Put(ctx, path.String(), v); err != nil {
			return "", fmt.Errorf("program store: %w", err)
		}
	}

	settings := (&apiproject.ProjectSettings{}).
		SetEnabled(true).
		SetName(fmt.Sprintf("litterbox_%s", id)).
		SetMainPath(mainPath)

	if _, err := lb.Projects.Create(
		ctx,
		pid.AccountName(),
		pid,
		settings,
	); err != nil {
		if errors.Is(err, projectsstore.ErrAlreadyExists) {
			return id, nil
		}

		return "", fmt.Errorf("create project: %w", err)
	}

	if err := lb.EventSrcs.AddProjectBinding(
		ctx,
		lb.EventSourceID(),
		pid,
		"litterbox",
		fmt.Sprintf("litterbox:%s", id),
		"",
		true,
		(&apieventsrc.EventSourceProjectBindingSettings{}).SetEnabled(true),
	); err != nil {
		return "", fmt.Errorf("add source binding: %w", err)
	}

	return id, nil
}

func (lb *LitterBox) RunEvent(
	ctx context.Context,
	id litterbox.LitterBoxID,
	event *litterbox.LitterBoxEvent,
	ch chan<- *apievent.TrackIngestEventUpdate,
) (err error) {
	eid := apievent.NewEventID()

	l := lb.L.With("id", id, "event_id", eid)

	l.Debug("running event")

	if err := lb.Events.TrackIngestEvent(
		ctx,
		ch,
		eid,
		lb.EventSourceID(),
		fmt.Sprintf("litterbox:%s,%s", id, event.Src), // [# litterbox-assoc #]
		event.OriginalID,
		event.Type,
		event.Data,
		nil,
	); err != nil {
		return fmt.Errorf("ingest: %w", err)
	}

	return nil
}

func (lb *LitterBox) Run(
	ctx context.Context,
	id litterbox.LitterBoxID,
	ch chan<- *apievent.TrackIngestEventUpdate,
) (err error) {
	eid := apievent.NewEventID()

	l := lb.L.With("id", id, "event_id", eid)

	l.Debug("running")

	if err := lb.Events.MonitorProjectEvents(
		ctx,
		ch,
		lb.projectID(id),
	); err != nil {
		return fmt.Errorf("ingest: %w", err)
	}

	return nil
}

func (lb *LitterBox) Scoop(ctx context.Context, id litterbox.LitterBoxID) error {
	// TODO
	return nil
}
