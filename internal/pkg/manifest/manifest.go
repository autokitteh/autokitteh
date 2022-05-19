package manifest

import (
	"fmt"

	multierror "github.com/hashicorp/go-multierror"
)

type Manifest struct {
	Accounts     map[string]Account     `json:"accounts"`  // name -> acount
	EventSources map[string]EventSource `json:"eventsrcs"` // id (=account.name) -> eventsrc
	Plugins      map[string]Plugin      `json:"plugins"`   // id (=account.name) -> plugin
	Projects     map[string]Project     `json:"projects"`  // id (=account.unique_id)-> project
}

func (d Manifest) Compile() (acts []*Action, errs error) {
	add := func(x interface {
		Compile(string) ([]*Action, error)
	}, what string, desc string, args ...interface{}) {
		a, err := x.Compile(what)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("%s: %w", fmt.Sprintf(desc, args...), err))
			return
		}

		acts = append(acts, a...)
	}

	for k, v := range d.Accounts {
		add(v, k, "account %q", k)
	}

	for k, v := range d.EventSources {
		add(v, k, "eventsrc %q", k)
	}

	for k, v := range d.Projects {
		add(v, k, "project %q", k)
	}

	for k, v := range d.Plugins {
		add(v, k, "Plugins %q", k)
	}

	return
}
