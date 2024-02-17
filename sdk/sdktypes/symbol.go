package sdktypes

import (
	"encoding/json"
	"regexp"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

var symbolRE = regexp.MustCompile(`^[a-zA-Z_][\w]*$`)

type Symbol = *symbol

type symbol struct{ s string }

func StrictParseSymbol(s string) (Symbol, error) {
	if !symbolRE.MatchString(s) {
		return nil, sdkerrors.ErrInvalidArgument
	}
	return &symbol{s: s}, nil
}

func ParseSymbol(s string) (Symbol, error) {
	if s == "" {
		return nil, nil
	}
	return StrictParseSymbol(s)
}

func (s *symbol) String() string {
	if s == nil {
		return ""
	}
	return s.s
}

func (s *symbol) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}

	return json.Marshal(s.s)
}

func (s *symbol) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}

	s1, err := StrictParseSymbol(text)
	if err != nil {
		return err
	}

	*s = *s1
	return nil
}
