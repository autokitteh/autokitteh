//go:build !temporalite

package aksvc

import (
	"errors"

	"gitlab.com/softkitteh/autokitteh/pkg/svc"
)

var TemporaliteComponent = svc.Component{
	Name:     "temporalite",
	Disabled: true,
	Init:     func() error { return errors.New("embedded temporalite is not supported") },
}
