package systest

import (
	"fmt"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authjwttokens"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	userNames = []string{"zumi", "gizmo", "midnight", "pepurr"}

	users = kittehs.Transform(userNames, func(name string) sdktypes.User {
		return sdktypes.NewUser(name + "@localhost").WithID(sdktypes.NewTestUserID(name)).WithDisplayName(name)
	})

	seedCommand = strings.Join(kittehs.Transform(users, func(u sdktypes.User) string {
		return fmt.Sprintf(
			`insert into users(user_id,email,display_name) values (%q,%q,%q)`,
			u.ID().UUIDValue().String(),
			u.Email(),
			u.DisplayName(),
		)
	}), ";") + ";"

	tokens = make(map[string]string)

	token = "INVALID_TOKEN"
)

func init() {
	js := kittehs.Must1(authjwttokens.New(authjwttokens.Configs.Test))

	for _, u := range users {
		consts[strings.ToUpper(u.DisplayName()+"_uid")] = u.ID().String()

		tokens[u.DisplayName()] = kittehs.Must1(js.Create(u))
	}

	tokens["anon"] = ""

	token = tokens[userNames[0]]
}

func setUser(name string) error {
	var ok bool
	if token, ok = tokens[name]; !ok {
		return fmt.Errorf("unknown user: %q", name)
	}

	return nil
}
