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
		mandatory("email", m.Email),
	)
}

func UserFromProto(m *UserPB) (User, error) { return FromProto[User](m) }

func NewUser(email, displayName string) User {
	return kittehs.Must1(UserFromProto(&UserPB{Email: email, DisplayName: displayName}))
}

func (u User) WithID(id UserID) User {
	return User{u.forceUpdate(func(m *UserPB) { m.UserId = id.String() })}
}

func (u User) WithNewID() User { return u.WithID(NewUserID()) }

func (u User) ID() UserID          { return kittehs.Must1(ParseUserID(u.read().UserId)) }
func (u User) Email() string       { return u.read().Email }
func (u User) Disabled() bool      { return u.read().Disabled }
func (u User) DisplayName() string { return u.read().DisplayName }

var DefaultUser = kittehs.Must1(UserFromProto(&UserPB{
	UserId:      kittehs.Must1(ParseUserID("usr_3vser000000000000000000001")).String(),
	Email:       kittehs.GetenvOr("DEFAULT_USER_EMAIL", "autokitteh@localhost"),
	DisplayName: "default",
}))
