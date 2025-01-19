package sdktypes

import (
	"encoding/json"
	"strings"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

type enumTraits interface {
	Prefix() string
	Names() map[int32]string
	Values() map[string]int32
}

func AllEnumNames[T enumTraits]() (values []string) {
	var t T
	for _, s := range t.Names() {
		values = append(values, strings.TrimPrefix(s, t.Prefix()))
	}
	return
}

type enum[T enumTraits, E ~int32] struct{ v E }

func (e enum[T, E]) IsZero() bool { return e.v == 0 }

func (e enum[T, E]) String() string {
	var t T
	return strings.TrimPrefix(t.Names()[int32(e.v)], t.Prefix())
}

func (e enum[T, E]) Prefix() string { var t T; return t.Prefix() }

func (e enum[T, E]) ToProto() E { return e.v }

func (e enum[T, E]) ToInt() int { return int(e.v) }

func (e enum[T, E]) Strict() error {
	if e.v == 0 {
		return sdkerrors.NewInvalidArgumentError("unspecified")
	}

	return nil
}

func (e enum[T, E]) MarshalJSON() ([]byte, error) {
	return []byte(`"` + e.String() + `"`), nil
}

func (e *enum[T, E]) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return
	}

	*e, err = parseEnum[T, E](s)
	return
}

func parseEnum[T enumTraits, E ~int32](raw string) (e enum[T, E], err error) {
	if raw == "" {
		return enum[T, E]{}, nil
	}

	var t T

	upper := strings.ToUpper(raw)
	if !strings.HasPrefix(upper, t.Prefix()) {
		upper = t.Prefix() + upper
	}

	state, ok := t.Values()[upper]
	if !ok {
		err = sdkerrors.NewInvalidArgumentError("unknown state %v", raw)
		return
	}

	e = enum[T, E]{E(state)}
	return
}

func ParseEnum[W ~struct{ enum[T, E] }, T enumTraits, E ~int32](raw string) (w W, err error) {
	var e enum[T, E]
	if e, err = parseEnum[T, E](raw); err != nil {
		return
	}
	w = W{e}
	return
}

func EnumFromProto[W ~struct{ enum[T, E] }, T enumTraits, E ~int32](e E) (w W, err error) {
	var t T
	if _, ok := t.Names()[int32(e)]; ok {
		w = W{enum[T, E]{e}}
		return
	}

	err = sdkerrors.NewInvalidArgumentError("value")
	return
}

func forceEnumFromProto[W ~struct{ enum[T, E] }, T enumTraits, E ~int32](e E) W {
	return W{enum[T, E]{e}}
}
