package records

import (
	"errors"
	"strings"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type stateString string

var possibleStates = sdktypes.PossibleEventRecordStates

// Type is only used in help text.
func (s *stateString) Type() string {
	return "state"
}

// String is used both by fmt.Print and by Cobra in help text.
func (s *stateString) String() string {
	return string(*s)
}

// Set must have pointer receiver so it doesn't change the value of a copy.
func (s *stateString) Set(v string) error {
	for _, ps := range possibleStates {
		if strings.EqualFold(v, ps) {
			*s = stateString(v)
			return nil
		}
	}

	return errors.New("must be one of: " + strings.Join(possibleStates, ", "))
}
