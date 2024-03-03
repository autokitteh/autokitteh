package sdktypes

import (
	"errors"
	"regexp"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

type Symbol struct{ validatedString[symbolTraits] }

type symbolTraits struct{}

var symbolRE = kittehs.Must1(regexp.Compile(`^[a-zA-Z_][\w]*$`))

func (symbolTraits) Validate(s string) error {
	if s != "" && !symbolRE.MatchString(s) {
		return errors.New("invalid")
	}
	return nil
}

func ParseSymbol(s string) (Symbol, error)       { return ParseValidatedString[Symbol](s) }
func StrictParseSymbol(s string) (Symbol, error) { return Strict(ParseSymbol(s)) }

func IsValidSymbol(s string) bool { _, err := StrictParseSymbol(s); return err == nil }

func forceSymbol(s string) Symbol { return forceValidatedString[Symbol](s) }
