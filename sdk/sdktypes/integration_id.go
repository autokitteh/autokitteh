package sdktypes

import (
	"fmt"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

const integrationIDKind = "int"

type IntegrationID = id[integrationIDTraits]

var InvalidIntegrationID IntegrationID

type integrationIDTraits struct{}

func (integrationIDTraits) Prefix() string { return integrationIDKind }

func NewIntegrationID() IntegrationID                          { return newID[IntegrationID]() }
func ParseIntegrationID(s string) (IntegrationID, error)       { return ParseID[IntegrationID](s) }
func StrictParseIntegrationID(s string) (IntegrationID, error) { return Strict(ParseIntegrationID(s)) }

func IsIntegrationID(s string) bool { return IsIDOf[integrationIDTraits](s) }

func IntegrationIDFromName(name string) IntegrationID {
	const chars = "0123456789abcdefghjkmnpqrstvwxyz"

	name = strings.Map(func(r rune) rune {
		if strings.ContainsRune(chars, r) {
			return r
		}
		return 'x'
	}, name)

	txt := fmt.Sprintf("%s_3kth%s%s", integrationIDKind, strings.Repeat("0", 26-4-len(name)), name)

	return kittehs.Must1(ParseIntegrationID(txt))
}
