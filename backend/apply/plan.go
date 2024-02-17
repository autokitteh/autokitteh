package apply

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// TODO: deletions? updates?
// TODO: rewrite this with clearer design patterns, documentation, and cleaner code.

func plan[T namer](ctx context.Context, what string, x T, xs []T, f func(context.Context, T) error,
) error {
	var zero T // nil
	if x != zero {
		xs = append(xs, x)
	}

	for i, x := range xs {
		if err := f(ctx, x); err != nil {
			return fmt.Errorf("%s %d %q: %w", what, i, x.name(), err)
		}
	}

	return nil
}

func (a *Applicator) Plan(ctx context.Context, root *Root) error {
	if root.Version != "" && root.Version != version {
		return fmt.Errorf("unhandled version %q != %q", root.Version, version)
	}

	if err := plan(ctx, "connection", root.Connection, root.Connections, a.planConnection); err != nil {
		return err
	}

	if err := plan(ctx, "env", root.Env, root.Envs, a.planEnv); err != nil {
		return err
	}

	if err := plan(ctx, "project", root.Project, root.Projects, a.planProject); err != nil {
		return err
	}

	return nil
}

func (a *Applicator) planConnection(ctx context.Context, conn *Connection) error {
	name, err := sdktypes.StrictParseName(conn.Name)
	if err != nil {
		return fmt.Errorf("name: %w", err)
	}

	intID, err := a.g().GetIntegrationID(ctx, conn.Integration)
	if err != nil {
		return fmt.Errorf("integration %q: %w", conn.Integration, err)
	}

	if intID == nil {
		return fmt.Errorf("integration %q: not found", conn.Integration)
	}

	var projectID sdktypes.ProjectID

	if conn.Project != "" {
		if projectID, err = a.g().GetProjectID(ctx, conn.Project); err != nil {
			return err
		}
	}

	var c sdktypes.Connection

	if projectID != nil || conn.Project == "" {
		conns, err := a.Svcs.Connections().List(ctx, sdkservices.ListConnectionsFilter{ProjectID: projectID})
		if err != nil {
			return fmt.Errorf("connections.list: %w", err)
		}

		_, c = kittehs.FindFirst(conns, func(conn sdktypes.Connection) bool {
			return sdktypes.GetConnectionName(conn).String() == name.String()
		})
	}

	token := conn.Token
	if conn.TokenEnvVar != "" {
		if a.LookupEnv == nil {
			a.log("os env var access is disabled, using default values").set("conn", conn.Name)

			if token == "" {
				return fmt.Errorf("env lookup is disabled")
			}
		} else if envVar, ok := a.LookupEnv(conn.TokenEnvVar); ok {
			token = envVar
		}
	}

	if c != nil {
		if sdktypes.GetConnectionIntegrationToken(c) != token {
			a.log("connection already exists, but token differs").
				set("conn_name", conn.Name).
				set("conn_id", sdktypes.GetConnectionID(c)).
				set("project_name", conn.Project).
				set("project_id", projectID)

			a.op(
				&Operation{
					Description: fmt.Sprintf("update connection %q token", conn.Name),
					Action: func(ctx context.Context) error {
						sdkConn := kittehs.Must1(sdktypes.ConnectionFromProto(&sdktypes.ConnectionPB{
							ConnectionId:     sdktypes.GetConnectionID(c).String(),
							IntegrationToken: token,
						}))

						if err := a.Svcs.Connections().Update(ctx, sdkConn); err != nil {
							return fmt.Errorf("connections.connect: %w", err)
						}

						a.log("connection token updated").
							set("conn_name", conn.Name).
							set("conn_id", sdktypes.GetConnectionID(c).String()).
							set("project_name", conn.Project).
							set("project_id", projectID)

						return nil
					},
				},
			)

			return nil
		}

		a.log("connection already exists with same token, not checking other differences").
			set("conn_name", conn.Name).
			set("conn_id", sdktypes.GetConnectionID(c)).
			set("project_name", conn.Project).
			set("project_id", projectID)

		return nil
	}

	a.op(
		&Operation{
			Description: fmt.Sprintf("create connection %q under %q", conn.Name, conn.Project),
			Action: func(ctx context.Context) error {
				if projectID == nil && conn.Project != "" {
					var err error

					if projectID, err = a.g().GetProjectID(ctx, conn.Project); err != nil {
						return err
					}

					if projectID == nil {
						return fmt.Errorf("project %q not found", conn.Project)
					}
				}

				sdkConn := kittehs.Must1(sdktypes.ConnectionFromProto(&sdktypes.ConnectionPB{
					Name:             name.String(),
					ProjectId:        projectID.String(),
					IntegrationToken: token,
					IntegrationId:    intID.String(),
				}))

				cid, err := a.Svcs.Connections().Create(ctx, sdkConn)
				if err != nil {
					return fmt.Errorf("connections.connect: %w", err)
				}

				a.log("connection created").
					set("conn_name", conn.Name).
					set("conn_id", cid).
					set("project_name", conn.Project).
					set("project_id", projectID)

				return nil
			},
		},
	)

	return nil
}

func (a *Applicator) planEnv(ctx context.Context, env *Env) error {
	name, err := sdktypes.StrictParseName(env.Name)
	if err != nil {
		return fmt.Errorf("name: %w", err)
	}

	var projectID sdktypes.ProjectID

	if env.Project != "" {
		if projectID, err = a.g().GetProjectID(ctx, env.Project); err != nil {
			return err
		}
	}

	work := new(struct{ env sdktypes.Env })

	if projectID != nil || env.Project == "" {
		envs, err := a.Svcs.Envs().List(ctx, projectID)
		if err != nil {
			return fmt.Errorf("connections.list: %w", err)
		}

		_, work.env = kittehs.FindFirst(envs, func(env sdktypes.Env) bool {
			return sdktypes.GetEnvName(env).String() == name.String()
		})
	}

	if work.env != nil {
		a.log("environment already exists").
			set("env_name", env.Name).
			set("env_id", sdktypes.GetEnvID(work.env)).
			set("project_name", env.Project).
			set("project_id", projectID)
	} else {
		a.op(
			&Operation{
				Description: fmt.Sprintf("create env %q", env.Name),
				Action: func(ctx context.Context) error {
					if projectID == nil && env.Project != "" {
						var err error

						if projectID, err = a.g().GetProjectID(ctx, env.Project); err != nil {
							return err
						}

						if projectID == nil {
							return fmt.Errorf("project %q not found", env.Project)
						}
					}

					sdkEnv := kittehs.Must1(sdktypes.EnvFromProto(&sdktypes.EnvPB{
						Name:      env.Name,
						ProjectId: projectID.String(),
					}))

					eid, err := a.Svcs.Envs().Create(ctx, sdkEnv)
					if err != nil {
						return fmt.Errorf("create: %w", err)
					}

					work.env = kittehs.Must1(sdkEnv.Update(func(pb *sdktypes.EnvPB) {
						pb.EnvId = eid.String()
					}))

					a.log("created env").
						set("name", env.Name).
						set("env_id", eid.String())

					return nil
				},
			},
		)
	}

	// TODO: for now always update args, but we might want to compare in the future.
	for _, v := range env.Vars {
		if err := func(v *Var) error {
			vn, err := sdktypes.StrictParseSymbol(v.Name)
			if err != nil {
				return fmt.Errorf("var %q: %w", v.Name, err)
			}

			val := v.Value
			if v.EnvVar != "" {
				if a.LookupEnv == nil {
					a.log("os env var access is disabled, using default values").set("env_id", sdktypes.GetEnvID(work.env).String()).set("var_name", vn.String())

					if val == "" {
						return fmt.Errorf("env lookup is disabled")
					}
				} else if envVar, ok := a.LookupEnv(v.EnvVar); ok {
					val = envVar
				}
			}

			a.op(&Operation{
				Description: fmt.Sprintf("env %q: set var %q", env.Name, v.Name),
				Action: func(ctx context.Context) error {
					sdkVar := kittehs.Must1(sdktypes.EnvVarFromProto(&sdktypes.EnvVarPB{
						EnvId:    sdktypes.GetEnvID(work.env).String(),
						Name:     vn.String(),
						Value:    val,
						IsSecret: v.IsSecret,
					}))

					if err := a.Svcs.Envs().SetVar(ctx, sdkVar); err != nil {
						return err
					}

					a.log("env var set").
						set("env_id", sdktypes.GetEnvID(work.env).String()).
						set("var_name", vn.String())

					return nil
				},
			})

			return nil
		}(v); err != nil {
			return err
		}
	}

	hasMapping := func(string) bool { return false }

	if work.env != nil {
		if len(env.Mappings) != 0 && sdktypes.GetEnvProjectID(work.env).Kind() != sdktypes.ProjectIDKind {
			return fmt.Errorf("mappings must belong to an environment with a project project id")
		}

		ms, err := a.Svcs.Mappings().List(ctx, sdkservices.ListMappingsFilter{EnvID: sdktypes.GetEnvID(work.env)})
		if err != nil {
			return fmt.Errorf("mappings.list: %w", err)
		}

		hasMapping = kittehs.ContainedIn(kittehs.TransformToStrings(kittehs.Transform(ms, sdktypes.GetMappingModuleName))...)
	}

	for _, m := range env.Mappings {
		if err := func(m *Mapping) error {
			if hasMapping(m.Name) {
				a.log("mapping already exists, not checking for updates").set("env_id", sdktypes.GetEnvID(work.env)).set("mapping_name", m.Name)
				return nil
			}

			sdkEvents, err := kittehs.TransformError(m.Events, func(e *MappingEvent) (sdktypes.MappingEvent, error) {
				cl, err := sdktypes.ParseCodeLocation(e.EntryPoint)
				if err != nil {
					return nil, fmt.Errorf("event: %w", err)
				}

				return sdktypes.MappingEventFromProto(&sdktypes.MappingEventPB{
					EventType:    e.EventType,
					CodeLocation: cl.ToProto(),
				})
			})
			if err != nil {
				return err
			}

			sdkMapping, err := sdktypes.MappingFromProto(&sdktypes.MappingPB{
				ModuleName: m.Name,
				Events:     kittehs.Transform(kittehs.FilterNils(sdkEvents), sdktypes.ToProto),
			})
			if err != nil {
				return fmt.Errorf("mapping %q: %w", m.Name, err)
			}

			a.op(&Operation{
				Description: fmt.Sprintf("env %q: add mapping %q", env.Name, m.Name),
				Action: func(ctx context.Context) error {
					if sdktypes.GetEnvProjectID(work.env).Kind() != sdktypes.ProjectIDKind {
						return fmt.Errorf("mapping must belong to an environment with a project project id")
					}

					connID, err := a.g().GetConnectionID(ctx, projectID.String(), m.Connection)
					if err != nil {
						return fmt.Errorf("connection %q: %w", m.Connection, err)
					}

					if connID == nil {
						return fmt.Errorf("connection %q: not found", m.Connection)
					}

					sdkMapping = kittehs.Must1(sdkMapping.Update(func(pb *sdktypes.MappingPB) {
						pb.EnvId = sdktypes.GetEnvID(work.env).String()
						pb.ConnectionId = connID.String()
					}))

					mid, err := a.Svcs.Mappings().Create(ctx, sdkMapping)
					if err != nil {
						return err
					}

					a.log("created mapping").
						set("env_id", sdktypes.GetEnvID(work.env).String()).
						set("mapping_id", mid.String()).
						set("mapping_name", m.Name)

					return nil
				},
			})

			return nil
		}(m); err != nil {
			return err
		}
	}

	return nil
}

func (a *Applicator) planProject(ctx context.Context, project *Project) error {
	name, err := sdktypes.StrictParseName(project.Name)
	if err != nil {
		return fmt.Errorf("name: %w", err)
	}

	work := new(struct{ project sdktypes.Project })

	work.project, err = a.Svcs.Projects().GetByName(ctx, name)
	if err != nil {
		return fmt.Errorf("projects.get_by_name: %w", err)
	}

	if work.project != nil {
		a.log("project already exists. not checking for differences.").
			set("project_name", project.Name).
			set("project_id", sdktypes.GetProjectID(work.project))
	} else {
		a.op(
			&Operation{
				Description: fmt.Sprintf("create project %q", project.Name),
				Action: func(ctx context.Context) error {
					rootPath := project.RootPath
					if rootPath == "" {
						if a.Path == "" {
							return errors.New("project root path must be specified")
						}

						rootPath = filepath.Dir(a.Path)
					}

					sdkProject := kittehs.Must1(sdktypes.ProjectFromProto(&sdktypes.ProjectPB{
						Name:             project.Name,
						ResourcesRootUrl: rootPath,
						ResourcePaths:    project.Paths,
					}))

					pid, err := a.Svcs.Projects().Create(ctx, sdkProject)
					if err != nil {
						return fmt.Errorf("create: %w", err)
					}

					work.project = kittehs.Must1(sdkProject.Update(func(pb *sdktypes.ProjectPB) {
						pb.ProjectId = pid.String()
					}))

					a.log("created project").
						set("name", project.Name).
						set("project_id", pid.String())

					return nil
				},
			},
		)
	}

	for _, conn := range project.Connections {
		if conn.Project != "" {
			return fmt.Errorf("conn %q: project must not be specified", conn.Name)
		}

		conn.Project = project.Name

		if err := a.planConnection(ctx, conn); err != nil {
			return err
		}
	}

	for _, env := range project.Envs {
		if env.Project != "" {
			return fmt.Errorf("env %q: project must not be specified", env.Name)
		}

		env.Project = project.Name

		if err := a.planEnv(ctx, env); err != nil {
			return err
		}
	}

	return nil
}
