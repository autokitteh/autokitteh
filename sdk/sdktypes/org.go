package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	orgv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/orgs/v1"
)

type Org struct {
	object[*OrgPB, OrgTraits]
}

var InvalidOrg Org

type OrgPB = orgv1.Org

type OrgTraits struct{}

func (OrgTraits) Validate(m *OrgPB) error {
	return errors.Join(
		nameField("name", m.Name),
		idField[OrgID]("org_id", m.OrgId),
	)
}

func (OrgTraits) StrictValidate(m *OrgPB) error {
	return mandatory("name", m.Name)
}

func OrgFromProto(m *OrgPB) (Org, error)       { return FromProto[Org](m) }
func StrictOrgFromProto(m *OrgPB) (Org, error) { return Strict(OrgFromProto(m)) }

func (p Org) ID() OrgID    { return kittehs.Must1(ParseOrgID(p.read().OrgId)) }
func (p Org) Name() Symbol { return kittehs.Must1(ParseSymbol(p.read().Name)) }

func NewOrg() Org {
	return kittehs.Must1(OrgFromProto(&OrgPB{}))
}

func (p Org) WithName(name Symbol) Org {
	return Org{p.forceUpdate(func(pb *OrgPB) { pb.Name = name.String() })}
}

func (p Org) WithNewID() Org { return p.WithID(NewOrgID()) }

func (p Org) WithID(id OrgID) Org {
	return Org{p.forceUpdate(func(pb *OrgPB) { pb.OrgId = id.String() })}
}
