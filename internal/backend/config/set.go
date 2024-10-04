package config

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

type ComponentConfig interface {
	Validate() error
}

type emptyComponentConfig struct{}

func (emptyComponentConfig) Validate() error { return nil }

var EmptyComponentConfig emptyComponentConfig

type baseSet interface {
	hasDefault() bool
	default_() ComponentConfig

	hasDev() bool
	dev() ComponentConfig

	hasTest() bool
	test() ComponentConfig
}

type Set[T ComponentConfig] struct {
	Default, Dev, Test *T
}

func (s Set[T]) default_() ComponentConfig { return *s.Default }
func (s Set[T]) dev() ComponentConfig      { return *s.Dev }
func (s Set[T]) test() ComponentConfig     { return *s.Test }

func (s Set[T]) hasDefault() bool { return s.Default != nil }
func (s Set[T]) hasDev() bool     { return s.Dev != nil }
func (s Set[T]) hasTest() bool    { return s.Test != nil }

var EmptySet = Set[emptyComponentConfig]{Default: &EmptyComponentConfig}

func (set *Set[T]) Choose(mode Mode) (chosen T, err error) {
	var zero T

	if set == nil {
		return
	}

	switch mode {
	case "", Default:
		if set.Default == nil {
			return zero, fmt.Errorf("config mode %q: %w", mode, sdkerrors.ErrNotFound)
		}
		chosen = *set.Default
	case Dev:
		if set.Dev == nil {
			return set.Choose(Default)
		}
		chosen = *set.Dev
	case Test:
		if set.Test == nil {
			return set.Choose(Dev)
		}
		chosen = *set.Test
	default:
		return zero, sdkerrors.NewInvalidArgumentError("invalid mode %q", mode)
	}

	if err = chosen.Validate(); err != nil {
		chosen = zero
		err = fmt.Errorf("config for %q: %w", mode, err)
	}

	return
}
