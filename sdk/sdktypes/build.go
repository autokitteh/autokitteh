package sdktypes

import (
	"fmt"
	"time"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	buildsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/builds/v1"
)

type BuildPB = buildsv1.Build

type Build = *object[*BuildPB]

var (
	BuildFromProto       = makeFromProto(validateBuild)
	StrictBuildFromProto = makeFromProto(strictValidateBuild)
	ToStrictBuild        = makeWithValidator(strictValidateBuild)
)

func strictValidateBuild(pb *buildsv1.Build) error {
	if err := ensureNotEmpty(pb.BuildId, pb.ProjectId); err != nil {
		return err
	}

	return validateBuild(pb)
}

func validateBuild(pb *buildsv1.Build) error {
	if _, err := ParseBuildID(pb.BuildId); err != nil {
		return fmt.Errorf("build ID: %w", err)
	}

	if _, err := ParseProjectID(pb.ProjectId); err != nil {
		return fmt.Errorf("project ID: %w", err)
	}

	// TODO: validate created_at ?

	return nil
}

func GetBuildID(b Build) BuildID {
	if b == nil {
		return nil
	}
	return kittehs.Must1(ParseBuildID(b.pb.BuildId))
}

func GetBuildProjectID(b Build) ProjectID {
	if b == nil {
		return nil
	}
	return MustParseProjectID(b.pb.ProjectId)
}

func GetBuildCreatedAt(b Build) time.Time {
	if b == nil {
		return time.Time{}
	}

	return b.pb.CreatedAt.AsTime()
}
