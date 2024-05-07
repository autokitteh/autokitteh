package sdktypes

import (
	"fmt"
	"regexp"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/catnames"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

type Symbol struct{ validatedString[symbolTraits] }

var InvalidSymbol Symbol

type symbolTraits struct{}

var symbolRE = kittehs.Must1(regexp.Compile(`^[a-zA-Z_][\w]*$`))

func (symbolTraits) Validate(s string) error {
	if s != "" && !symbolRE.MatchString(s) {
		return fmt.Errorf(`illegal symbol (expected [a-zA-Z_][\w]*), got: %s`, s)
	}
	return nil
}

func ParseSymbol(s string) (Symbol, error)       { return ParseValidatedString[Symbol](s) }
func StrictParseSymbol(s string) (Symbol, error) { return Strict(ParseSymbol(s)) }

func IsValidSymbol(s string) bool { _, err := StrictParseSymbol(s); return err == nil }

var generateSymbolString = catnames.NewGenerator(intn)

func NewRandomSymbol() Symbol {
	return NewSymbol(
		fmt.Sprintf(
			"%s_%4.4d",
			strings.ReplaceAll(generateSymbolString(), " ", "_"),
			intn(1000),
		),
	)
}

// Helper function to easily create new symbols. Will panic if given string is invalid.
func NewSymbol(s string) Symbol { return forceValidatedString[Symbol](s) }

func NewSymbols(s ...string) []Symbol {
	return kittehs.Transform(s, func(s string) Symbol { return NewSymbol(s) })
}
