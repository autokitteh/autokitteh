package systest

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens/authjwttokens"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	users = []struct {
		name string
		org  string
	}{
		{"zumi", "cats"}, // <-- first user is used by default by the test.
		{"gizmo", "cats"},
		{"shoogy", "dogs"},
		{"bonny", "dogs"},
	}

	seedCommands []string

	tokens = make(map[string]string, len(users))

	token = "INVALID_TOKEN"
)

func init() {
	js := kittehs.Must1(authjwttokens.New(authjwttokens.Configs.Test))

	// org name -> org id.
	orgs := make(map[string]uuid.UUID)

	for _, u := range users {
		uu := sdktypes.NewUser().
			WithEmail(fmt.Sprintf("%s@%s", u.name, u.org)).
			WithDisplayName(u.name).
			WithID(sdktypes.NewTestUserID(u.name))

		consts[strings.ToUpper(u.name+"_uid")] = uu.ID().String()

		personalOrgID := sdktypes.NewTestOrgID(u.name + "org")

		// seed user.
		seedCommands = append(seedCommands, fmt.Sprintf(
			`insert into users(user_id,email,display_name,created_by,default_org_id) values (%q,%q,%q,%q,%q)`,
			uu.ID().UUIDValue(),
			uu.Email(),
			uu.DisplayName(),
			uu.ID().UUIDValue(),
			personalOrgID.UUIDValue(),
		))

		consts[strings.ToUpper(u.name+"_oid")] = personalOrgID.String()

		// seed personal org.
		seedCommands = append(seedCommands, fmt.Sprintf(
			`insert into orgs(org_id,created_by) values (%q,%q)`,
			personalOrgID.UUIDValue(),
			uu.ID().UUIDValue(),
		))

		// add user to personal org.
		seedCommands = append(seedCommands, fmt.Sprintf(
			`insert into org_members(org_id,user_id,status,created_by) values (%q,%q,%d,%q)`,
			personalOrgID.UUIDValue(),
			uu.ID().UUIDValue(),
			sdktypes.OrgMemberStatusActive.ToProto(),
			uu.ID().UUIDValue(),
		))

		oid, ok := orgs[u.org]
		if !ok {
			sdkOrgID := sdktypes.NewTestOrgID(u.org)
			oid = sdkOrgID.UUIDValue()
			orgs[u.org] = oid

			consts[strings.ToUpper(u.org+"_oid")] = sdkOrgID.String()

			// seed shared org.
			seedCommands = append(seedCommands, fmt.Sprintf(
				`insert into orgs(org_id,created_by) values (%q,%q)`,
				oid,
				uu.ID().UUIDValue(),
			))
		}

		// add user to shared org.
		seedCommands = append(seedCommands, fmt.Sprintf(
			`insert into org_members(org_id,user_id,status,created_by) values (%q,%q,%d,%q)`,
			oid,
			uu.ID().UUIDValue(),
			sdktypes.OrgMemberStatusActive.ToProto(),
			uu.ID().UUIDValue(),
		))

		tokens[u.name] = kittehs.Must1(js.Create(uu))
	}

	tokens["anon"] = ""

	setFirstUser()
}

func setFirstUser() { _ = setUser(users[0].name) }

func setUser(name string) error {
	var ok bool
	if token, ok = tokens[name]; !ok {
		return fmt.Errorf("unknown user: %q", name)
	}

	return nil
}
