package sdktypes

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	connectionsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/connections/v1"
)

type ConnectionPB = connectionsv1.Connection

type Connection = *object[*ConnectionPB]

var (
	ConnectionFromProto       = makeFromProto(validateConnection)
	StrictConnectionFromProto = makeFromProto(strictValidateConnection)
	ToStrictConnection        = makeWithValidator(strictValidateConnection)
)

func strictValidateConnection(pb *connectionsv1.Connection) error {
	if err := ensureNotEmpty(pb.IntegrationId, pb.IntegrationToken, pb.Name, pb.ProjectId); err != nil {
		return err
	}
	return validateConnection(pb)
}

func validateConnection(pb *connectionsv1.Connection) error {
	if _, err := ParseConnectionID(pb.ConnectionId); err != nil {
		return err
	}
	if _, err := ParseIntegrationID(pb.IntegrationId); err != nil {
		return err
	}
	if _, err := ParseProjectID(pb.ProjectId); err != nil {
		return err
	}
	if _, err := ParseName(pb.Name); err != nil {
		return err
	}
	return nil
}

func GetConnectionID(c Connection) ConnectionID {
	if c == nil {
		return nil
	}
	return kittehs.Must1(ParseConnectionID(c.pb.ConnectionId))
}

func GetConnectionIntegrationID(c Connection) IntegrationID {
	if c == nil {
		return nil
	}
	return kittehs.Must1(ParseIntegrationID(c.pb.IntegrationId))
}

func GetConnectionIntegrationToken(c Connection) string {
	if c == nil {
		return ""
	}
	return c.pb.IntegrationToken
}

func GetConnectionProjectID(c Connection) ProjectID {
	if c == nil {
		return nil
	}
	return kittehs.Must1(ParseProjectID(c.pb.ProjectId))
}

func GetConnectionName(c Connection) Name {
	if c == nil {
		return nil
	}
	return kittehs.Must1(ParseName(c.pb.Name))
}
