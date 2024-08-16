package sdktypes

const runIDKind = "run"

type RunID = id[runIDTraits]

type runIDTraits struct{}

var InvalidRunID RunID

func (runIDTraits) Prefix() string { return runIDKind }

func NewRunID() RunID                    { return newID[RunID]() }
func ParseRunID(s string) (RunID, error) { return ParseID[RunID](s) }
