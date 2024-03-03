package sdktypes

const projectIDKind = "project"

type ProjectID = id[projectIDTraits]

type projectIDTraits struct{}

func (projectIDTraits) Prefix() string { return projectIDKind }

func NewProjectID() ProjectID                          { return newID[ProjectID]() }
func ParseProjectID(s string) (ProjectID, error)       { return ParseID[ProjectID](s) }
func StrictParseProjectID(s string) (ProjectID, error) { return Strict(ParseProjectID(s)) }

func IsProjectID(s string) bool { return IsIDOf[projectIDTraits](s) }

var InvalidProjectID ProjectID
