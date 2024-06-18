package sdktypes

const varIDKind = "var"

type VarID = id[varIDTraits]

type varIDTraits struct{}

func (varIDTraits) Prefix() string { return varIDKind }

func NewVarID() VarID { return newID[VarID]() }

var InvalidVarID VarID
