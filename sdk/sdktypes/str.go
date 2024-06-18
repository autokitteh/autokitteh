package sdktypes

import (
	"encoding/json"

	"google.golang.org/protobuf/types/known/wrapperspb"
	"gopkg.in/yaml.v3"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

// This is a string that is validated by a trait.
// We assume that an empty string is not a valid value.
type validatedString[T validatedStringTraits] struct{ s string }

type validatedStringTraits interface{ Validate(s string) error }

func (s validatedString[T]) IsValid() bool { var zero validatedString[T]; return s != zero }
func (s validatedString[T]) Hash() string {
	if s.s == "" {
		return ""
	}
	return hash(wrapperspb.String(s.s))
}

func parseValidatedString[T validatedStringTraits](v string) (s validatedString[T], err error) {
	var t T
	if err = t.Validate(v); err != nil {
		err = sdkerrors.ErrInvalidArgument{Underlying: err}
		return
	}

	s = validatedString[T]{v}
	return
}

func ParseValidatedString[N ~struct{ validatedString[T] }, T validatedStringTraits](v string) (n N, err error) {
	var s validatedString[T]

	if s, err = parseValidatedString[T](v); err != nil {
		return
	}

	n = N{s}
	return
}

func (s validatedString[T]) String() string { return s.s }

func (s validatedString[T]) GobEncode() ([]byte, error) {
	return []byte(s.s), nil
}

func (s *validatedString[T]) GobDecode(b []byte) (err error) {
	if *s, err = parseValidatedString[T](string(b)); err != nil {
		return err
	}

	return
}

func (s validatedString[T]) MarshalYAML() (any, error) { return s.s, nil }

func (s *validatedString[T]) UnmarshalYAML(v *yaml.Node) (err error) {
	if *s, err = parseValidatedString[T](v.Value); err != nil {
		return err
	}

	return nil
}

func (s validatedString[T]) MarshalJSON() ([]byte, error) { return json.Marshal(s.s) }

func (s *validatedString[T]) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}

	if *s, err = parseValidatedString[T](str); err != nil {
		return err
	}

	return nil
}

func (s validatedString[T]) Strict() error {
	if s.s == "" {
		return sdkerrors.NewInvalidArgumentError("empty")
	}

	return nil
}

func forceValidatedString[W ~struct{ validatedString[T] }, T validatedStringTraits](s string) W {
	return W{validatedString[T]{s}}
}
