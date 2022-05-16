package manifest

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiplugin"
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

func (a Plugin) API() (*apiplugin.Plugin, error) {
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

	return apiplugin.NewPlugin(
		apiplugin.PluginID(a.ID),
		s,
		time.Now(),
		nil,
	)
}

func (a Plugin) Compile() ([]*Action, error) {
	api, err := a.API()
	if err != nil {
		return nil, fmt.Errorf("invalid plugin: %w", err)
	}

	return []*Action{{
		Desc: fmt.Sprintf("create plugin %s", api.ID()),
		Run: func(ctx context.Context, env *Env) (string, error) {
			err := env.Plugins.RegisterExternalPlugin(ctx, api.ID(), api.Settings())
			if err != nil {
				return "failed", err
			}

			return "created", nil
		},
	}}, nil
}
