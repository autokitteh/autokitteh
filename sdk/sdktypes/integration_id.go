package sdktypes

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

const IntegrationIDKind = "integration"

type IntegrationID = *id[integrationIDTraits]

var _ ID = (IntegrationID)(nil)

type integrationIDTraits struct{}

func (integrationIDTraits) Kind() string                   { return IntegrationIDKind }
func (integrationIDTraits) ValidateValue(raw string) error { return validateUUID(raw) }

func ParseIntegrationID(raw string) (IntegrationID, error) {
	return parseTypedID[integrationIDTraits](raw)
}

func StrictParseIntegrationID(raw string) (IntegrationID, error) {
	return strictParseTypedID[integrationIDTraits](raw)
}

func NewIntegrationID() IntegrationID { return newID[integrationIDTraits]() }

func IntegrationIDFromName(name string) IntegrationID {
	txt := fmt.Sprintf("%s:8%031x", IntegrationIDKind, kittehs.HashString64(name))

	return kittehs.Must1(ParseIntegrationID(txt))
}
