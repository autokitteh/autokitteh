package litterboxlocal

import (
	"context"
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

func (lb *LitterBox) idFromProjectID(pid apiproject.ProjectID) litterbox.LitterBoxID {
	return litterbox.LitterBoxID(pid.Unique())
}

func (lb *LitterBox) projectIDFromID(id litterbox.LitterBoxID) apiproject.ProjectID {
	return apiproject.NewProjectID(apiaccount.AccountName(lb.Config.AccountName), string(id))
}

func (lb *LitterBox) Loader(ctx context.Context, path *apiprogram.Path) ([]byte, error) {
	return lb.ProgramsStore.Get(ctx, path.String())
}

func (lb *LitterBox) Setup(ctx context.Context, name, source string) (litterbox.LitterBoxID, error) {
	pid := apiproject.NewProjectID(apiaccount.AccountName(lb.Config.AccountName), name)
	id := lb.idFromProjectID(pid)

	path := apiprogram.MustNewPath("litterbox", string(id), "")

	if err := lb.ProgramsStore.Put(ctx, path.String(), []byte(source)); err != nil {
		return "", fmt.Errorf("program store: %w", err)
	}

	settings := (&apiproject.ProjectSettings{}).
		SetEnabled(true).
		SetName(name).
		SetMainPath(path)

	if _, err := lb.Projects.Create(
		ctx,
		pid.AccountName(),
		pid,
		settings,
	); err != nil {
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
		&apieventsrc.EventSourceProjectBindingSettings{},
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
		fmt.Sprintf("litterbox:%s", id),
		fmt.Sprintf("%s/%s", event.SrcBinding, event.OriginalID),
		event.Type,
		event.Data,
		nil,
	); err != nil {
		return fmt.Errorf("ingest: %w", err)
	}

	return nil
}

func (lb *LitterBox) Scoop(ctx context.Context, id litterbox.LitterBoxID) error {
	// TODO
	return nil
}
