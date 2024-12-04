package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	projectv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/projects/v1"
)

type Project struct {
	object[*ProjectPB, ProjectTraits]
}

var InvalidProject Project

type ProjectPB = projectv1.Project

type ProjectTraits struct{}

func (ProjectTraits) Validate(m *ProjectPB) error {
	return errors.Join(
		nameField("name", m.Name),
		idField[ProjectID]("project_id", m.ProjectId),
	)
}

func (ProjectTraits) StrictValidate(m *ProjectPB) error {
	return mandatory("name", m.Name)
}

func ProjectFromProto(m *ProjectPB) (Project, error)       { return FromProto[Project](m) }
func StrictProjectFromProto(m *ProjectPB) (Project, error) { return Strict(ProjectFromProto(m)) }

func (p Project) ID() ProjectID { return kittehs.Must1(ParseProjectID(p.read().ProjectId)) }
func (p Project) Name() Symbol  { return kittehs.Must1(ParseSymbol(p.read().Name)) }

func NewProject() Project {
	return kittehs.Must1(ProjectFromProto(&ProjectPB{}))
}

func (p Project) WithName(name Symbol) Project {
	return Project{p.forceUpdate(func(pb *ProjectPB) { pb.Name = name.String() })}
}

func (p Project) WithNewID() Project { return p.WithID(NewProjectID()) }

func (p Project) WithID(id ProjectID) Project {
	return Project{p.forceUpdate(func(pb *ProjectPB) { pb.ProjectId = id.String() })}
}

type (
	CheckViolation = projectv1.CheckViolation
	ViolationLevel = projectv1.CheckViolation_Level
)

const (
	ViolationError   ViolationLevel = projectv1.CheckViolation_LEVEL_ERROR
	ViolationWarning ViolationLevel = projectv1.CheckViolation_LEVEL_WARNING
)
