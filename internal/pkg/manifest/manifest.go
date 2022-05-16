package manifest

import (
	"fmt"

	multierror "github.com/hashicorp/go-multierror"
)

type Manifest struct {
	Accounts     []Account     `json:"accounts"`
	EventSources []EventSource `json:"eventsrcs"`
	Plugins      []Plugin      `json:"plugins"`
	Projects     []Project     `json:"projects"`
}

func (d Manifest) Compile() (acts []*Action, errs error) {
	add := func(x interface{ Compile() ([]*Action, error) }, what string, args ...interface{}) {
		a, err := x.Compile()
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("%s: %w", fmt.Sprintf(what, args...), err))
			return
		}

		acts = append(acts, a...)
	}

	for i, a := range d.Accounts {
		add(a, "account %d", i)
	}

	for i, s := range d.EventSources {
		add(s, "eventsrc %d", i)
	}

	for i, s := range d.Projects {
		add(s, "project %d", i)
	}

	for i, s := range d.Plugins {
		add(s, "Plugins %d", i)
	}

	return
}
