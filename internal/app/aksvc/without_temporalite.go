//go:build !temporalite

package aksvc

import (
	"errors"

	"github.com/autokitteh/autokitteh/pkg/svc"
)

var TemporaliteComponent = svc.Component{
	Name:     "temporalite",
	Disabled: true,
	Init:     func() error { return errors.New("embedded temporalite is not supported") },
}
