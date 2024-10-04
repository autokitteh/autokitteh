package config

import (
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

type Mode string

const (
	Default Mode = "default"
	Dev     Mode = "dev"
	Test    Mode = "test"
)

func (m Mode) IsDefault() bool { return m == "" || m == Default }
func (m Mode) IsDev() bool     { return m == Dev }
func (m Mode) IsTest() bool    { return m == Test }

func ParseMode(s string) (Mode, error) {
	switch s {
	case "", string(Default):
		return Default, nil
	case string(Dev):
		return Dev, nil
	case string(Test):
		return Test, nil
	default:
		return "", sdkerrors.NewInvalidArgumentError("invalid mode %q", s)
	}
}
