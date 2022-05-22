package manifest

import (
	"context"
	"fmt"
	"time"

	"github.com/autokitteh/autokitteh/sdk/api/apiplugin"
)

type Plugin struct {
	ID       apiplugin.PluginID `json:"id"`
	Address  string             `json:"address"`
	Port     uint16             `json:"port"`
	Disabled bool               `json:"disabled"`
	Exec     *struct {
		Name string `json:"name"`
	} `json:"exec"`
}

func (a Plugin) API(id string) (*apiplugin.Plugin, error) {
	s := (&apiplugin.PluginSettings{}).SetEnabled(!a.Disabled)

	if a.Port != 0 {
		s = s.SetPort(a.Port)
	}

	if a.Address != "" {
		s = s.SetAddress(a.Address)
	}

	if a.Exec != nil {
		s = s.SetExec((&apiplugin.PluginExecSettings{}).SetName(a.Exec.Name))
	}

	if a.ID != "" {
		id = a.ID.String()
	}

	return apiplugin.NewPlugin(
		apiplugin.PluginID(id),
		s,
		time.Now(),
		nil,
	)
}

func (a Plugin) Compile(id string) ([]*Action, error) {
	api, err := a.API(id)
	if err != nil {
		return nil, fmt.Errorf("invalid plugin: %w", err)
	}

	return []*Action{{
		Desc: fmt.Sprintf("create plugin %q", api.ID()),
		Run: func(ctx context.Context, env *Env) (string, error) {
			err := env.Plugins.RegisterExternalPlugin(ctx, api.ID(), api.Settings())
			if err != nil {
				return "failed", err
			}

			return "created", nil
		},
	}}, nil
}
