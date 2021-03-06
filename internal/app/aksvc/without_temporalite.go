//go:build !temporalite

package aksvc

import (
	"errors"

	"github.com/autokitteh/svc"
)

var TemporaliteComponent = svc.Component{
	Name:     "temporalite",
	Disabled: true,
	Init:     func() error { return errors.New("embedded temporalite is not supported") },
}
