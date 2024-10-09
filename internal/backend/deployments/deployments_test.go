package deployments

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	ids = []sdktypes.DeploymentID{
		sdktypes.NewDeploymentID(),
		sdktypes.NewDeploymentID(),
		sdktypes.NewDeploymentID(),
	}

	envID   = sdktypes.NewEnvID()
	buildID = sdktypes.NewBuildID()
)

type testDeployment struct {
	State              sdktypes.DeploymentState
	NumRunningSessions int64
}

type testDB struct {
	db.DB
	deployments map[sdktypes.DeploymentID]*testDeployment
}

func (db *testDB) Transaction(_ context.Context, f func(tx db.DB) error) error { return f(db) }

func (db *testDB) DeploymentHasActiveSessions(_ context.Context, id sdktypes.DeploymentID) (bool, error) {
	return db.deployments[id].NumRunningSessions > 0, nil
}

func (db *testDB) GetDeployment(_ context.Context, id sdktypes.DeploymentID) (sdktypes.Deployment, error) {
	d, ok := db.deployments[id]
	if !ok {
		return sdktypes.InvalidDeployment, sdkerrors.ErrNotFound
	}

	return sdktypes.NewDeployment(id, envID, buildID).WithID(id).WithState(d.State), nil
}

func (db *testDB) ListDeployments(ctx context.Context, filter sdkservices.ListDeploymentsFilter) (deps []sdktypes.Deployment, _ error) {
	for id, v := range db.deployments {
		if !filter.State.IsZero() && filter.State != v.State {
			continue
		}

		if filter.EnvID.IsValid() && filter.EnvID != envID {
			continue
		}

		deps = append(deps, kittehs.Must1(db.GetDeployment(ctx, id)))
	}

	return
}

func (db *testDB) UpdateDeploymentState(_ context.Context, id sdktypes.DeploymentID, state sdktypes.DeploymentState) (oldState sdktypes.DeploymentState, _ error) {
	oldState = db.deployments[id].State
	db.deployments[id].State = state
	return
}

func (db *testDB) ListSessions(_ context.Context, f sdkservices.ListSessionsFilter) (sdkservices.ListSessionResult, error) {
	if !f.CountOnly {
		return sdkservices.ListSessionResult{}, nil
	}

	var n int64

	if f.StateType == sdktypes.SessionStateTypeRunning {
		n = db.deployments[f.DeploymentID].NumRunningSessions
	}

	return sdkservices.ListSessionResult{
		PaginationResult: sdktypes.PaginationResult{
			TotalCount: n,
		},
	}, nil
}

func newTestDeployments(deps map[sdktypes.DeploymentID]*testDeployment) *deployments {
	return &deployments{
		z:  zap.NewNop(),
		db: &testDB{deployments: deps},
	}
}

func TestActivateSimple(t *testing.T) {
	deps := newTestDeployments(map[sdktypes.DeploymentID]*testDeployment{
		ids[0]: {State: sdktypes.DeploymentStateInactive},
	})

	if assert.NoError(t, deps.Activate(context.Background(), ids[0])) {
		d, err := deps.Get(context.Background(), ids[0])
		if assert.NoError(t, err) {
			assert.Equal(t, sdktypes.DeploymentStateActive, d.State())
		}
	}
}

func TestActivateAndDeactivateOthers(t *testing.T) {
	deps := newTestDeployments(map[sdktypes.DeploymentID]*testDeployment{
		ids[0]: {State: sdktypes.DeploymentStateInactive},
		ids[1]: {State: sdktypes.DeploymentStateActive},
	})

	if assert.NoError(t, deps.Activate(context.Background(), ids[0])) {
		d, err := deps.Get(context.Background(), ids[0])
		if assert.NoError(t, err) {
			assert.Equal(t, sdktypes.DeploymentStateActive, d.State())
		}

		d, err = deps.Get(context.Background(), ids[1])
		if assert.NoError(t, err) {
			assert.Equal(t, sdktypes.DeploymentStateInactive, d.State())
		}
	}
}

func TestActivateAndDrainOthers(t *testing.T) {
	deps := newTestDeployments(map[sdktypes.DeploymentID]*testDeployment{
		ids[0]: {State: sdktypes.DeploymentStateInactive},
		ids[1]: {State: sdktypes.DeploymentStateActive},
		ids[2]: {State: sdktypes.DeploymentStateActive, NumRunningSessions: 1},
	})

	if assert.NoError(t, deps.Activate(context.Background(), ids[0])) {
		d, err := deps.Get(context.Background(), ids[0])
		if assert.NoError(t, err) {
			assert.Equal(t, sdktypes.DeploymentStateActive, d.State())
		}

		d, err = deps.Get(context.Background(), ids[1])
		if assert.NoError(t, err) {
			assert.Equal(t, sdktypes.DeploymentStateInactive, d.State())
		}

		d, err = deps.Get(context.Background(), ids[2])
		if assert.NoError(t, err) {
			assert.Equal(t, sdktypes.DeploymentStateDraining, d.State())
		}
	}
}

func TestDeactivateSimple(t *testing.T) {
	deps := newTestDeployments(map[sdktypes.DeploymentID]*testDeployment{
		ids[0]: {State: sdktypes.DeploymentStateActive},
	})

	if assert.NoError(t, deps.Deactivate(context.Background(), ids[0])) {
		d, err := deps.Get(context.Background(), ids[0])
		if assert.NoError(t, err) {
			assert.Equal(t, sdktypes.DeploymentStateInactive, d.State())
		}
	}
}

func TestDeactivateDrain(t *testing.T) {
	deps := newTestDeployments(map[sdktypes.DeploymentID]*testDeployment{
		ids[0]: {State: sdktypes.DeploymentStateActive, NumRunningSessions: 1},
	})

	if assert.NoError(t, deps.Deactivate(context.Background(), ids[0])) {
		d, err := deps.Get(context.Background(), ids[0])
		if assert.NoError(t, err) {
			assert.Equal(t, sdktypes.DeploymentStateDraining, d.State())
		}
	}
}
