package systest

import (
	"encoding/json"
	"fmt"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens/authjwttokens"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	users = []struct {
		name  string
		org   string
		roles []string
	}{
		{"zumi", "cats", []string{"admin"}}, // <-- first user is used by default by the test.
		{"gizmo", "cats", nil},
		{"shoogy", "dogs", []string{"admin"}},
		{"bonny", "dogs", nil},
	}

	seedCommands []string

	tokens = make(map[string]string, len(users))

	token = "INVALID_TOKEN"
)

func addUser(u sdktypes.User) string {
	return fmt.Sprintf(
		`insert into users(user_id,email,display_name,created_by,default_org_id) values (%q,%q,%q,%q,%q)`,
		u.ID().UUIDValue(),
		u.Email(),
		u.DisplayName(),
		u.ID().UUIDValue(),
		u.DefaultOrgID().UUIDValue(),
	)
}

func addOrg(oid sdktypes.OrgID) string {
	return fmt.Sprintf(`insert into orgs(org_id) values (%q)`, oid.UUIDValue())
}

func addOrgMember(oid sdktypes.OrgID, uid sdktypes.UserID, roles ...string) string {
	if roles == nil {
		roles = []string{}
	}

	return fmt.Sprintf(
		`insert into org_members(org_id,user_id,status,roles) values (%q,%q,%d,'%s')`,
		oid.UUIDValue(),
		uid.UUIDValue(),
		sdktypes.OrgMemberStatusActive.ToProto(),
		kittehs.Must1(json.Marshal(roles)),
	)
}

func init() {
	js := kittehs.Must1(authjwttokens.New(authjwttokens.Configs.Test))

	// org name -> org id.
	orgs := make(map[string]sdktypes.OrgID)

	for _, u := range users {
		uu := sdktypes.NewUser().
			WithEmail(fmt.Sprintf("%s@%s", u.name, u.org)).
			WithDisplayName(u.name).
			WithID(sdktypes.NewTestUserID(u.name))

		consts[strings.ToUpper(u.name+"_uid")] = uu.ID().String()

		personalOrgID := sdktypes.NewTestOrgID(u.name + "org")

		// seed user.
		seedCommands = append(seedCommands, addUser(uu.WithDefaultOrgID(personalOrgID)))

		consts[strings.ToUpper(u.name+"_oid")] = personalOrgID.String()

		// seed personal org.
		seedCommands = append(seedCommands, addOrg(personalOrgID))

		// add user to personal org.
		seedCommands = append(seedCommands, addOrgMember(personalOrgID, uu.ID(), "admin"))

		oid, ok := orgs[u.org]
		if !ok {
			oid = sdktypes.NewTestOrgID(u.org)
			orgs[u.org] = oid

			consts[strings.ToUpper(u.org+"_oid")] = oid.String()

			// seed shared org.
			seedCommands = append(seedCommands, addOrg(oid))
		}

		// add user to shared org.
		seedCommands = append(seedCommands, addOrgMember(oid, uu.ID(), u.roles...))

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
