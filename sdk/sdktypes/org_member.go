package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	orgv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/orgs/v1"
)

type OrgMember struct {
	object[*OrgMemberPB, OrgMemberTraits]
}

var InvalidOrgMember OrgMember

type OrgMemberPB = orgv1.OrgMember

type OrgMemberTraits struct{}

func (OrgMemberTraits) Mutables() []string { return []string{"status", "roles"} }

func (OrgMemberTraits) Validate(m *OrgMemberPB) error {
	return errors.Join(
		idField[OrgID]("org_id", m.OrgId),
		idField[UserID]("user_id", m.UserId),
		enumField[OrgMemberStatus]("status", m.Status),
	)
}

func (OrgMemberTraits) StrictValidate(m *OrgMemberPB) error {
	return errors.Join(
		mandatory("org_id", m.OrgId),
		mandatory("user_id", m.UserId),
	)
}

func OrgMemberFromProto(m *OrgMemberPB) (OrgMember, error) { return FromProto[OrgMember](m) }

func NewOrgMember(oid OrgID, uid UserID) OrgMember {
	return kittehs.Must1(OrgMemberFromProto(&OrgMemberPB{
		OrgId:  oid.String(),
		UserId: uid.String(),
	}))
}

func (m OrgMember) OrgID() OrgID   { return kittehs.Must1(ParseOrgID(m.read().OrgId)) }
func (m OrgMember) UserID() UserID { return kittehs.Must1(ParseUserID(m.read().UserId)) }
func (m OrgMember) Status() OrgMemberStatus {
	return kittehs.Must1(OrgMemberStatusFromProto(m.read().Status))
}

func (m OrgMember) Roles() []Symbol {
	return kittehs.Must1(kittehs.TransformError(m.read().Roles, StrictParseSymbol))
}

func (m OrgMember) WithRoles(roles ...Symbol) OrgMember {
	return OrgMember{m.forceUpdate(func(pb *OrgMemberPB) { pb.Roles = kittehs.TransformToStrings(roles) })}
}

func (m OrgMember) WithUserID(uid UserID) OrgMember {
	return OrgMember{m.forceUpdate(func(pb *OrgMemberPB) { pb.UserId = uid.String() })}
}

func (m OrgMember) WithStatus(status OrgMemberStatus) OrgMember {
	return OrgMember{m.forceUpdate(func(pb *OrgMemberPB) { pb.Status = status.ToProto() })}
}
