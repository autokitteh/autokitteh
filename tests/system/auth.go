package systest

import (
	"fmt"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens/authjwttokens"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	adminRole = sdktypes.NewSymbol("admin")

	users = []struct {
		name  string
		org   string
		roles []sdktypes.Symbol
	}{
		{"zumi", "cats", []sdktypes.Symbol{adminRole}}, // <-- first user is used by default by the test.
		{"gizmo", "cats", nil},
		{"shoogy", "dogs", []sdktypes.Symbol{adminRole}},
		{"bonny", "dogs", nil},
	}

	seedObjects []sdktypes.Object

	tokens = make(map[string]string, len(users))

	token = "INVALID_TOKEN"
)

func init() {
	js := kittehs.Must1(authjwttokens.New(authjwttokens.Configs.Test))

	// org name -> org id.
	orgs := make(map[string]sdktypes.OrgID)

	for _, u := range users {
		uu := sdktypes.NewUser().
			WithEmail(fmt.Sprintf("%s@%s", u.name, u.org)).
			WithDisplayName(u.name).
			WithID(sdktypes.NewTestUserID(u.name)).
			WithStatus(sdktypes.UserStatusActive)

		consts[strings.ToUpper(u.name+"_uid")] = uu.ID().String()

		personalOrgID := sdktypes.NewTestOrgID(u.name + "org")

		// seed user.
		seedObjects = append(seedObjects, uu.WithDefaultOrgID(personalOrgID))

		consts[strings.ToUpper(u.name+"_oid")] = personalOrgID.String()

		// seed personal org.
		seedObjects = append(seedObjects, sdktypes.NewOrg().WithID(personalOrgID).WithName(sdktypes.NewSymbol(u.name+"_org")))

		// add user to personal org.
		seedObjects = append(
			seedObjects,
			sdktypes.NewOrgMember(personalOrgID, uu.ID()).
				WithStatus(sdktypes.OrgMemberStatusActive).
				WithRoles(adminRole),
		)

		oid, ok := orgs[u.org]
		if !ok {
			oid = sdktypes.NewTestOrgID(u.org)
			orgs[u.org] = oid

			consts[strings.ToUpper(u.org+"_oid")] = oid.String()

			// seed shared org.
			seedObjects = append(seedObjects, sdktypes.NewOrg().WithID(oid).WithName(sdktypes.NewSymbol(u.org)))
		}

		// add user to shared org.
		seedObjects = append(
			seedObjects,
			sdktypes.NewOrgMember(oid, uu.ID()).
				WithStatus(sdktypes.OrgMemberStatusActive).
				WithRoles(u.roles...),
		)

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
