package sdktypes

const runIDKind = "r"

type RunID = id[runIDTraits]

type runIDTraits struct{}

func (runIDTraits) Prefix() string { return runIDKind }

func NewRunID() RunID                    { return newID[RunID]() }
func ParseRunID(s string) (RunID, error) { return ParseID[RunID](s) }
