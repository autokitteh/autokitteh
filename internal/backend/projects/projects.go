package projects

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Projects struct {
	fx.In

	Z            *zap.Logger
	DB           db.DB
	Builds       sdkservices.Builds
	Runtimes     sdkservices.Runtimes
	Integrations sdkservices.Integrations
}

func New(p Projects, telemetry *telemetry.Telemetry) sdkservices.Projects {
	initMetrics(telemetry)
	return &p
}

func (ps *Projects) Create(ctx context.Context, project sdktypes.Project) (sdktypes.ProjectID, error) {
	project = project.WithNewID()

	if !project.Name().IsValid() {
		project = project.WithName(sdktypes.NewRandomSymbol())
	}

	if err := project.Strict(); err != nil {
		return sdktypes.InvalidProjectID, err
	}

	env := kittehs.Must1(sdktypes.EnvFromProto(&sdktypes.EnvPB{ProjectId: project.ID().String(), Name: "default"}))
	env = env.WithNewID()

	if err := ps.DB.Transaction(ctx, func(tx db.DB) error {
		if err := tx.CreateProject(ctx, project); err != nil {
			return err
		}
		return tx.CreateEnv(ctx, env)
	}); err != nil {
		return sdktypes.InvalidProjectID, err
	}

	projectsCreatedCounter.Add(ctx, 1)
	return project.ID(), nil
}

func (ps *Projects) Delete(ctx context.Context, pid sdktypes.ProjectID) error {
	return ps.DB.DeleteProject(ctx, pid)
}

func (ps *Projects) Update(ctx context.Context, project sdktypes.Project) error {
	return ps.DB.UpdateProject(ctx, project)
}

func (ps *Projects) GetByID(ctx context.Context, pid sdktypes.ProjectID) (sdktypes.Project, error) {
	return ps.DB.GetProjectByID(ctx, pid)
}

func (ps *Projects) GetByName(ctx context.Context, n sdktypes.Symbol) (sdktypes.Project, error) {
	return ps.DB.GetProjectByName(ctx, n)
}

func (ps *Projects) List(ctx context.Context) ([]sdktypes.Project, error) {
	return ps.DB.ListProjects(ctx)
}

func (ps *Projects) Build(ctx context.Context, projectID sdktypes.ProjectID) (sdktypes.BuildID, error) {
	fs, err := ps.openProjectResourcesFS(ctx, projectID)
	if err != nil {
		return sdktypes.InvalidBuildID, err
	}

	if fs == nil {
		return sdktypes.InvalidBuildID, errors.New("no resources set")
	}

	bi, err := sdkruntimes.Build(
		ctx,
		ps.Runtimes,
		fs,
		nil,
		nil,
	)
	if err != nil {
		return sdktypes.InvalidBuildID, err
	}

	var buf bytes.Buffer

	if err := bi.Write(&buf); err != nil {
		return sdktypes.InvalidBuildID, err
	}

	return ps.Builds.Save(ctx, sdktypes.NewBuild().WithProjectID(projectID), buf.Bytes())
}

func (ps *Projects) SetResources(ctx context.Context, projectID sdktypes.ProjectID, resources map[string][]byte) error {
	return ps.DB.SetProjectResources(ctx, projectID, resources)
}

func (ps *Projects) DownloadResources(ctx context.Context, projectID sdktypes.ProjectID) (map[string][]byte, error) {
	return ps.DB.GetProjectResources(ctx, projectID)
}

func (ps *Projects) Export(ctx context.Context, projectID sdktypes.ProjectID) ([]byte, error) {
	const manifestFileName = "autokitteh.yaml"
	hasManifest := false
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	rscs, err := ps.DownloadResources(ctx, projectID)
	if err != nil {
		return nil, err
	}

	for name, data := range rscs {
		f, err := w.Create(name)
		if err != nil {
			return nil, err
		}
		_, err = f.Write(data)
		if err != nil {
			return nil, err
		}
		if name == manifestFileName {
			hasManifest = true
		}
	}

	if !hasManifest {
		manifest, err := ps.exportManifest(ctx, projectID)
		if err != nil {
			return nil, err
		}
		f, err := w.Create("autokitteh.yaml")
		if err != nil {
			return nil, err
		}
		_, err = f.Write(manifest)
		if err != nil {
			return nil, err
		}
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (ps *Projects) exportManifest(ctx context.Context, projectID sdktypes.ProjectID) ([]byte, error) {
	prj, err := ps.DB.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	fmt.Fprintln(&buf, "version: v1")
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "project:")
	fmt.Fprintf(&buf, "  name: %s\n", prj.Name().String())

	cf := sdkservices.ListConnectionsFilter{
		ProjectID: prj.ID(),
	}
	conns, err := ps.DB.ListConnections(ctx, cf, false)
	if err != nil {
		return nil, err
	}

	if len(conns) > 0 {
		fmt.Fprintln(&buf, "  connections:")
		for _, c := range conns {
			name := c.Name().String()
			fmt.Fprintf(&buf, "    - name: %s\n", yamlize(name))
			integ, err := ps.Integrations.GetByID(ctx, c.IntegrationID())
			if err != nil {
				return nil, err
			}
			fmt.Fprintf(&buf, "      integration: %s\n", yamlize(integ.UniqueName().String()))
		}
	}

	tf := sdkservices.ListTriggersFilter{
		ProjectID: prj.ID(),
	}
	triggers, err := ps.DB.ListTriggers(ctx, tf)
	if err != nil {
		return nil, err
	}

	fmt.Fprintln(&buf, "  triggers:")

	for _, t := range triggers {
		fmt.Fprintf(&buf, "    - name: %s\n", t.Name().String())
		fmt.Fprintf(&buf, "      call: %s\n", t.CodeLocation().CanonicalString())
		if filter := t.Filter(); filter != "" {
			fmt.Fprintf(&buf, "      filter: %s\n", filter)
		}
		if etype := t.EventType(); etype != "" {
			fmt.Fprintf(&buf, "      event_type: %s\n", etype)
		}

		switch t.SourceType() {
		case sdktypes.TriggerSourceTypeWebhook:
			fmt.Fprintln(&buf, "      webhook: {}")
			// TODO: More types
		case sdktypes.TriggerSourceTypeSchedule:
			fmt.Fprintf(&buf, "      schedule: %s\n", yamlize(t.Schedule()))
		case sdktypes.TriggerSourceTypeConnection:
			conn, err := ps.DB.GetConnection(ctx, t.ConnectionID())
			if err != nil {
				return nil, err
			}
			fmt.Fprintf(&buf, "      connection: %s\n", yamlize(conn.Name().String()))
		}
	}

	envs, err := ps.DB.ListProjectEnvs(ctx, prj.ID())
	if err != nil {
		return nil, err
	}

	// Collect vars first, print only if there are some
	varsMap := make(map[string]string)
	for _, env := range envs {
		sid, err := sdktypes.Strict(sdktypes.ParseVarScopeID(env.ID().String()))
		if err != nil {
			return nil, err
		}
		vars, err := ps.DB.GetVars(ctx, sid, nil)
		if err != nil {
			return nil, err
		}

		for _, v := range vars {
			if v.IsSecret() {
				continue
			}
			varsMap[v.Name().String()] = v.Value()
		}
	}

	if len(varsMap) > 0 {
		fmt.Fprintln(&buf, "  vars:")
		for n, v := range varsMap {
			fmt.Fprintf(&buf, "    - name: %s\n", n)
			fmt.Fprintf(&buf, "      value: %s\n", yamlize(v))
		}
	}

	return buf.Bytes(), nil
}

func yamlize(v string) string {
	data, _ := yaml.Marshal(v)
	// Trim newline added by yaml.Marshal
	return string(data[:len(data)-1])
}
