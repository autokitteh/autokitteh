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

const integrationIDFromNamePrefix = "3kth"

func NewIntegrationIDFromName(name string) IntegrationID {
	// a hint to the original name is encoded in the ID.
	idName := strings.Map(func(r rune) rune {
		switch r {
		case 'i', 'l':
			return '1'
		case 'o':
			return '0'
		case 'u':
			return 'v'
		default:
			return r
		}
	}, name)

	if len(idName) > 6 {
		idName = idName[:6]
	}
	idName = fmt.Sprintf("%06s", idName)

	txt := fmt.Sprintf("%s_%s%06s%016x", integrationIDKind, integrationIDFromNamePrefix, idName, kittehs.HashString64(name))

	return kittehs.Must1(ParseIntegrationID(txt))
}
