package authusers

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	// SystemUser is a user that is allowed to do anything and is used only for internally invoked operations
	// that are guaranteed to be safe and thus do not require to pass through authorization.
	// This user cannot be an owner of any object.
	// This user cannot login.
	// Tokens cannot be generated for this user.
	SystemUser = kittehs.Must1(sdktypes.UserFromProto(&sdktypes.UserPB{
		UserId:      kittehs.Must1(sdktypes.ParseUserID("usr_3vser000000000000000000000")).String(),
		DisplayName: "System User",
	}))

	// DefaultUser is a user that is used when no user authentication is required but not enabled.
	// This user is a regular user and has no special privileges whatsoever.
	// This user cannot login.
	// Tokens cannot be generated for this user.
	DefaultUser = kittehs.Must1(sdktypes.UserFromProto(&sdktypes.UserPB{
		UserId:       kittehs.Must1(sdktypes.ParseUserID("usr_3vser000000000000000000001")).String(),
		DisplayName:  "Default User",
		DefaultOrgId: DefaultOrg.ToProto().OrgId,
	}))

	DefaultOrg = kittehs.Must1(sdktypes.OrgFromProto(&sdktypes.OrgPB{
		OrgId:       kittehs.Must1(sdktypes.ParseOrgID("org_30rg0000000000000000000002")).String(),
		DisplayName: "Default Org",
	}))
)

func IsInternalUserID(id sdktypes.UserID) bool {
	return id == SystemUser.ID() || id == DefaultUser.ID()
}
