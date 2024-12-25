package sdktypes

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

const IntegrationIDKind = "int"

type IntegrationID = id[integrationIDTraits]

var InvalidIntegrationID IntegrationID

type integrationIDTraits struct{}

func (integrationIDTraits) Prefix() string { return IntegrationIDKind }

func NewIntegrationID() IntegrationID                          { return newID[IntegrationID]() }
func ParseIntegrationID(s string) (IntegrationID, error)       { return ParseID[IntegrationID](s) }
func StrictParseIntegrationID(s string) (IntegrationID, error) { return Strict(ParseIntegrationID(s)) }

func IsIntegrationID(s string) bool { return IsIDOf[integrationIDTraits](s) }

func NewIntegrationIDFromName(name string) IntegrationID {
	return kittehs.Must1(ParseIntegrationID(newNamedIDString(name, IntegrationIDKind)))
}
