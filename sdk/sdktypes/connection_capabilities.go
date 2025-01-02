package sdktypes

import (
	connectionsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/connections/v1"
)

type ConnectionCapabilities struct {
	object[*ConnectionCapabilitiesPB, ConnectionCapabilitiesTraits]
}

var InvalidConnectionCapabilities ConnectionCapabilities

type ConnectionCapabilitiesPB = connectionsv1.Capabilities

type ConnectionCapabilitiesTraits struct{ immutableObjectTrait }

func (ConnectionCapabilitiesTraits) Validate(m *ConnectionCapabilitiesPB) error       { return nil }
func (ConnectionCapabilitiesTraits) StrictValidate(m *ConnectionCapabilitiesPB) error { return nil }

func ConnectionCapabilitiesFromProto(m *ConnectionCapabilitiesPB) (ConnectionCapabilities, error) {
	return FromProto[ConnectionCapabilities](m)
}

func (p ConnectionCapabilities) SupportsConnectionTest() bool { return p.read().SupportsConnectionTest }
func (p ConnectionCapabilities) SupportsConnectionInit() bool { return p.read().SupportsConnectionInit }
func (p ConnectionCapabilities) RequiresConnectionInit() bool { return p.read().RequiresConnectionInit }

func (p ConnectionCapabilities) WithSupportsConnectionTest(v bool) ConnectionCapabilities {
	return ConnectionCapabilities{p.forceUpdate(func(m *ConnectionCapabilitiesPB) { m.SupportsConnectionTest = v })}
}

func (p ConnectionCapabilities) WithSupportsConnectionInit(v bool) ConnectionCapabilities {
	return ConnectionCapabilities{p.forceUpdate(func(m *ConnectionCapabilitiesPB) { m.SupportsConnectionInit = v })}
}

func (p ConnectionCapabilities) WithRequiresConnectionInit(v bool) ConnectionCapabilities {
	return ConnectionCapabilities{p.forceUpdate(func(m *ConnectionCapabilitiesPB) { m.RequiresConnectionInit = v })}
}
