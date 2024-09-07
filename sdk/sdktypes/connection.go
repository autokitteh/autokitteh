package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	connectionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/connections/v1"
)

type Connection struct {
	object[*ConnectionPB, ConnectionTraits]
}

var InvalidConnection Connection

type ConnectionPB = connectionv1.Connection

type ConnectionTraits struct{}

func (ConnectionTraits) Validate(m *ConnectionPB) error {
	return errors.Join(
		nameField("name", m.Name),
		idField[ProjectID]("project_id", m.ProjectId),
		idField[IntegrationID]("integration_id", m.IntegrationId),
		objectField[Status]("status", m.Status),
		objectField[ConnectionCapabilities]("capabilities", m.Capabilities),
	)
}

func (ConnectionTraits) StrictValidate(m *ConnectionPB) error {
	return errors.Join(
		mandatory("name", m.Name),
		mandatory("project_id", m.ProjectId),
		mandatory("integration_id", m.IntegrationId),
	)
}

func ConnectionFromProto(m *ConnectionPB) (Connection, error) { return FromProto[Connection](m) }
func StrictConnectionFromProto(m *ConnectionPB) (Connection, error) {
	return Strict(ConnectionFromProto(m))
}

func (p Connection) ID() ConnectionID { return kittehs.Must1(ParseConnectionID(p.read().ConnectionId)) }
func (p Connection) Name() Symbol     { return kittehs.Must1(ParseSymbol(p.read().Name)) }

func NewConnection(id ConnectionID) Connection {
	return kittehs.Must1(ConnectionFromProto(&ConnectionPB{
		ConnectionId: id.String(),
	}))
}

func (p Connection) WithName(name Symbol) Connection {
	return Connection{p.forceUpdate(func(pb *ConnectionPB) { pb.Name = name.String() })}
}

func (p Connection) WithNewID() Connection { return p.WithID(NewConnectionID()) }

func (p Connection) WithID(id ConnectionID) Connection {
	return Connection{p.forceUpdate(func(pb *ConnectionPB) { pb.ConnectionId = id.String() })}
}

func (p Connection) WithProjectID(id ProjectID) Connection {
	return Connection{p.forceUpdate(func(pb *ConnectionPB) { pb.ProjectId = id.String() })}
}

func (p Connection) WithIntegrationID(id IntegrationID) Connection {
	return Connection{p.forceUpdate(func(pb *ConnectionPB) { pb.IntegrationId = id.String() })}
}

func (p Connection) IntegrationID() IntegrationID {
	return kittehs.Must1(ParseIntegrationID(p.read().IntegrationId))
}

func (p Connection) ProjectID() ProjectID { return kittehs.Must1(ParseProjectID(p.read().ProjectId)) }

func (p Connection) Status() Status {
	return kittehs.Must1(StatusFromProto(p.read().Status))
}

func (p Connection) WithStatus(status Status) Connection {
	return Connection{p.forceUpdate(func(pb *ConnectionPB) { pb.Status = status.ToProto() })}
}

func (p Connection) Capabilities() ConnectionCapabilities {
	return kittehs.Must1(ConnectionCapabilitiesFromProto(p.read().Capabilities))
}

func (p Connection) WithCapabilities(c ConnectionCapabilities) Connection {
	return Connection{p.forceUpdate(func(pb *ConnectionPB) { pb.Capabilities = c.ToProto() })}
}

func (p Connection) WithoutGeneratedFields() Connection {
	return Connection{p.forceUpdate(func(pb *ConnectionPB) {
		pb.Links = nil
		pb.Capabilities = nil
		pb.Status = nil
	})}
}

type Links map[string]string

func (l Links) InitURL() string {
	if l == nil {
		return ""
	}

	return l["init_url"]
}

func (p Connection) Links() Links { return p.read().Links }
func (p Connection) WithLinks(links Links) Connection {
	return Connection{p.forceUpdate(func(pb *ConnectionPB) { pb.Links = links })}
}

func (p Connection) AddLink(name, value string) Connection {
	return Connection{p.forceUpdate(func(pb *ConnectionPB) {
		if pb.Links == nil {
			pb.Links = make(Links)
		}

		pb.Links[name] = value
	})}
}
