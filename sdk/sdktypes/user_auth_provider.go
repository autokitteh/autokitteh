package sdktypes

import (
	"errors"

	userv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/users/v1"
)

type UserAuthProvider struct {
	object[*UserAuthProviderPB, UserAuthProviderTraits]
}

var InvalidUserAuthProvider UserAuthProvider

type UserAuthProviderPB = userv1.UserAuthProvider

type UserAuthProviderTraits struct{}

func (UserAuthProviderTraits) Validate(m *UserAuthProviderPB) error { return nil }

func (UserAuthProviderTraits) StrictValidate(m *UserAuthProviderPB) error {
	return errors.Join(
		mandatory("name", m.Name),
		mandatory("user_id", m.UserId),
		mandatory("email", m.Email),
	)
}

func UserAuthProviderFromProto(m *UserAuthProviderPB) (UserAuthProvider, error) {
	return FromProto[UserAuthProvider](m)
}

func NewUserAuthProvider(name, user_id, email string, data []byte) (UserAuthProvider, error) {
	return UserAuthProviderFromProto(&UserAuthProviderPB{
		Name:   name,
		UserId: user_id,
		Email:  email,
		Data:   data,
	})
}

func (u UserAuthProvider) Name() string   { return u.read().Name }
func (u UserAuthProvider) UserID() string { return u.read().UserId }
func (u UserAuthProvider) Email() string  { return u.read().Email }
func (u UserAuthProvider) Data() []byte   { return u.read().Data }
