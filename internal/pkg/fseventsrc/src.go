package fseventsrc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"

	"github.com/autokitteh/L"
	"github.com/autokitteh/autokitteh/internal/pkg/events"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsrcsstore"
	"go.autokitteh.dev/sdk/api/apieventsrc"
	"go.autokitteh.dev/sdk/api/apiproject"
	"go.autokitteh.dev/sdk/api/apivalues"
)

var EventTypes = []string{"create", "write", "remove", "rename", "chmod"}

const AllOps = fsnotify.Create | fsnotify.Write | fsnotify.Rename | fsnotify.Remove | fsnotify.Chmod

var fsOpsList = []fsnotify.Op{fsnotify.Create, fsnotify.Write, fsnotify.Rename, fsnotify.Remove, fsnotify.Chmod}

var ErrNotFound = errors.New("not found")

type Config struct {
	EventSourceID apieventsrc.EventSourceID `envconfig:"EVENT_SOURCE_ID" json:"event_source_id"`
}

type bindingConfig struct {
	assoc     string
	projectID apiproject.ProjectID
	name      string

	Path    string      `json:"path"`
	OpsMask fsnotify.Op `json:"ops_mask"`
}

// NOTE: this is not built for a large number of bindings.
// TODO: should it be?
type FSEventSource struct {
	Config       Config
	L            L.Nullable
	Events       *events.Events
	EventSources eventsrcsstore.Store

	mu           sync.RWMutex
	bindings     []*bindingConfig
	addCh, remCh chan string
}

func (s *FSEventSource) Start(ctx context.Context) error {
	s.L.Debug("started")

	if s.Config.EventSourceID.Empty() {
		return fmt.Errorf("event source id not configured")
	}

	bs, err := s.EventSources.GetProjectBindings(ctx, &s.Config.EventSourceID, nil, "", "", true)
	if err != nil {
		return fmt.Errorf("get bindings: %w", err)
	}

	s.bindings = make([]*bindingConfig, 0, len(bs))

	for _, b := range bs {
		l := s.L.With("project_id", b.ProjectID(), "name", b.Name())

		var cfg bindingConfig

		if err := json.Unmarshal([]byte(b.SourceConfig()), &cfg); err != nil {
			l.Error("failed unmarshalling binding config")
		}

		cfg.projectID = b.ProjectID()
		cfg.assoc = b.AssociationToken()
		cfg.name = b.Name()
		s.bindings = append(s.bindings, &cfg)
	}

	return s.watch()
}

func (s *FSEventSource) Remove(
	ctx context.Context,
	pid apiproject.ProjectID,
	name string,
	path string,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, b := range s.bindings {
		if b.projectID != pid || b.name != name {
			continue
		}

		if path != "" && b.Path != path {
			continue
		}

		s.bindings[i] = s.bindings[len(s.bindings)-1]
		s.bindings = s.bindings[:len(s.bindings)-1]
		go func() { s.remCh <- path }()
		return nil
	}

	return ErrNotFound
}

func (s *FSEventSource) Add(
	ctx context.Context,
	pid apiproject.ProjectID,
	name string,
	path string,
	mask fsnotify.Op,
) error {
	b := &bindingConfig{
		projectID: pid,
		name:      name,
		assoc:     pid.String(),
		Path:      path,
		OpsMask:   mask,
	}

	cfg, err := json.Marshal(&b)
	if err != nil {
		s.L.Panic("binding marshal error", "err", err)
	}

	if err := s.EventSources.AddProjectBinding(
		ctx,
		s.Config.EventSourceID,
		pid,
		name,
		pid.String(),
		string(cfg),
		true,
		(&apieventsrc.EventSourceProjectBindingSettings{}).SetEnabled(true),
	); err != nil {
		return fmt.Errorf("add binding: %w", err)
	}

	s.mu.Lock()
	s.bindings = append(s.bindings, b)
	s.mu.Unlock()

	s.addCh <- path

	s.L.Debug("binding added", "binding", b)

	return nil
}

func (s *FSEventSource) watch() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("new watcher: %w", err)
	}

	s.addCh = make(chan string, 1)
	s.remCh = make(chan string, 1)

	go func() {
		l := s.L

		for {
			select {
			case path := <-s.addCh:
				l.Debug("adding path to watcher", "path", path)

				if err := watcher.Add(path); err != nil {
					l.Error("add path failed", "path", path, "err", err)
				}

			case path := <-s.remCh:
				l.Debug("removing path to watcher", "path", path)

				if err := watcher.Remove(path); err != nil {
					l.Error("remove path failed", "path", path, "err", err)
				}

			case ev := <-watcher.Events:
				s.rx(context.Background(), &ev)

			case err := <-watcher.Errors:
				l.Error("received error from watcher", "err", err)
			}
		}
	}()

	// When starting with a persistent db, fetch all previously added paths.
	for _, b := range s.bindings {
		if err := watcher.Add(b.Path); err != nil {
			// TODO: send error event?
			s.L.Error("watcher add failed", "err", err, "project_id", b.projectID)
		}
	}

	return nil
}

func (s *FSEventSource) rx(ctx context.Context, ev *fsnotify.Event) {
	l := s.L.With("event", ev)

	l.Debug("received event")

	s.mu.RLock()
	defer s.mu.RUnlock()

	// association is really by project id in fsnotify, but if it'll change this
	// will hold true.
	assocs := make(map[string]bool, len(s.bindings))

	// (a relevant binding might not be found - that's ok since it might have
	// been removed)
	for _, b := range s.bindings {
		l := l.With("path", b.Path, "project_id", b.projectID, "mask", b.OpsMask)

		if rel, err := filepath.Rel(b.Path, ev.Name); err != nil || rel[0] == '.' {
			continue
		}

		if b.OpsMask&ev.Op == 0 {
			continue
		}

		l.Debug("match")

		assocs[b.assoc] = true
	}

	allops := strings.Split(strings.ToLower(ev.Op.String()), "|")
	lops := make(apivalues.ListValue, len(allops))
	for i, op := range allops {
		lops[i] = apivalues.String(op)
	}

	for assoc := range assocs {
		l := l.With("assoc", assoc)

		for _, op := range fsOpsList {
			if op&ev.Op == 0 {
				continue
			}

			opstr := strings.ToLower(op.String())

			l := l.With("op", opstr)

			eid, err := s.Events.IngestEvent(
				ctx,
				"",
				s.Config.EventSourceID,
				assoc,
				"", // no original id
				opstr,
				map[string]*apivalues.Value{
					"op":          apivalues.String(opstr),
					"related_ops": apivalues.MustNewValue(lops),
					"path":        apivalues.String(ev.Name),
				},
				map[string]string{
					"description": fmt.Sprintf("%q: %s", ev.Name, ev.Op.String()),
				},
			)
			if err != nil {
				l.Error("ingest error: %w", err)
			}

			l.Debug("dispatched", "event_id", eid)
		}
	}
}
