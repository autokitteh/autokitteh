package sdktypes

import (
	"fmt"
	"path/filepath"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	runtimesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1"
)

type BuildArtifactPB = runtimesv1.BuildArtifact

type BuildArtifact = *object[*BuildArtifactPB]

var (
	BuildArtifactFromProto       = makeFromProto(validateBuildArtifact)
	StrictBuildArtifactFromProto = makeFromProto(strictValidateBuildArtifact)
	ToStrictBuildArtifact        = makeWithValidator(strictValidateBuildArtifact)

	strictValidateBuildArtifact = validateBuildArtifact
)

func validateBuildArtifact(pb *runtimesv1.BuildArtifact) error {
	if _, err := kittehs.TransformError(pb.Exports, StrictExportFromProto); err != nil {
		return fmt.Errorf("exports: %w", err)
	}

	if _, err := kittehs.TransformError(pb.Requirements, StrictRequirementFromProto); err != nil {
		return fmt.Errorf("requirements: %w", err)
	}

	err := kittehs.ValidateMap(pb.CompiledData, func(k string, _ []byte) error {
		if !filepath.IsLocal(k) {
			return fmt.Errorf("%q is not local", k)
		}

		return nil
	})
	if err != nil {
		return err
	}

	/* TODO(ENG-141,ENG-154): Fail on duplicates. See also relevant comment in buildfile.go.

	if !kittehs.IsUnique(pb.Exports, func(x *ExportPB) string { return x.Symbol }) {
		return fmt.Errorf("exports conflict: %v", pb.Exports)
	}

	*/

	/* TODO(ENG-154): Fail on duplicates. See also relevant comment in buildfile.go.

	if !kittehs.IsUnique(pb.Requirements, func(r *RequirementPB) string {
		return fmt.Sprintf("%s/%s", r.Path, getCodeLocationPBCanonicalString(r.Location))
	}) {
		return errors.New("requirements conflict")
	}
	*/

	return nil
}

func GetBuildArtifactRequirements(p BuildArtifact) []Requirement {
	if p == nil {
		return nil
	}
	return kittehs.Transform(p.pb.Requirements, kittehs.Must11(RequirementFromProto))
}

func GetBuildArtifactExports(p BuildArtifact) []Export {
	if p == nil {
		return nil
	}
	return kittehs.Transform(p.pb.Exports, kittehs.Must11(ExportFromProto))
}

func GetBuildArtifactCompiledData(p BuildArtifact) map[string][]byte {
	if p == nil {
		return nil
	}
	return p.pb.CompiledData
}

func BuildArtifactReplaceCompiledData(p BuildArtifact, f func([]byte) []byte) BuildArtifact {
	return kittehs.Must1(p.Update(func(a *BuildArtifactPB) {
		for k, v := range a.CompiledData {
			a.CompiledData[k] = f(v)
		}
	}))
}
