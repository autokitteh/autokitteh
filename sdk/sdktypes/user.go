package sdktypes

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"

	userv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/users/v1"
)

type User struct{ object[*UserPB, UserTraits] }

var InvalidUser User

type UserPB = userv1.User

type UserTraits struct{}

func (UserTraits) Validate(m *UserPB) error       { return nil }
func (UserTraits) StrictValidate(m *UserPB) error { return nil }

func UserFromProto(m *UserPB) (User, error) { return FromProto[User](m) }

func (u User) Data() map[string]string { return u.read().Data }
func (u User) Provider() string        { return u.read().Provider }

func NewUser(provider string, data map[string]string) User {
	return kittehs.Must1(UserFromProto(&UserPB{
		Provider: provider,
		Data:     data,
	}))
}

func (u User) UniqueID() (id string) {
	if id = u.Data()["email"]; id == "" {
		id = "<unknown>"
	}
	return
}
