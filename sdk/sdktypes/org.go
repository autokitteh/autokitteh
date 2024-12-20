package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"

	orgv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/orgs/v1"
)

type Org struct{ object[*OrgPB, OrgTraits] }

var InvalidOrg Org

type OrgPB = orgv1.Org

type OrgTraits struct{}

func (OrgTraits) Validate(m *OrgPB) error {
	return errors.Join(
		idField[OrgID]("org_id", m.OrgId),
		symbolField("name", m.Name),
	)
}

func (OrgTraits) StrictValidate(m *OrgPB) error {
	return errors.Join(
		mandatory("name", m.Name),
	)
}

func OrgFromProto(m *OrgPB) (Org, error) { return FromProto[Org](m) }

func NewOrg(n Symbol) Org {
	return kittehs.Must1(OrgFromProto(&OrgPB{Name: n.String()}))
}

func (u Org) WithID(id OrgID) Org {
	return Org{u.forceUpdate(func(m *OrgPB) { m.OrgId = id.String() })}
}

func (u Org) WithNewID() Org { return u.WithID(NewOrgID()) }

func (u Org) ID() OrgID           { return kittehs.Must1(ParseOrgID(u.read().OrgId)) }
func (u Org) Name() Symbol        { return NewSymbol(u.read().Name) }
func (u Org) DisplayName() string { return u.read().DisplayName }

func (u Org) WithName(n Symbol) Org {
	return Org{u.forceUpdate(func(m *OrgPB) { m.Name = n.String() })}
}

func (u Org) WithDisplayName(n string) Org {
	return Org{u.forceUpdate(func(m *OrgPB) { m.DisplayName = n })}
}
