package sdktypes

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"

	userv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/users/v1"
)

type User struct{ object[*UserPB, UserTraits] }

var (
	DefaultUser = NewUser("ak", map[string]string{"name": "dflt", "email": "a@k"})
	InvalidUser User
)

type UserPB = userv1.User

type UserTraits struct{}

func (UserTraits) Validate(m *UserPB) error       { return nil }
func (UserTraits) StrictValidate(m *UserPB) error { return nil }

func UserFromProto(m *UserPB) (User, error) { return FromProto[User](m) }

func (u User) Provider() string        { return u.read().Provider }
func (u User) Data() map[string]string { return u.read().Data }

func NewUser(provider string, data map[string]string) User {
	return kittehs.Must1(UserFromProto(&UserPB{
		Provider: provider,
		Data:     data,
	}))
}

func (u User) Login() string {
	data := u.Data()
	if id := data["email"]; id != "" {
		return id
	}

	if id := data["id"]; id != "" {
		return fmt.Sprintf("%s:%s", u.Provider(), id)
	}

	return ""
}

// Used for display only.
func (u User) Title() (id string) {
	if id = u.Login(); id == "" {
		id = "<unknown>"
	}

	return
}
