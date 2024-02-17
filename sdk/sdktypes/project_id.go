package sdktypes

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

const ProjectIDKind = "p"

type ProjectID = *id[projectIDTraits]

var _ ID = (ProjectID)(nil)

type projectIDTraits struct{}

func (projectIDTraits) Kind() string                   { return ProjectIDKind }
func (projectIDTraits) ValidateValue(raw string) error { return validateUUID(raw) }

func ParseProjectID(raw string) (ProjectID, error) { return parseTypedID[projectIDTraits](raw) }

func StrictParseProjectID(raw string) (ProjectID, error) {
	return strictParseTypedID[projectIDTraits](raw)
}

func MustParseProjectID(raw string) ProjectID { return kittehs.Must1(ParseProjectID(raw)) }

func NewProjectID() ProjectID { return newID[projectIDTraits]() }

func ParseProjectIDOrName(raw string) (Name, ProjectID, error) {
	return parseIDOrName[projectIDTraits](raw)
}
