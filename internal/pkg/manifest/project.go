package manifest

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/samber/lo"

	"go.autokitteh.dev/sdk/api/apiaccount"
	"go.autokitteh.dev/sdk/api/apieventsrc"
	"go.autokitteh.dev/sdk/api/apiplugin"
	"go.autokitteh.dev/sdk/api/apiprogram"
	"go.autokitteh.dev/sdk/api/apiproject"
	"go.autokitteh.dev/sdk/api/apivalues"

	"github.com/autokitteh/autokitteh/internal/pkg/akcue"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsrcsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/projectsstore"
)

type ProjectPlugin struct {
	Disabled bool `json:"disabled"`
}

type ProjectSourceBinding struct {
	SourceID     apieventsrc.EventSourceID `json:"src_id"`
	Assoc        string                    `json:"assoc"`
	SourceConfig string                    `json:"src_config"`
	Disabled     bool                      `json:"disabled"`
}

type Project struct {
	ID          string                          `json:"id"`
	AccountName string                          `json:"account_name"`
	Name        string                          `json:"name"`
	MainPath    string                          `json:"main_path"`
	Disabled    bool                            `json:"disabled"`
	Memo        map[string]string               `json:"memo"`
	Plugins     map[string]ProjectPlugin        `json:"plugins"`      // pluginID -> plugin
	Bindings    map[string]ProjectSourceBinding `json:"src_bindings"` // name -> binding
	Predecls    map[string]string               `json:"predecls"`     // TODO: allow more than just strings.
}

func ParseProject(ctx context.Context, src []byte) (*Project, error) {
	var p Project

	if err := akcue.Parse(ctx, src, &p); err != nil {
		return nil, err
	}

	return &p, nil
}

func (p Project) API(id string) (*apiproject.Project, error) {
	mainPath, err := apiprogram.ParsePathString(p.MainPath)
	if err != nil {
		return nil, fmt.Errorf("main path: %w", err)
	}

	plugins := make([]*apiproject.ProjectPlugin, 0, len(p.Plugins))

	for id, pl := range p.Plugins {
		id := apiplugin.PluginID(id)

		apipl, err := apiproject.NewProjectPlugin(id, !pl.Disabled)
		if err != nil {
			return nil, fmt.Errorf("plugin %q: %w", id, err)
		}

		plugins = append(plugins, apipl)
	}

	predecls := lo.MapValues(p.Predecls, func(v, _ string) *apivalues.Value {
		return apivalues.String(v)
	})

	if p.ID != "" {
		id = p.ID
	}

	accountName, projectName, _ := strings.Cut(id, ".")

	if p.AccountName != "" {
		accountName = p.AccountName
	}

	if p.Name != "" {
		projectName = p.Name
	}

	return apiproject.NewProject(
		apiproject.ProjectID(id),
		apiaccount.AccountName(accountName),
		(&apiproject.ProjectSettings{}).
			SetName(projectName).
			SetEnabled(!p.Disabled).
			SetMemo(p.Memo).
			SetMainPath(mainPath).
			SetPlugins(plugins).
			SetPredecls(predecls),
		time.Now(),
		nil,
	)
}

func (a Project) Compile(id string) ([]*Action, error) {
	api, err := a.API(id)
	if err != nil {
		return nil, fmt.Errorf("invalid project: %w", err)
	}

	acts := []*Action{
		{
			Desc: fmt.Sprintf("create project %q as %q", api.ID(), api.Settings().Name()),
			Run: func(ctx context.Context, env *Env) (string, error) {
				if env.Projects == nil {
					return "", fmt.Errorf("project %q: have no projects access", api.ID())
				}

				id, err := env.Projects.Create(ctx, api.AccountName(), api.ID(), api.Settings())
				if err != nil {
					if errors.Is(err, projectsstore.ErrAlreadyExists) {
						return fmt.Sprintf("project %q: already exists", api.ID()), nil
					}

					return "", err
				}

				return fmt.Sprintf("id=%s", id), err
			},
		},
	}

	for name, b := range a.Bindings {
		func(name string, b ProjectSourceBinding) {
			srcid := apieventsrc.EventSourceID(b.SourceID)

			acts = append(
				acts,
				&Action{
					Desc: fmt.Sprintf("bind eventsource %q to %q as %q", srcid, api.ID(), name),
					Run: func(ctx context.Context, env *Env) (string, error) {
						if env.EventSources == nil {
							return "", fmt.Errorf("have no event sources access")
						}

						err := env.EventSources.AddProjectBinding(
							ctx,
							srcid,
							api.ID(),
							name,
							b.Assoc,
							b.SourceConfig,
							true,
							(&apieventsrc.EventSourceProjectBindingSettings{}).SetEnabled(!b.Disabled),
						)

						if errors.Is(err, eventsrcsstore.ErrAlreadyExists) {
							return fmt.Sprintf("event source project binding %q already exists", api.ID()), nil
						}

						return "", err
					},
				},
			)
		}(name, b)
	}

	return acts, nil
}
