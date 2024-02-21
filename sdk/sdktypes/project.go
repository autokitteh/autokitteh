package sdktypes

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	projectv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/projects/v1"
)

type ProjectPB = projectv1.Project

type Project = *object[*ProjectPB]

var (
	ProjectFromProto       = makeFromProto(validateProject)
	StrictProjectFromProto = makeFromProto(strictValidateProject)
	ToStrictProject        = makeWithValidator(strictValidateProject)
)

func strictValidateProject(pb *projectv1.Project) error {
	if err := ensureNotEmpty(pb.Name, pb.ProjectId); err != nil {
		return err
	}

	return validateProject(pb)
}

func validateProject(pb *projectv1.Project) error {
	if _, err := ParseProjectID(pb.ProjectId); err != nil {
		return fmt.Errorf("project id: %w", err)
	}

	if _, err := ParseName(pb.Name); err != nil {
		return fmt.Errorf("name: %w", err)
	}

	return nil
}

func ProjectHasID(p Project) bool { return p.pb.ProjectId != "" }

func GetProjectID(p Project) ProjectID {
	if p == nil {
		return nil
	}

	return MustParseProjectID(p.pb.ProjectId)
}

func GetProjectName(p Project) Name {
	if p == nil {
		return nil
	}

	return kittehs.Must1(ParseName(p.pb.Name))
}

func GetProjectResourcePaths(p Project) []string {
	if p == nil {
		return nil
	}

	return p.pb.ResourcePaths
}
