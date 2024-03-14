package sdktypes

const orgIDKind = "org"

type OrgID = id[orgIDTraits]

type orgIDTraits struct{}

func (orgIDTraits) Prefix() string { return orgIDKind }

func NewOrgID() OrgID                    { return newID[OrgID]() }
func ParseOrgID(s string) (OrgID, error) { return ParseID[OrgID](s) }

func IsOrgID(s string) bool { return IsIDOf[orgIDTraits](s) }

var InvalidOrgID OrgID
