package langrun

import (
	"fmt"

	"gitlab.com/softkitteh/autokitteh/pkg/idgen"
)

type RunID string

func (r RunID) Child(n int) RunID { return RunID(fmt.Sprintf("%s.%d", r, n)) }
func NewRunID() RunID             { return RunID(idgen.New("R")) }
