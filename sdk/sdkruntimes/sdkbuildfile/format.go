package sdkbuildfile

/*
	A build file is a tar.gz archive that contains metadata and artifacts
	for multiple builds from multiple runtimes. This is needed when
	a runtime refers to values from another file that is processed
	by another runtime (for example, starlark code that calls to
	js code, or yaml files).

	Tree structure:

		- version.txt             # build file format version.
		- info.json               # general info, including the root path.
		- runtimes/               # outputs from multiple runtimes.
		  - name/                 # runtime name
		    - info.json           # runtime metadata
			- exports.json        # exports list
			- requirements.json   # requirements list
			- resources.json      # resources list
			- compiled_data/      # runtime build outputs
			                      #  data files written in sorted path order.
*/

const (
	version = "1"
)

var filenames = struct {
	exports, requirements, info, version,
	resourcesDir, resourcesIndex, compiledDir,
	runtimes string
}{
	exports:        "exports.json",
	requirements:   "requirements.json",
	info:           "info.json",
	version:        "version.txt",
	compiledDir:    "compiled",
	resourcesIndex: "resources.json",
	runtimes:       "runtimes",
}
