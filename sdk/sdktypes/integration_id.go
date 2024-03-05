package sdktypes

import (
	"fmt"

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
	txt := fmt.Sprintf("%s_%026x", integrationIDKind, kittehs.HashString64(name))

	return kittehs.Must1(ParseIntegrationID(txt))
}
