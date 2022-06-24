package litterboxlocal

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"

	"go.autokitteh.dev/sdk/api/apiaccount"
	"go.autokitteh.dev/sdk/api/apievent"
	"go.autokitteh.dev/sdk/api/apieventsrc"
	"go.autokitteh.dev/sdk/api/apiprogram"
	"go.autokitteh.dev/sdk/api/apiproject"
	"golang.org/x/tools/txtar"

	"github.com/autokitteh/L"
	"github.com/autokitteh/autokitteh/internal/pkg/events"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsrcsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/litterbox"
	"github.com/autokitteh/autokitteh/internal/pkg/manifest"
	"github.com/autokitteh/autokitteh/internal/pkg/programs/loaders"
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

func (lb *LitterBox) Loader(ctx context.Context, path *apiprogram.Path) ([]byte, string, error) {
	lbid, actual, _ := strings.Cut(path.Path(), "/")

	root, err := apiprogram.NewPath("litterbox", lbid, "")
	if err != nil {
		return nil, "", fmt.Errorf("invalid path")
	}

	bs, err := lb.ProgramsStore.Get(ctx, root.String())
	if err != nil {
		if errors.Is(err, kvstore.ErrNotFound) {
			return nil, "", loaders.ErrNotFound
		}

		return nil, "", fmt.Errorf("get %q: %w", root.String(), err)
	}

	arch := txtar.Parse(bs)
	if len(arch.Files) == 0 {
		if actual == "auto.kitteh" {
			return bs, fmt.Sprintf("%x", sha256.Sum256(bs)), nil
		}

		return nil, "", loaders.ErrNotFound
	}

	for _, f := range arch.Files {
		if f.Name == actual {
			return f.Data, fmt.Sprintf("%x", sha256.Sum256(f.Data)), nil
		}
	}

	return nil, "", loaders.ErrNotFound
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
	files []byte,
) (litterbox.LitterBoxID, error) {
	arch := txtar.Parse(files)
	if len(arch.Files) == 0 {
		if len(files) == 0 {
			return "", litterbox.ErrNoSources
		}

		arch.Files = []txtar.File{
			{
				Name: "auto.kitteh",
				Data: files,
			},
		}
	}

	pid := lb.projectID(id)

	id = litterbox.LitterBoxID(pid.Unique())

	root := apiprogram.MustNewPath("litterbox", string(id), "")

	var mainPath *apiprogram.Path

	for _, f := range arch.Files {
		if f.Name == "project.cue" {
			continue
		}

		main := arch.Files[0].Name
		mainPath_, err := apiprogram.NewPath("", main, "")
		if err != nil {
			return "", fmt.Errorf("invalid main path %q: %w", main, err)
		}

		if mainPath, err = apiprogram.JoinPaths(root, mainPath_); err != nil {
			return "", fmt.Errorf("invalid main path %q: %w", main, err)
		}

		break
	}

	if err := lb.ProgramsStore.Put(ctx, root.String(), files); err != nil {
		return "", fmt.Errorf("program store: %w", err)
	}

	var manifestSrc []byte
	for _, f := range arch.Files {
		if f.Name == "project.cue" {
			manifestSrc = f.Data
			break
		}
	}

	defaultProject := &manifest.Project{
		MainPath:    mainPath.String(),
		Name:        fmt.Sprintf("litterbox_%s", id),
		AccountName: pid.AccountName().String(),
		Bindings: map[string]manifest.ProjectSourceBinding{
			"litterbox": {
				SourceID: lb.EventSourceID(),
				Assoc:    fmt.Sprintf("litterbox:%s", id),
			},
		},
	}

	p := defaultProject

	if manifestSrc != nil {
		var tags []string

		addTag := func(k, v string) {
			// TODO: *HACK* For some reason if the manifest does not
			// contain mentioning of the tag, cue will scream.
			if strings.Contains(string(manifestSrc), fmt.Sprintf("@tag(%s)", k)) {
				tags = append(tags, fmt.Sprintf("%s=%s", k, v))
			}
		}

		addTag("project_id", pid.String())
		addTag("project_name", pid.Unique()) // TODO: not really project name, but ehh.
		addTag("account_name", pid.AccountName().String())

		var err error
		if p, err = manifest.ParseProject(
			ctx,
			manifestSrc,
			tags,
		); err != nil {
			return "", fmt.Errorf("invalid manifest: %w", err)
		}

	}

	if p.MainPath == "" {
		p.MainPath = defaultProject.MainPath
	}

	if p.Name == "" {
		p.Name = defaultProject.Name
	}

	if p.AccountName == defaultProject.AccountName {
		p.AccountName = defaultProject.AccountName
	}

	if _, ok := p.Bindings["litterbox"]; !ok {
		p.Bindings["litterbox"] = defaultProject.Bindings["litterbox"]
	}

	acts, err := p.Compile(pid.String())
	if err != nil {
		return "", fmt.Errorf("compile manifest: %w", err)
	}

	env := manifest.Env{
		Projects:     lb.Projects,
		EventSources: lb.EventSrcs,
	}

	for _, act := range acts {
		msg, err := act.Run(ctx, &env)
		if err != nil {
			return "", fmt.Errorf("%q: manifest action error: %w", act.Desc, err)
		}

		lb.L.Debug("manifest action run", "msg", msg)
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

func (lb *LitterBox) Get(ctx context.Context, id litterbox.LitterBoxID) ([]byte, error) {
	root := apiprogram.MustNewPath("litterbox", string(id), "")

	src, err := lb.ProgramsStore.Get(ctx, root.String())
	if errors.Is(err, kvstore.ErrNotFound) {
		return nil, litterbox.ErrNotFound
	}

	return src, err
}

func (lb *LitterBox) Scoop(ctx context.Context, id litterbox.LitterBoxID) error {
	// TODO
	return nil
}
