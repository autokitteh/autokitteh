package sdktypes

const buildIDKind = "bld"

type BuildID = id[buildIDTraits]

type buildIDTraits struct{}

func (buildIDTraits) Prefix() string { return buildIDKind }

func NewBuildID() BuildID                          { return newID[BuildID]() }
func ParseBuildID(s string) (BuildID, error)       { return ParseID[BuildID](s) }
func StrictParseBuildID(s string) (BuildID, error) { return Strict(ParseBuildID(s)) }

var InvalidBuildID BuildID
