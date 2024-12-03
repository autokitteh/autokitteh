package systest

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authjwttokens"
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

	orgs := make(map[string]uuid.UUID)

	for _, u := range users {
		uu := sdktypes.NewUser(fmt.Sprintf("%s@%s.org", u.name, u.org)).
			WithDisplayName(u.name).
			WithID(sdktypes.NewTestUserID(u.name))

		consts[strings.ToUpper(u.name+"_uid")] = uu.ID().String()

		seedCommands = append(seedCommands, fmt.Sprintf(
			`insert into users(user_id,email,display_name,created_by) values (%q,%q,%q,%q)`,
			uu.ID().UUIDValue(),
			uu.Email(),
			uu.DisplayName(),
			uu.ID().UUIDValue(),
		))

		oid, ok := orgs[u.org]
		if !ok {
			orgs[u.org] = kittehs.Must1(uuid.NewV7())

			consts[strings.ToUpper(u.org+"_oid")] = orgs[u.org].String()

			seedCommands = append(seedCommands, fmt.Sprintf(
				`insert into orgs(org_id,name,created_by) values (%q,%q,%q)`,
				uuid.New(),
				u.org,
				uu.ID().UUIDValue(),
			))
		}

		seedCommands = append(seedCommands, fmt.Sprintf(
			`insert into org_members(org_id,user_id,created_by) values (%q,%q,%q)`,
			uu.ID().UUIDValue(),
			oid,
			uu.ID().UUIDValue(),
		))

		tokens[u.name] = kittehs.Must1(js.Create(uu))
	}

	tokens["anon"] = ""

	token = tokens[users[0].name]
}

func setUser(name string) error {
	var ok bool
	if token, ok = tokens[name]; !ok {
		return fmt.Errorf("unknown user: %q", name)
	}

	return nil
}
