package pluginsreg

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/samber/lo"

	"go.autokitteh.dev/sdk/api/apiaccount"
	"go.autokitteh.dev/sdk/api/apiplugin"
	"go.autokitteh.dev/sdk/plugin"
	"go.autokitteh.dev/sdk/plugin/builtinplugin"
	"go.autokitteh.dev/sdk/plugin/grpcplugin"
	"go.autokitteh.dev/sdk/pluginimpl"

	"github.com/autokitteh/L"
	"github.com/autokitteh/stores/pkvstore"

	"github.com/autokitteh/autokitteh/internal/pkg/akprocs"
)

var ErrNotFound = pkvstore.ErrNotFound

type Registry struct {
	L     L.Nullable
	Procs *akprocs.Procs

	Store pkvstore.Store

	InternalPlugins map[apiplugin.PluginName]*pluginimpl.Plugin

	sessionsMu sync.Mutex
	sessions   map[string][]*os.Process
}

func (s *Registry) startExec(id apiplugin.PluginID, sid string, exec *apiplugin.PluginExecSettings) (string, error) {
	cmd, addr, err := s.Procs.Start(
		exec.Name(),
		nil,
		map[string]string{
			"AK_PLUGIN_ID": id.String(),
		},
	)
	if err != nil {
		return "", err
	}

	if sid != "" {
		s.sessionsMu.Lock()
		defer s.sessionsMu.Unlock()

		if s.sessions == nil {
			s.sessions = make(map[string][]*os.Process, 128)
		}

		procs := s.sessions[sid]
		if procs == nil {
			procs = make([]*os.Process, 0, 16)
		}

		procs = append(procs, cmd.Process)

		s.sessions[sid] = procs
	}

	return addr, nil
}

func (s *Registry) CloseSession(sid string) {
	l := s.L.With("session_id", sid)

	l.Debug("closing session")

	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()

	go func(procs []*os.Process) {
		for _, proc := range procs {
			l := l.With("pid", proc.Pid)

			l.Debug("killing")

			if err := proc.Kill(); err != nil {
				l.Error("kill error", "err", err)
			}
		}
	}(s.sessions[sid])

	delete(s.sessions, sid)
}

func (s *Registry) NewPlugin(ctx context.Context, l L.L, id apiplugin.PluginID, sessionID string) (plugin.Plugin, error) {
	if id.IsInternal() {
		pl, ok := s.InternalPlugins[id.PluginName()]
		if !ok {
			return nil, ErrNotFound
		}

		return &builtinplugin.BuiltinPlugin{Plugin: pl, ID: id}, nil
	}

	pl, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	var addr string

	if pl.Settings().Exec().Name() != "" {
		if addr, err = s.startExec(id, sessionID, pl.Settings().Exec()); err != nil {
			return nil, fmt.Errorf("exec: %w", err)
		}
	} else {
		addr = fmt.Sprintf("%s:%d", pl.Settings().Address(), pl.Settings().Port())
	}

	return grpcplugin.NewFromHostPort(l, id, addr)
}

func (s *Registry) RegisterInternalPlugin(name apiplugin.PluginName, pl *pluginimpl.Plugin) {
	if s.InternalPlugins == nil {
		s.InternalPlugins = make(map[apiplugin.PluginName]*pluginimpl.Plugin)
	}

	s.InternalPlugins[name] = pl
}

func (s *Registry) RegisterExternalPlugin(ctx context.Context, id apiplugin.PluginID, settings *apiplugin.PluginSettings) error {
	p, err := apiplugin.NewPlugin(id, settings, time.Now(), nil)
	if err != nil {
		return fmt.Errorf("invalid data: %w", err)
	}

	bs, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	a, n := id.Split()

	return s.Store.Put(ctx, a.String(), n.String(), []byte(bs))
}

func (s *Registry) Get(ctx context.Context, id apiplugin.PluginID) (*apiplugin.Plugin, error) {
	a, n := id.Split()

	bs, err := s.Store.Get(ctx, a.String(), n.String())
	if err != nil {
		return nil, err
	}

	var p apiplugin.Plugin

	if err := json.Unmarshal(bs, &p); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return &p, nil
}

func (s *Registry) List(ctx context.Context, a *apiaccount.AccountName) ([]apiplugin.PluginID, error) {
	if a == nil {
		return nil, fmt.Errorf("account name must be specified") // TODO
	}

	names, err := s.Store.List(ctx, a.String())
	if err != nil {
		return nil, err
	}

	return lo.Map(names, func(name string, _ int) apiplugin.PluginID {
		return apiplugin.NewPluginID(*a, apiplugin.PluginName(name))
	}), nil
}
