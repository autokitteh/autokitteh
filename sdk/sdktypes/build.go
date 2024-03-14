package sdktypes

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	buildsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/builds/v1"
)

type Build struct{ object[*BuildPB, BuildTraits] }

type BuildPB = buildsv1.Build

type BuildTraits struct{}

var InvalidBuild Build

func (BuildTraits) Validate(m *BuildPB) error       { return idField[BuildID]("build_id", m.BuildId) }
func (BuildTraits) StrictValidate(m *BuildPB) error { return nil }

func BuildFromProto(m *BuildPB) (Build, error)       { return FromProto[Build](m) }
func StrictBuildFromProto(m *BuildPB) (Build, error) { return Strict(BuildFromProto(m)) }

func NewBuild() Build { return kittehs.Must1(BuildFromProto(&BuildPB{})) }

func (p Build) ID() (_ BuildID)      { return kittehs.Must1(ParseBuildID(p.read().BuildId)) }
func (p Build) CreatedAt() time.Time { return p.read().CreatedAt.AsTime() }

func (p Build) WithNewID() Build {
	return Build{p.forceUpdate(func(m *BuildPB) { m.BuildId = NewBuildID().String() })}
}

func (p Build) WithCreatedAt(t time.Time) Build {
	return Build{p.forceUpdate(func(m *BuildPB) { m.CreatedAt = timestamppb.New(t) })}
}

func (p Build) OwnerID() OwnerID {
	return kittehs.Must1(ParseOwnerID(p.read().OwnerId))
}
