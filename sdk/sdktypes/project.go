package sdktypes

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	projectv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/projects/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
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
		return fmt.Errorf("%w: missing name | project id", sdkerrors.ErrInvalidArgument)
	}

	return validateProject(pb)
}

func validateProject(pb *projectv1.Project) error {
	if _, err := ParseProjectID(pb.ProjectId); err != nil {
		return err
	}

	if _, err := ParseName(pb.Name); err != nil {
		return err
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
