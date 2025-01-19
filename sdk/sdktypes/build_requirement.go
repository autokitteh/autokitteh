package sdktypes

import (
	"errors"
	"net/url"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	runtimesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1"
)

type BuildRequirement struct {
	object[*BuildRequirementPB, BuildRequirementTraits]
}

func init() { registerObject[BuildRequirement]() }

var InvalidBuildRequirement BuildRequirement

type BuildRequirementPB = runtimesv1.Requirement

type BuildRequirementTraits struct{ immutableObjectTrait }

func (BuildRequirementTraits) Validate(m *BuildRequirementPB) error {
	return errors.Join(
		objectField[CodeLocation]("location", m.Location),
		urlField("url", m.Url),
		symbolField("symbol", m.Symbol),
	)
}

func (BuildRequirementTraits) StrictValidate(m *BuildRequirementPB) error {
	return nonzeroMessage(m)
}

func BuildRequirementFromProto(m *BuildRequirementPB) (BuildRequirement, error) {
	return FromProto[BuildRequirement](m)
}

func NewBuildRequirement() BuildRequirement { return zeroObject[BuildRequirement]() }

func (r BuildRequirement) WithLocation(loc CodeLocation) BuildRequirement {
	return BuildRequirement{r.forceUpdate(func(pb *BuildRequirementPB) { pb.Location = loc.ToProto() })}
}

func (r BuildRequirement) WithURL(url *url.URL) BuildRequirement {
	return BuildRequirement{r.forceUpdate(func(pb *BuildRequirementPB) { pb.Url = url.String() })}
}

func (r BuildRequirement) WithSymbol(sym Symbol) BuildRequirement {
	return BuildRequirement{r.forceUpdate(func(pb *BuildRequirementPB) { pb.Symbol = sym.String() })}
}

func (r BuildRequirement) URL() *url.URL { return kittehs.Must1(url.Parse(r.read().Url)) }
