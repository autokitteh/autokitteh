package sessionworkflows

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap/zaptest"
	"gotest.tools/v3/assert"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionsvcs"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type mockDB struct {
	db.DB
	mock.Mock
}

func (m *mockDB) GetSession(_ context.Context, sid sdktypes.SessionID) (sdktypes.Session, error) {
	args := m.Called(sid)
	return args.Get(0).(sdktypes.Session), args.Error(1)
}

func (m *mockDB) AddSessionStopRequest(_ context.Context, sid sdktypes.SessionID, reason string) error {
	args := m.Called(sid, reason)
	return args.Error(0)
}

func (m *mockDB) UpdateSessionState(ctx context.Context, sessionID sdktypes.SessionID, state sdktypes.SessionState) error {
	args := m.Called(sessionID, state)
	return args.Error(0)
}

type fakeTemporalClient struct {
	temporalclient.Client
	t client.Client
}

func (f fakeTemporalClient) Temporal() client.Client { return f.t }

type mockTemporalClient struct {
	client.Client
	mock.Mock
}

func (m *mockTemporalClient) CancelWorkflow(_ context.Context, workflowID, runID string) error {
	args := m.Called(workflowID, runID)
	return args.Error(0)
}

func (m *mockTemporalClient) ExecuteWorkflow(_ context.Context, options client.StartWorkflowOptions, workflow any, wargs ...any) (client.WorkflowRun, error) {
	args := m.Called(options, workflow, wargs)
	return args.Get(0).(client.WorkflowRun), args.Error(1)
}

type mockTemporalWorkflowRun struct {
	client.WorkflowRun
	mock.Mock
}

func (m *mockTemporalWorkflowRun) Get(_ context.Context, valuePtr any) error {
	args := m.Called(valuePtr)
	return args.Error(0)
}

func setup(t *testing.T) (*workflows, *mockDB, *mockTemporalClient) {
	db := mockDB{}
	tc := mockTemporalClient{}

	ws := &workflows{
		l: zaptest.NewLogger(t),
		svcs: &sessionsvcs.Svcs{
			DB:       &db,
			Temporal: fakeTemporalClient{t: &tc},
		},
	}

	return ws, &db, &tc
}

var (
	sid     = sdktypes.NewSessionID()
	session = sdktypes.NewSession(sdktypes.NewBuildID(), kittehs.Must1(sdktypes.ParseCodeLocation("meow:42")), nil, nil).WithID(sid).WithState(sdktypes.SessionStateTypeRunning)
)

func TestStopWorkflowNotFound(t *testing.T) {
	ws, db, _ := setup(t)

	db.On("GetSession", sid).Return(sdktypes.InvalidSession, sdkerrors.ErrNotFound).Once()

	assert.ErrorIs(t, ws.StopWorkflow(context.Background(), sid, "test", false, 0), sdkerrors.ErrNotFound)

	mock.AssertExpectationsForObjects(t, db)
}

func TestStopWorkflowFinal(t *testing.T) {
	ws, db, _ := setup(t)

	db.On("GetSession", sid).Return(session.WithState(sdktypes.SessionStateTypeStopped), nil).Once()

	assert.ErrorIs(t, ws.StopWorkflow(context.Background(), sid, "test", false, 0), sdkerrors.ErrConflict)

	mock.AssertExpectationsForObjects(t, db)
}

func TestStopWorkflowSimple(t *testing.T) {
	ws, db, tc := setup(t)

	db.On("GetSession", sid).Return(session, nil).Once()
	db.On("AddSessionStopRequest", sid, "test").Return(nil).Once()
	tc.On("CancelWorkflow", sid.String(), "").Return(nil).Once()

	assert.NilError(t, ws.StopWorkflow(context.Background(), sid, "test", false, 0))

	mock.AssertExpectationsForObjects(t, db, tc)
}

func TestStopWorkflowLost(t *testing.T) {
	ws, db, tc := setup(t)

	db.On("GetSession", sid).Return(session, nil).Twice()
	db.On("AddSessionStopRequest", sid, "test").Return(nil).Once()
	tc.On("CancelWorkflow", sid.String(), "").Return(&serviceerror.NotFound{}).Once()
	db.On("UpdateSessionState", sid, sdktypes.NewSessionStateError(errors.New("workflow lost"), nil)).Return(nil).Once()

	assert.NilError(t, ws.StopWorkflow(context.Background(), sid, "test", false, 0))

	mock.AssertExpectationsForObjects(t, db, tc)
}

func TestStopWorkflowQuick(t *testing.T) {
	ws, db, tc := setup(t)

	firstGet := db.On("GetSession", sid).Return(session, nil).Once()
	db.On("GetSession", sid).Return(session.WithState(sdktypes.SessionStateTypeStopped), nil).Once().NotBefore(firstGet)
	db.On("AddSessionStopRequest", sid, "test").Return(nil).Once()
	tc.On("CancelWorkflow", sid.String(), "").Return(&serviceerror.NotFound{}).Once()

	assert.NilError(t, ws.StopWorkflow(context.Background(), sid, "test", false, 0))

	mock.AssertExpectationsForObjects(t, db, tc)
}

func TestStopWorkflowForceImmediate(t *testing.T) {
	ws, db, tc := setup(t)

	db.On("GetSession", sid).Return(session, nil).Once()
	db.On("AddSessionStopRequest", sid, "[forced] test").Return(nil).Once()
	tc.On("CancelWorkflow", sid.String(), "").Return(nil).Once()

	opts := &terminateSessionWorkflowParams{
		SessionID: sid,
		Reason:    "test",
	}

	mtwr := &mockTemporalWorkflowRun{}
	tc.On("ExecuteWorkflow", mock.Anything, "delayed_terminate_session", []any{opts}).Return(mtwr, nil).Once()
	mtwr.On("Get", mock.Anything).Return(nil).Once()

	assert.NilError(t, ws.StopWorkflow(context.Background(), sid, "test", true, 0))

	mock.AssertExpectationsForObjects(t, db, tc, mtwr)
}

func TestStopWorkflowForceDelayed(t *testing.T) {
	ws, db, tc := setup(t)

	db.On("GetSession", sid).Return(session, nil).Once()
	db.On("AddSessionStopRequest", sid, "[forced] test").Return(nil).Once()
	tc.On("CancelWorkflow", sid.String(), "").Return(nil).Once()

	opts := &terminateSessionWorkflowParams{
		SessionID: sid,
		Reason:    "test",
		Delay:     100 * time.Millisecond,
	}

	mtwr := &mockTemporalWorkflowRun{}
	tc.On("ExecuteWorkflow", mock.Anything, "delayed_terminate_session", []any{opts}).Return(mtwr, nil).Once()
	mtwr.AssertNotCalled(t, "Get", mock.Anything)

	assert.NilError(t, ws.StopWorkflow(context.Background(), sid, "test", true, 100*time.Millisecond))

	mock.AssertExpectationsForObjects(t, db, tc, mtwr)
}
