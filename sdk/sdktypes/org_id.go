package sdktypes

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

const OrgIDKind = "org"

type OrgID = id[orgIDTraits]

type orgIDTraits struct{}

func (orgIDTraits) Prefix() string { return OrgIDKind }

func ParseOrgID(s string) (OrgID, error)       { return ParseID[OrgID](s) }
func StrictParseOrgID(s string) (OrgID, error) { return Strict(ParseOrgID(s)) }

var InvalidOrgID OrgID

func NewOrgID() OrgID       { return newID[OrgID]() }
func IsOrgID(s string) bool { return IsIDOf[userIDTraits](s) }

func NewTestOrgID(name string) OrgID {
	return kittehs.Must1(ParseOrgID(newNamedIDString(name, OrgIDKind)))
}
