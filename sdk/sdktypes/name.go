package sdktypes

import (
	"encoding/json"
	"regexp"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

var nameRE = regexp.MustCompile(`^\w[\w-]*$`)

type Name = *name

type name struct{ h string }

func (h *name) String() string {
	if h == nil {
		return ""
	}

	return h.h
}

func (h *name) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.h)
}

func (h *name) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &h.h)
}

// See comment for IsID for reasoning.
func IsName(s string) bool { return s != "" && !IsID(s) }

func IsValidName(s string) bool { return nameRE.MatchString(s) }

func ParseName(h string) (Name, error) {
	if h == "" {
		return nil, nil
	}

	return StrictParseName(h)
}

func StrictParseName(h string) (Name, error) {
	if !IsValidName(h) {
		return nil, sdkerrors.ErrInvalidArgument
	}

	return &name{h: h}, nil
}
