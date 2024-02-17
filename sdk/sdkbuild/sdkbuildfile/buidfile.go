package sdkbuildfile

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type BuildFile struct {
	Info     BuildInfo      `json:"info"`
	Runtimes []*RuntimeData `json:"runtimes"`

	// requirements that are not satisfiable at build time.
	RuntimeRequirements []sdktypes.Requirement `json:"runtime_requirements"`
}

type BuildInfo struct {
	Memo map[string]string `json:"memo,omitempty"`
}

type RuntimeData struct {
	Info     RuntimeInfo            `json:"info"`
	Artifact sdktypes.BuildArtifact `json:"artifact"`
}

func (bf *BuildFile) OmitContent() {
	// TODO: replace data with its length. Problem is that JSON renders it as a bytes buffer
	// so it shows it in base64 like `"content": "PDY4IGJ5dGVzPg=="`` which is not useful.
	measure := func(bs []byte) []byte { return nil }

	for _, rt := range bf.Runtimes {
		rt.Artifact = sdktypes.BuildArtifactReplaceCompiledData(rt.Artifact, measure)
	}
}

// This does not overwrite Info.
func (d *RuntimeData) MergeFrom(other *RuntimeData) error {
	var err error

	if other.Artifact != nil {
		if d.Artifact == nil {
			d.Artifact = other.Artifact
		} else if d.Artifact, err = d.Artifact.UpdateError(func(pb *sdktypes.BuildArtifactPB) error {
			otherpb := other.Artifact.ToProto()

			// var overwrites map[string]bool
			pb.CompiledData, _ /* overwrites */ = kittehs.JoinMaps(pb.CompiledData, otherpb.CompiledData)

			/*TODO(ENG-154): Make sure not to enter dups. See also relevant comment in build_resources.go.
			if kittehs.Any(kittehs.MapValues(overwrites)...) {
				return errors.New("compiled data conflict")
			}
			*/

			pb.Exports = append(pb.Exports, otherpb.Exports...)
			pb.Requirements = append(pb.Requirements, otherpb.Requirements...)

			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

type RuntimeInfo struct {
	Name sdktypes.Name `json:"name"`
}

type ResourceInfo struct {
	Path   string `json:"path"`
	Status string `json:"status"`
}
