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
	"go.autokitteh.dev/autokitteh/internal/manifest"
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

	if err := ps.DB.CreateProject(ctx, project); err != nil {
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

var origHeader = []byte(`# This is the original autokitteh.yaml specific by the user.
# Look at autokitteh.yaml for current state of the project.

`)

func (ps *Projects) Export(ctx context.Context, projectID sdktypes.ProjectID) ([]byte, error) {
	const manifestFileName = "autokitteh.yaml"
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	rscs, err := ps.DownloadResources(ctx, projectID)
	if err != nil {
		return nil, err
	}

	for name, data := range rscs {
		writeHeader := false
		if name == manifestFileName {
			name = "autokitteh.yaml.user"
			writeHeader = true
		}

		f, err := w.Create(name)
		if err != nil {
			return nil, err
		}

		if writeHeader {
			if _, err := f.Write(origHeader); err != nil {
				return nil, err
			}
		}
		_, err = f.Write(data)
		if err != nil {
			return nil, err
		}
	}

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

	p := manifest.Project{Name: prj.Name().String()}

	cf := sdkservices.ListConnectionsFilter{ProjectID: projectID}
	conns, err := ps.DB.ListConnections(ctx, cf, false)
	if err != nil {
		return nil, err
	}
	for _, c := range conns {
		mc := manifest.Connection{
			Name: c.Name().String(),
		}

		integ, err := ps.Integrations.GetByID(ctx, c.IntegrationID())
		if err != nil {
			return nil, err
		}
		mc.IntegrationKey = integ.UniqueName().String()
		p.Connections = append(p.Connections, &mc)
	}

	tf := sdkservices.ListTriggersFilter{
		ProjectID: projectID,
	}
	triggers, err := ps.DB.ListTriggers(ctx, tf)
	if err != nil {
		return nil, err
	}

	for _, t := range triggers {
		mt := manifest.Trigger{
			Name: t.Name().String(),
			Call: t.CodeLocation().CanonicalString(),
		}
		if filter := t.Filter(); filter != "" {
			mt.Filter = filter
		}
		if etype := t.EventType(); etype != "" {
			mt.EventType = etype
		}

		switch t.SourceType() {
		case sdktypes.TriggerSourceTypeWebhook:
			var wh struct{}
			mt.Webhook = &wh
		case sdktypes.TriggerSourceTypeSchedule:
			sched := t.Schedule()
			mt.Schedule = &sched
		case sdktypes.TriggerSourceTypeConnection:
			conn, found := findConnection(t.ConnectionID(), conns)
			if !found {
				return nil, fmt.Errorf("trigger %s: connection %s not found", t.ID(), t.ConnectionID())
			}
			cname := conn.Name().String()
			mt.ConnectionKey = &cname
		}
		p.Triggers = append(p.Triggers, &mt)
	}

	vars, err := ps.DB.GetVars(ctx, sdktypes.NewVarScopeID(projectID), nil)
	if err != nil {
		return nil, err
	}

	for _, v := range vars {
		if v.IsSecret() {
			continue
		}
		v := manifest.Var{
			Name:  v.Name().String(),
			Value: v.Value(),
		}
		p.Vars = append(p.Vars, &v)
	}

	m := manifest.Manifest{
		Version: manifest.Version,
		Project: &p,
	}

	return yaml.Marshal(m)
}

func findConnection(id sdktypes.ConnectionID, conns []sdktypes.Connection) (sdktypes.Connection, bool) {
	for _, c := range conns {
		if c.ID() == id {
			return c, true
		}
	}

	return sdktypes.Connection{}, false
}

func (ps *Projects) Lint(ctx context.Context, projectID sdktypes.ProjectID, resources map[string][]byte) ([]*sdktypes.CheckViolation, error) {
	data, ok := resources["autokitteh.yaml"]
	if !ok {
		var err error
		data, err = ps.exportManifest(ctx, projectID)
		if err != nil {
			return nil, err
		}
	}

	manifest, err := manifest.Read(data, "autokitteh.yaml")
	if err != nil {
		return nil, err
	}

	violations := Validate(projectID, manifest, resources)
	return violations, nil
}
