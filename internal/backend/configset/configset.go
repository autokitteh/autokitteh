package configset

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

type Mode string

const (
	Default     Mode = "default"
	VolatileDev Mode = "vdev"
	Dev         Mode = "dev"
	Test        Mode = "test"
)

func (m Mode) IsDefault() bool     { return m == "" || m == Default }
func (m Mode) IsVolatileDev() bool { return m == VolatileDev }
func (m Mode) IsDev() bool         { return m == Dev }
func (m Mode) IsTest() bool        { return m == Test }

func ParseMode(s string) (Mode, error) {
	switch s {
	case "", string(Default):
		return Default, nil
	case string(VolatileDev):
		return VolatileDev, nil
	case string(Dev):
		return Dev, nil
	case string(Test):
		return Test, nil
	default:
		return "", sdkerrors.NewInvalidArgumentError("invalid mode %q", s)
	}
}

type Set[T any] struct {
	Default, VolatileDev, Dev, Test *T
}

var Empty = Set[struct{}]{
	Default: &struct{}{},
}

func (set *Set[T]) Choose(mode Mode) (zero T, err error) {
	if set == nil {
		return
	}

	switch mode {
	case "", Default:
		if set.Default == nil {
			return zero, fmt.Errorf("config mode %q: %w", mode, sdkerrors.ErrNotFound)
		}
		return *set.Default, nil
	case VolatileDev:
		if set.VolatileDev == nil {
			return set.Choose(Default)
		}
		return *set.VolatileDev, nil
	case Dev:
		if set.Dev == nil {
			return set.Choose(VolatileDev)
		}
		return *set.Dev, nil
	case Test:
		if set.Test == nil {
			return set.Choose(VolatileDev)
		}
		return *set.Test, nil
	default:
		return zero, sdkerrors.NewInvalidArgumentError("invalid mode %q", mode)
	}
}
