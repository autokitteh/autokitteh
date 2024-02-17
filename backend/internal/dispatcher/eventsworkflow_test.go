package dispatcher

/* TODO
import (
	"context"
	"errors"
	"strings"
	"testing"

	"go.autokitteh.dev/autokitteh/internal/temporalclient"
	"go.autokitteh.dev/kittehs"
	"go.autokitteh.dev/sdk/sdkservices"
	"go.autokitteh.dev/sdk/sdktypes"
)

func TestCreateEventRecordSaveRecordErrorPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	events := mockEvents{
		addEventRecord: func(ctx context.Context, er sdktypes.EventRecord) error {
			return errors.New("this will cause panic")
		},
	}

	svc := Services{Events: &events}
	d := New(l, nil, nil, svc, &temporalclient.Client{}).(*dispatcher)

	eid := sdktypes.NewEventID()
	state := sdktypes.EventStateCompleted
	_ = d.createEventRecord(context.TODO(), eid, state)
}

func TestCreateEventRecordParseEventError(t *testing.T) {
	events := mockEvents{}

	svc := Services{Events: &events}
	d := New(l, nil, nil, svc, &temporalclient.Client{}).(*dispatcher)

	state := sdktypes.EventStateCompleted
	err := d.createEventRecord(context.TODO(), nil, state)

	// no asserting to actual error since it's an internal proto error
	if err == nil {
		t.Error("should return an error")
	}
}

type mockConnections struct {
	sdkservices.Connections
	list func(ctx context.Context, filter sdkservices.ListConnectionsFilter) ([]sdktypes.Connection, error)
}

func (m *mockConnections) List(ctx context.Context, filter sdkservices.ListConnectionsFilter) ([]sdktypes.Connection, error) {
	return m.list(ctx, filter)
}

type mockMappings struct {
	sdkservices.Mappings
	list func(ctx context.Context, filter sdkservices.ListMappingsFilter) ([]sdktypes.Mapping, error)
}

func (m *mockMappings) List(ctx context.Context, filter sdkservices.ListMappingsFilter) ([]sdktypes.Mapping, error) {
	return m.list(ctx, filter)
}

type mockDeployments struct {
	sdkservices.Deployments
	list func(ctx context.Context, filter sdkservices.ListDeploymentsFilter) ([]sdktypes.Deployment, error)
}

func (m *mockDeployments) List(ctx context.Context, filter sdkservices.ListDeploymentsFilter) ([]sdktypes.Deployment, error) {
	return m.list(ctx, filter)
}

func TestGetEventSessionDataGetEventError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	events := mockEvents{
		get: func(ctx context.Context, ei sdktypes.EventID) (sdktypes.Event, error) {
			return nil, errors.New("get event error")
		},
	}

	svc := Services{Events: &events}
	d := New(l, nil, nil, svc, &temporalclient.Client{}).(*dispatcher)

	eid := sdktypes.NewEventID()

	_, _ = d.getEventSessionData(context.TODO(), eid)
}

func TestGetEventSessionDataGetConnectionsError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	event := makeEvent()
	events := mockEvents{
		get: func(ctx context.Context, ei sdktypes.EventID) (sdktypes.Event, error) {
			return event, nil
		},
	}

	getConnectionsError := errors.New("get connections error")
	connections := mockConnections{
		list: func(ctx context.Context, filter sdkservices.ListConnectionsFilter) ([]sdktypes.Connection, error) {
			return nil, getConnectionsError
		},
	}

	svc := Services{Events: &events, Connections: &connections}
	d := New(l, nil, nil, svc, &temporalclient.Client{}).(*dispatcher)

	eid := sdktypes.GetEventID(event)

	_, _ = d.getEventSessionData(context.TODO(), eid)
}

func TestGetEventSessionDataZeroConnections(t *testing.T) {
	event := makeEvent()
	events := mockEvents{
		get: func(ctx context.Context, ei sdktypes.EventID) (sdktypes.Event, error) {
			return event, nil
		},
	}

	connections := mockConnections{
		list: func(ctx context.Context, filter sdkservices.ListConnectionsFilter) ([]sdktypes.Connection, error) {
			return []sdktypes.Connection{}, nil
		},
	}

	svc := Services{Events: &events, Connections: &connections}
	d := New(l, nil, nil, svc, &temporalclient.Client{}).(*dispatcher)

	eid := sdktypes.GetEventID(event)

	_, err := d.getEventSessionData(context.TODO(), eid)

	if err != nil {
		t.Error("should not throw an error")
	}
}

func makeConnection(name string) sdktypes.Connection {
	connection, _ := sdktypes.ConnectionFromProto(
		&sdktypes.ConnectionPB{
			ConnectionId:     sdktypes.NewConnectionID().String(),
			IntegrationId:    sdktypes.NewIntegrationID().String(),
			IntegrationToken: "token",
			Name:             name,
			ParentId:         "",
		})
	return connection
}

func TestGetEventSessionDataZeroMappings(t *testing.T) {
	event := makeEvent()
	events := mockEvents{
		get: func(ctx context.Context, ei sdktypes.EventID) (sdktypes.Event, error) {
			return event, nil
		},
	}

	connection := makeConnection("con1")
	connections := mockConnections{
		list: func(ctx context.Context, filter sdkservices.ListConnectionsFilter) ([]sdktypes.Connection, error) {
			return []sdktypes.Connection{connection}, nil
		},
	}

	mappings := mockMappings{
		list: func(ctx context.Context, filter sdkservices.ListMappingsFilter) ([]sdktypes.Mapping, error) {
			return []sdktypes.Mapping{}, nil
		},
	}

	svc := Services{Events: &events, Connections: &connections, Mappings: &mappings}
	d := New(l, nil, nil, svc, &temporalclient.Client{}).(*dispatcher)

	eid := sdktypes.GetEventID(event)

	_, err := d.getEventSessionData(context.TODO(), eid)

	if err != nil {
		t.Error("should not throw an error")
	}
}

func TestGetEventSessionDataMappingsError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	event := makeEvent()
	events := mockEvents{
		get: func(ctx context.Context, ei sdktypes.EventID) (sdktypes.Event, error) {
			return event, nil
		},
	}

	connection := makeConnection("con1")
	connections := mockConnections{
		list: func(ctx context.Context, filter sdkservices.ListConnectionsFilter) ([]sdktypes.Connection, error) {
			return []sdktypes.Connection{connection}, nil
		},
	}

	mappingError := errors.New("mapping error")
	mappings := mockMappings{
		list: func(ctx context.Context, filter sdkservices.ListMappingsFilter) ([]sdktypes.Mapping, error) {
			return nil, mappingError
		},
	}

	svc := Services{Events: &events, Connections: &connections, Mappings: &mappings}
	d := New(l, nil, nil, svc, &temporalclient.Client{}).(*dispatcher)

	eid := sdktypes.GetEventID(event)

	_, _ = d.getEventSessionData(context.TODO(), eid)

}

func makeEntrypoint(eventType string) sdktypes.MappingEntrypoint {
	ep, _ := sdktypes.MappingEntrypointFromProto(
		&sdktypes.MappingEntrypointPB{
			EventType: eventType,
			CodeLocation: &sdktypes.CodeLocationPB{
				Path: "path",
				Row:  0,
				Col:  0,
				Name: "codelocation",
			},
		})
	return ep
}

func makeMapping(entryPoints []sdktypes.MappingEntrypoint) sdktypes.Mapping {
	mapping, err := sdktypes.MappingFromProto(
		&sdktypes.MappingPB{
			MappingId:    sdktypes.NewMappingID().String(),
			EnvId:        sdktypes.NewEnvID().String(),
			ConnectionId: sdktypes.NewConnectionID().String(),
			EntryPoints:  kittehs.Transform(entryPoints, sdktypes.ToProto),
			ModuleName:   "test",
		})
	if err != nil {
		panic(err)
	}
	return mapping
}
func TestGetEventSessionDataMappingNoEntrypoints(t *testing.T) {
	event := makeEvent()
	events := mockEvents{
		get: func(ctx context.Context, ei sdktypes.EventID) (sdktypes.Event, error) {
			return event, nil
		},
	}

	connection := makeConnection("con1")
	connections := mockConnections{
		list: func(ctx context.Context, filter sdkservices.ListConnectionsFilter) ([]sdktypes.Connection, error) {
			return []sdktypes.Connection{connection}, nil
		},
	}

	mapping := makeMapping([]sdktypes.MappingEntrypoint{})
	mappings := mockMappings{
		list: func(ctx context.Context, filter sdkservices.ListMappingsFilter) ([]sdktypes.Mapping, error) {
			return []sdktypes.Mapping{mapping}, nil
		},
	}

	svc := Services{Events: &events, Connections: &connections, Mappings: &mappings}
	d := New(l, nil, nil, svc, &temporalclient.Client{}).(*dispatcher)

	eid := sdktypes.GetEventID(event)

	result, err := d.getEventSessionData(context.TODO(), eid)

	if err != nil {
		t.Error("should not throw an error")
	}

	if result == nil {
		t.Error("should not be nil result")
		return // this reduces lint errors on null check
	}

	if len(result) != 0 {
		t.Error("should be 0 session data")
	}
}

func TestGetEventSessionDataEntrypointOtherType(t *testing.T) {
	event := makeEvent()
	events := mockEvents{
		get: func(ctx context.Context, ei sdktypes.EventID) (sdktypes.Event, error) {
			return event, nil
		},
	}

	connection := makeConnection("con1")
	connections := mockConnections{
		list: func(ctx context.Context, filter sdkservices.ListConnectionsFilter) ([]sdktypes.Connection, error) {
			return []sdktypes.Connection{connection}, nil
		},
	}

	eventType := sdktypes.GetEventType(event)

	ep := makeEntrypoint(eventType + "-other")
	mapping := makeMapping([]sdktypes.MappingEntrypoint{ep})
	mappings := mockMappings{
		list: func(ctx context.Context, filter sdkservices.ListMappingsFilter) ([]sdktypes.Mapping, error) {
			return []sdktypes.Mapping{mapping}, nil
		},
	}

	svc := Services{Events: &events, Connections: &connections, Mappings: &mappings}
	d := New(l, nil, nil, svc, &temporalclient.Client{}).(*dispatcher)

	eid := sdktypes.GetEventID(event)

	result, err := d.getEventSessionData(context.TODO(), eid)

	if err != nil {
		t.Error("should not throw an error")
	}

	if result == nil {
		t.Error("should not be nil result")
		return
	}

	if len(result) != 0 {
		t.Error("should be 0 session data")
	}
}

func TestGetEventSessionDataDeploymentError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	event := makeEvent()
	events := mockEvents{
		get: func(ctx context.Context, ei sdktypes.EventID) (sdktypes.Event, error) {
			return event, nil
		},
	}

	connection := makeConnection("con1")
	connections := mockConnections{
		list: func(ctx context.Context, filter sdkservices.ListConnectionsFilter) ([]sdktypes.Connection, error) {
			return []sdktypes.Connection{connection}, nil
		},
	}

	eventType := sdktypes.GetEventType(event)

	ep := makeEntrypoint(eventType)
	mapping := makeMapping([]sdktypes.MappingEntrypoint{ep})
	mappings := mockMappings{
		list: func(ctx context.Context, filter sdkservices.ListMappingsFilter) ([]sdktypes.Mapping, error) {
			return []sdktypes.Mapping{mapping}, nil
		},
	}

	deplpoymentsError := errors.New("deployment error")
	deployments := mockDeployments{
		list: func(ctx context.Context, filter sdkservices.ListDeploymentsFilter) ([]sdktypes.Deployment, error) {
			return nil, deplpoymentsError
		},
	}

	svc := Services{Events: &events, Connections: &connections, Mappings: &mappings, Deployments: &deployments}
	d := New(l, nil, nil, svc, &temporalclient.Client{}).(*dispatcher)

	eid := sdktypes.GetEventID(event)

	_, _ = d.getEventSessionData(context.TODO(), eid)
}

func makeDeployment() sdktypes.Deployment {
	dp, _ := sdktypes.DeploymentFromProto(&sdktypes.DeploymentPB{
		DeploymentId: sdktypes.NewDeploymentID().String(),
		EnvId:        sdktypes.NewEnvID().String(),
		BuildId:      sdktypes.NewBuildID().String(),
	})
	return dp
}

func TestGetEventSessionDataSessionStart(t *testing.T) {
	event := makeEvent()
	events := mockEvents{
		get: func(ctx context.Context, ei sdktypes.EventID) (sdktypes.Event, error) {
			return event, nil
		},
	}

	connection := makeConnection("con1")
	connections := mockConnections{
		list: func(ctx context.Context, filter sdkservices.ListConnectionsFilter) ([]sdktypes.Connection, error) {
			return []sdktypes.Connection{connection}, nil
		},
	}

	eventType := sdktypes.GetEventType(event)

	ep := makeEntrypoint(eventType)
	mapping := makeMapping([]sdktypes.MappingEntrypoint{ep})
	mappings := mockMappings{
		list: func(ctx context.Context, filter sdkservices.ListMappingsFilter) ([]sdktypes.Mapping, error) {
			return []sdktypes.Mapping{mapping}, nil
		},
	}

	deployment := makeDeployment()
	deployments := mockDeployments{
		list: func(ctx context.Context, filter sdkservices.ListDeploymentsFilter) ([]sdktypes.Deployment, error) {
			return []sdktypes.Deployment{deployment}, nil
		},
	}

	svc := Services{Events: &events, Connections: &connections, Mappings: &mappings, Deployments: &deployments}
	d := New(l, nil, nil, svc, &temporalclient.Client{}).(*dispatcher)

	eid := sdktypes.GetEventID(event)

	result, err := d.getEventSessionData(context.TODO(), eid)

	if err != nil {
		t.Error("should not throw an error")
	}

	if len(result) != 1 {
		t.Error("should return 1 session data")
	}
	sd := result[0]

	if sd.deploymentID.String() != sdktypes.GetDeploymentID(deployment).String() {
		t.Error("session deployment id is wrong")
	}

	providedCodeLocation := sdktypes.GetMappingEntrypointCodeLocation(ep)

	if !strings.EqualFold(
		sdktypes.GetCodeLocationCanonicalString(providedCodeLocation),
		sdktypes.GetCodeLocationCanonicalString(sd.codeLocation),
	) {
		t.Error("invalid code location")
	}
}

type mockSessions struct {
	sdkservices.Sessions
	start func(ctx context.Context, session sdktypes.Session) (sdktypes.SessionID, error)
}

func (m *mockSessions) Start(ctx context.Context, session sdktypes.Session) (sdktypes.SessionID, error) {
	return m.start(ctx, session)
}

// startSessionCounter := 0

func TestStartSessionsEmptyList(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("The code did panic")
		}
	}()
	sessions := mockSessions{
		start: func(ctx context.Context, session sdktypes.Session) (sdktypes.SessionID, error) {
			return nil, errors.New("cause panic")
		},
	}
	svc := Services{Sessions: &sessions}
	d := New(l, nil, nil, svc, &temporalclient.Client{}).(*dispatcher)
	eid := sdktypes.NewEventID()
	d.startSessions(context.Background(), eid, []sessionData{})
}

func TestStartSessionsErrorShouldPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	sessions := mockSessions{
		start: func(ctx context.Context, session sdktypes.Session) (sdktypes.SessionID, error) {
			return nil, errors.New("cause panic")
		},
	}
	svc := Services{Sessions: &sessions}
	d := New(l, nil, nil, svc, &temporalclient.Client{}).(*dispatcher)
	eid := sdktypes.NewEventID()
	sd := sessionData{
		deploymentID: sdktypes.NewDeploymentID(),
		codeLocation: nil,
	}
	d.startSessions(context.Background(), eid, []sessionData{sd})
}

func TestStartSessionsOK(t *testing.T) {
	var startSessionCounter uint
	sessions := mockSessions{
		start: func(ctx context.Context, session sdktypes.Session) (sdktypes.SessionID, error) {
			startSessionCounter += 1
			return nil, nil
		},
	}
	svc := Services{Sessions: &sessions}
	d := New(l, nil, nil, svc, &temporalclient.Client{}).(*dispatcher)
	eid := sdktypes.NewEventID()
	sd := sessionData{
		deploymentID: sdktypes.NewDeploymentID(),
		codeLocation: kittehs.Must1(sdktypes.CodeLocationFromProto(&sdktypes.CodeLocationPB{})),
	}
	d.startSessions(context.Background(), eid, []sessionData{sd})

	if startSessionCounter != 1 {
		t.Error("session counter was not increased")
	}
}
*/
