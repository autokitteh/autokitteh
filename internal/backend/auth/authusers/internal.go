package authusers

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func uid(s string) sdktypes.UserID { return kittehs.Must1(sdktypes.ParseUserID(s)) }
func oid(s string) sdktypes.OrgID  { return kittehs.Must1(sdktypes.ParseOrgID(s)) }

var (
	// SystemUser is a user that is allowed to do anything and is used only for internally invoked operations
	// that are guaranteed to be safe and thus do not require to pass through authorization.
	// This user cannot be an owner of any object.
	// This user cannot login.
	// Tokens cannot be generated for this user.
	SystemUser = sdktypes.NewUser().
			WithID(uid("usr_3vser000000000000000000000")).
			WithDisplayName("System User").
			WithStatus(sdktypes.UserStatusActive)

	// DefaultUser is a user that is used when no user authentication is required but not enabled.
	// This user is a regular user and has no special privileges whatsoever.
	// This user cannot login, hence it must not have an email associated with it.
	DefaultUser = sdktypes.NewUser().
			WithID(uid("usr_3vser000000000000000000001")).
			WithDisplayName("Default User").
			WithDefaultOrgID(DefaultOrg.ID()).
			WithStatus(sdktypes.UserStatusActive)

	DefaultOrg = sdktypes.NewOrg().
			WithID(oid("org_30rg0000000000000000000002")).
			WithDisplayName("Default Org")
)

func IsSystemUserID(id sdktypes.UserID) bool { return id == SystemUser.ID() }
