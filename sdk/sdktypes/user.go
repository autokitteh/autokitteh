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
		idField[OrgID]("default_org_id", m.DefaultOrgId),
	)
}

func (UserTraits) StrictValidate(m *UserPB) error { return nil }

func (UserTraits) Mutables() []string {
	// email is not mutable since we do not validate it actually belongs to the user.
	// a user might change their email to someone's else email.
	return []string{"display_name", "disabled", "default_org_id"}
}

func UserFromProto(m *UserPB) (User, error) { return FromProto[User](m) }

func NewUser() User {
	return kittehs.Must1(UserFromProto(&UserPB{}))
}

func (u User) WithID(id UserID) User {
	return User{u.forceUpdate(func(m *UserPB) { m.UserId = id.String() })}
}

func (u User) WithNewID() User { return u.WithID(NewUserID()) }

func (u User) ID() UserID          { return kittehs.Must1(ParseUserID(u.read().UserId)) }
func (u User) Email() string       { return u.read().Email }
func (u User) Disabled() bool      { return u.read().Disabled }
func (u User) DisplayName() string { return u.read().DisplayName }
func (u User) DefaultOrgID() OrgID { return kittehs.Must1(ParseOrgID(u.read().DefaultOrgId)) }

func (u User) WithDisplayName(n string) User {
	return User{u.forceUpdate(func(m *UserPB) { m.DisplayName = n })}
}

func (u User) WithDisabled(b bool) User {
	return User{u.forceUpdate(func(m *UserPB) { m.Disabled = b })}
}

func (u User) WithDefaultOrgID(oid OrgID) User {
	return User{u.forceUpdate(func(m *UserPB) { m.DefaultOrgId = oid.String() })}
}

func (u User) WithEmail(email string) User {
	return User{u.forceUpdate(func(m *UserPB) { m.Email = email })}
}
