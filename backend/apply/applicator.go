package apply

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

// TODO: change the cli arch so that the applicator could emit cli commands, translated from operations.
//       the cli in turn can invoke operations as a result of commands given.

type Applicator struct {
	Svcs sdkservices.Services
	Path string

	ops  []*Operation
	logs []*Log

	LookupEnv func(string) (string, bool)
}

func (a *Applicator) g() *Getters { return &Getters{Svcs: a.Svcs} }

func (a *Applicator) Operations() []*Operation { return a.ops }
func (a *Applicator) Logs() []*Log             { return a.logs }

func (a *Applicator) Reset() {
	a.ops = nil
	a.ResetLogs()
}

func (a *Applicator) ResetLogs() { a.logs = nil }

func (a *Applicator) log(s string, vs ...any) *Log {
	l := &Log{Msg: fmt.Sprintf(s, vs...), Data: make(map[string]any)}

	a.logs = append(a.logs, l)

	return l
}

func (a *Applicator) op(ops ...*Operation) { a.ops = append(a.ops, ops...) }

func (a *Applicator) LogsAndOps() []string {
	result := make([]string, 0, len(a.logs)+len(a.ops))
	for _, log := range a.logs {
		result = append(result, log.String())
	}

	for _, op := range a.ops {
		result = append(result, op.String())
	}

	return result
}
