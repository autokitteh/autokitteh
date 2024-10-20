package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"

	userv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/users/v1"
)

type User struct{ object[*UserPB, UserTraits] }

var InvalidUser User

type UserPB = userv1.User

type UserTraits struct{}

func (UserTraits) Validate(m *UserPB) error {
	return errors.Join(
		idField[UserID]("user_id", m.UserId),
	)
}

func (UserTraits) StrictValidate(m *UserPB) error {
	return errors.Join(
		mandatory("user_id", m.UserId),
		mandatory("primary_email", m.PrimaryEmail),
	)
}

func UserFromProto(m *UserPB) (User, error) { return FromProto[User](m) }

func NewUser(id UserID, primaryEmail string) User {
	return kittehs.Must1(UserFromProto(&UserPB{
		UserId:       id.String(),
		PrimaryEmail: primaryEmail,
	}))
}

func (u User) ID() UserID           { return kittehs.Must1(ParseUserID(u.read().UserId)) }
func (u User) PrimaryEmail() Symbol { return kittehs.Must1(ParseSymbol(u.read().PrimaryEmail)) }

func (u User) AuthProviders() []UserAuthProvider {
	return kittehs.Must1(kittehs.TransformError(u.read().AuthProviders, UserAuthProviderFromProto))
}

func (u User) WithAuthProvider(p UserAuthProvider) User {
	return User{u.forceUpdate(func(m *UserPB) { m.AuthProviders = append(m.AuthProviders, p.read()) })}
}

var DefaultUser = kittehs.Must1(UserFromProto(&UserPB{
	UserId:       kittehs.Must1(ParseUserID("usr_3vser000000000000000000001")).String(),
	PrimaryEmail: kittehs.GetenvOr("DEFAULT_USER_EMAIL", "autokitteh@localhost"),
}))
