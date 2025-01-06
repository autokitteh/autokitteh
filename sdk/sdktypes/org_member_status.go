package sdktypes

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	orgsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/orgs/v1"
)

type orgMemberStatusTraits struct{}

var _ enumTraits = orgMemberStatusTraits{}

func (orgMemberStatusTraits) Prefix() string           { return "ORG_MEMBER_STATUS_" }
func (orgMemberStatusTraits) Names() map[int32]string  { return orgsv1.OrgMemberStatus_name }
func (orgMemberStatusTraits) Values() map[string]int32 { return orgsv1.OrgMemberStatus_value }

type OrgMemberStatus struct {
	enum[orgMemberStatusTraits, orgsv1.OrgMemberStatus]
}

type OrgMemberStatusPB = orgsv1.OrgMemberStatus

func orgMemberStatusFromProto(e orgsv1.OrgMemberStatus) OrgMemberStatus {
	return kittehs.Must1(OrgMemberStatusFromProto(e))
}

var (
	PossibleOrgMemberStatusNames = AllEnumNames[orgMemberStatusTraits]()

	OrgMemberStatusUnspecified = orgMemberStatusFromProto(orgsv1.OrgMemberStatus_ORG_MEMBER_STATUS_UNSPECIFIED)
	OrgMemberStatusActive      = orgMemberStatusFromProto(orgsv1.OrgMemberStatus_ORG_MEMBER_STATUS_ACTIVE)
	OrgMemberStatusInvited     = orgMemberStatusFromProto(orgsv1.OrgMemberStatus_ORG_MEMBER_STATUS_INVITED)
	OrgMemberStatusDeclined    = orgMemberStatusFromProto(orgsv1.OrgMemberStatus_ORG_MEMBER_STATUS_DECLINED)
)

func OrgMemberStatusFromProto(e orgsv1.OrgMemberStatus) (OrgMemberStatus, error) {
	return EnumFromProto[OrgMemberStatus](e)
}

func ParseOrgMemberStatus(raw string) (OrgMemberStatus, error) {
	return ParseEnum[OrgMemberStatus](raw)
}
