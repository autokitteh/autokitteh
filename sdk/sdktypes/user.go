package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	userv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/users/v1"
)

type User struct {
	object[*UserPB, UserTraits]
}

var InvalidUser User

type UserPB = userv1.User

type UserTraits struct{}

func (UserTraits) Validate(m *UserPB) error {
	return errors.Join(
		nameField("name", m.Name),
		idField[UserID]("user_id", m.UserId),
	)
}

func (UserTraits) StrictValidate(m *UserPB) error {
	return mandatory("name", m.Name)
}

func UserFromProto(m *UserPB) (User, error)       { return FromProto[User](m) }
func StrictUserFromProto(m *UserPB) (User, error) { return Strict(UserFromProto(m)) }

func (p User) ID() UserID   { return kittehs.Must1(ParseUserID(p.read().UserId)) }
func (p User) Name() Symbol { return kittehs.Must1(ParseSymbol(p.read().Name)) }

func NewUser() User {
	return kittehs.Must1(UserFromProto(&UserPB{}))
}

func (p User) WithName(name Symbol) User {
	return User{p.forceUpdate(func(pb *UserPB) { pb.Name = name.String() })}
}

func (p User) WithNewID() User { return p.WithID(NewUserID()) }

func (p User) WithID(id UserID) User {
	return User{p.forceUpdate(func(pb *UserPB) { pb.UserId = id.String() })}
}
