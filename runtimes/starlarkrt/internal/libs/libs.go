package libs

import (
	"github.com/qri-io/starlib/bsoup"
	"github.com/qri-io/starlib/encoding/base64"
	"github.com/qri-io/starlib/encoding/csv"
	"github.com/qri-io/starlib/encoding/json"
	"github.com/qri-io/starlib/encoding/yaml"
	"github.com/qri-io/starlib/geo"
	"github.com/qri-io/starlib/hash"
	"github.com/qri-io/starlib/html"
	"github.com/qri-io/starlib/math"
	"github.com/qri-io/starlib/re"
	"github.com/qri-io/starlib/xlsx"
	"github.com/qri-io/starlib/zipfile"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/starlarktest"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt/internal/libs/parsers"
	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt/internal/libs/pongo2"
	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt/internal/libs/rand"
)

func wrapModule(name string, members starlark.StringDict) starlark.StringDict {
	return starlark.StringDict{
		name: &starlarkstruct.Module{
			Name:    name,
			Members: members,
		},
	}
}

func LoadModules(seed int64) starlark.StringDict {
	modules := map[string]starlark.StringDict{
		pongo2.ModuleName:  wrapModule("pongo2", kittehs.Must1(pongo2.LoadModule())),
		parsers.ModuleName: wrapModule("parsers", kittehs.Must1(parsers.LoadModule())),
		"assert":           wrapModule("assert", kittehs.Must1(starlarktest.LoadAssertModule())),
		"rand":             wrapModule("rand", kittehs.Must1(rand.LoadModule(seed))),

		// starlib
		xlsx.ModuleName:    kittehs.Must1(xlsx.LoadModule()),
		html.ModuleName:    kittehs.Must1(html.LoadModule()),
		bsoup.ModuleName:   kittehs.Must1(bsoup.LoadModule()),
		zipfile.ModuleName: kittehs.Must1(zipfile.LoadModule()),
		re.ModuleName:      kittehs.Must1(re.LoadModule()),
		base64.ModuleName:  kittehs.Must1(base64.LoadModule()),
		csv.ModuleName:     kittehs.Must1(csv.LoadModule()),
		yaml.ModuleName:    kittehs.Must1(yaml.LoadModule()),
		geo.ModuleName:     kittehs.Must1(geo.LoadModule()),
		hash.ModuleName:    kittehs.Must1(hash.LoadModule()),
		math.ModuleName:    wrapModule("math", math.Module.Members),
		json.ModuleName:    wrapModule("json", json.Module.Members),
	}

	symImplMap := make(starlark.StringDict, len(modules))
	for _, module := range modules {
		for sym, v := range module {
			symImplMap[sym] = v
		}
	}
	return symImplMap
}
