package dispatcher

/* TODO
import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/temporalclient"
	"go.autokitteh.dev/sdk/sdkservices"
	"go.autokitteh.dev/sdk/sdktypes"
)

func TestNew(t *testing.T) {
	tc := temporalclient.Client{}
	assert.NotNil(t, New(nil, nil, nil, Services{}, &tc))
}

type mockEvents struct {
	sdkservices.Events
	save           func(context.Context, sdktypes.Event) (sdktypes.EventID, error)
	addEventRecord func(context.Context, sdktypes.EventRecord) error
	get            func(context.Context, sdktypes.EventID) (sdktypes.Event, error)
}

func (e *mockEvents) Save(ctx context.Context, event sdktypes.Event) (sdktypes.EventID, error) {
	return e.save(ctx, event)
}

func (e *mockEvents) AddEventRecord(ctx context.Context, record sdktypes.EventRecord) error {
	return e.addEventRecord(ctx, record)
}

func (e *mockEvents) Get(ctx context.Context, eventID sdktypes.EventID) (sdktypes.Event, error) {
	return e.get(ctx, eventID)
}

type mockTemporal struct {
	client.Client
	executeWorkflow func(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error)
}

func (t *mockTemporal) ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error) {
	return t.executeWorkflow(ctx, options, workflow, args)
}

var l = zap.NewExample()

func makeEvent() sdktypes.Event {
	eid := sdktypes.NewEventID()
	iid := sdktypes.NewIntegrationID()
	it := "token"
	event, _ := sdktypes.EventFromProto(
		&sdktypes.EventPB{
			EventId:          eid.String(),
			IntegrationId:    iid.String(),
			IntegrationToken: it,
			EventType:        "type1",
		})
	return event
}

func TestDispatchErrorOnSaveEvent(t *testing.T) {
	saveEventError := errors.New("save event error")
	events := mockEvents{
		save: func(ctx context.Context, e sdktypes.Event) (sdktypes.EventID, error) {
			return nil, saveEventError
		},
	}

	svc := Services{Events: &events}
	d := New(l, nil, nil, svc, &temporalclient.Client{}).(*dispatcher)

	event := makeEvent()
	_, err := d.Dispatch(context.TODO(), event)

	if err == nil {
		t.Error("should throw an error")
	}

	if !strings.HasSuffix(err.Error(), saveEventError.Error()) {
		t.Error("incorrect error")
	}
}

func TestDispatchErrorStartWorkflow(t *testing.T) {
	// Prepare
	eid := sdktypes.NewEventID()
	events := mockEvents{
		save: func(ctx context.Context, e sdktypes.Event) (sdktypes.EventID, error) {
			return eid, nil
		},
	}

	startWorkflowError := errors.New("start workflow error")
	temporal := mockTemporal{
		executeWorkflow: func(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error) {
			return nil, startWorkflowError
		},
	}
	tc := temporalclient.Client{}
	tc.Client = &temporal

	svc := Services{Events: &events}
	d := New(l, nil, nil, svc, &tc).(*dispatcher)

	event := makeEvent()

	// Execute
	_, err := d.Dispatch(context.TODO(), event)

	// Assert
	if err == nil {
		t.Error("should throw an error")
	}

	if !strings.HasSuffix(err.Error(), startWorkflowError.Error()) {
		t.Error("incorrect error")
	}
}

func TestDispatchStartWorkflowOK(t *testing.T) {
	// Prepare
	eid := sdktypes.NewEventID()
	events := mockEvents{
		save: func(ctx context.Context, e sdktypes.Event) (sdktypes.EventID, error) {
			return eid, nil
		},
	}

	var (
		calledID        string
		calledTaskQueue string
		// calledInput     interface{}
	)

	temporal := mockTemporal{
		executeWorkflow: func(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error) {
			calledID = options.ID
			calledTaskQueue = options.TaskQueue
			// calledInput = args[0]
			return nil, nil
		},
	}
	tc := temporalclient.Client{}
	tc.Client = &temporal

	svc := Services{Events: &events}
	d := New(l, nil, nil, svc, &tc).(*dispatcher)

	event := makeEvent()

	// Execute
	resultEventID, err := d.Dispatch(context.TODO(), event)

	// Assert
	if err != nil {
		t.Error("should not throw error")
	}

	if !strings.EqualFold(resultEventID.String(), eid.String()) {
		t.Error("incorrect event ID returned")
	}

	if !strings.EqualFold(calledID, fmt.Sprintf("%q_%q", workflowName, eid.String())) {
		t.Error("incorrect id for workflow")
	}

	if !strings.EqualFold(calledTaskQueue, calledTaskQueue) {
		t.Error("incorrect task queue")
	}
}
*/
