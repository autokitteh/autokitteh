package sdktypes

import (
	"errors"
	"maps"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	runtimesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1"
)

type BuildArtifact struct {
	object[*BuildArtifactPB, BuildArtifactTraits]
}

func init() { registerObject[BuildArtifact]() }

var InvalidBuildArtifact BuildArtifact

type BuildArtifactPB = runtimesv1.Artifact

type BuildArtifactTraits struct{ immutableObjectTrait }

func (BuildArtifactTraits) Validate(m *BuildArtifactPB) error {
	return errors.Join(
		objectsSliceField[BuildRequirement]("requirements", m.Requirements),
		objectsSliceField[BuildExport]("exports", m.Exports),
	)
}

func (BuildArtifactTraits) StrictValidate(m *BuildArtifactPB) error { return nil }

func BuildArtifactFromProto(m *BuildArtifactPB) (BuildArtifact, error) {
	return FromProto[BuildArtifact](m)
}

func (a BuildArtifact) CompiledData() map[string][]byte { return a.read().CompiledData }

func (a BuildArtifact) Requirements() []BuildRequirement {
	return kittehs.Transform(a.read().Requirements, forceFromProto[BuildRequirement])
}

func (a BuildArtifact) Exports() []BuildExport {
	return kittehs.Transform(a.read().Exports, forceFromProto[BuildExport])
}

func (r BuildArtifact) WithExports(exports []BuildExport) BuildArtifact {
	return BuildArtifact{r.forceUpdate(func(pb *BuildArtifactPB) { pb.Exports = kittehs.Transform(exports, ToProto) })}
}

func (r BuildArtifact) WithRequirements(reqs []BuildRequirement) BuildArtifact {
	return BuildArtifact{r.forceUpdate(func(pb *BuildArtifactPB) { pb.Requirements = kittehs.Transform(reqs, ToProto) })}
}

func (r BuildArtifact) WithCompiledData(data map[string][]byte) BuildArtifact {
	return BuildArtifact{r.forceUpdate(func(pb *BuildArtifactPB) { pb.CompiledData = data })}
}

func (r BuildArtifact) MergeFrom(other BuildArtifact) BuildArtifact {
	if !r.IsValid() {
		return other
	}

	if !other.IsValid() {
		return r
	}

	/*TODO(ENG-154): Make sure not to enter dups.
	if kittehs.Any(kittehs.MapValues(overwrites)...) {
		return errors.New("compiled data conflict")
	}
	*/

	compiledData := r.CompiledData()
	maps.Copy(compiledData, other.CompiledData())

	return r.
		WithExports(append(r.Exports(), other.Exports()...)).
		WithRequirements(append(r.Requirements(), other.Requirements()...)).
		WithCompiledData(compiledData)
}
