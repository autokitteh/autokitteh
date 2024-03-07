package sdkbuildfile

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type BuildFile struct {
	Info     BuildInfo      `json:"info"`
	Runtimes []*RuntimeData `json:"runtimes"`

	// requirements that are not satisfiable at build time.
	RuntimeRequirements []sdktypes.BuildRequirement `json:"runtime_requirements"`
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
		all := rt.Artifact.CompiledData()
		rt.Artifact = rt.Artifact.WithCompiledData(kittehs.TransformMapValues(all, measure))
	}
}

// This does not overwrite Info.
func (d *RuntimeData) MergeFrom(other *RuntimeData) error {
	d.Artifact = d.Artifact.MergeFrom(other.Artifact)
	return nil
}

type RuntimeInfo struct {
	Name sdktypes.Symbol `json:"name"`
}

type ResourceInfo struct {
	Path   string `json:"path"`
	Status string `json:"status"`
}
