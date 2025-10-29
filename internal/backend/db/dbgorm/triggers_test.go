package dbgorm

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (f *dbFixture) createTriggersAndAssert(t *testing.T, triggers ...scheme.Trigger) {
	for _, trigger := range triggers {
		assert.NoError(t, f.gormdb.createTrigger(f.ctx, &trigger))
		findAndAssertOne(t, f, trigger, "trigger_id = ?", trigger.TriggerID)
	}
}

func (f *dbFixture) assertTriggersDeleted(t *testing.T, triggers ...scheme.Trigger) {
	for _, trigger := range triggers {
		assertSoftDeleted(t, f, trigger)
	}
}

func (f *dbFixture) createProjectConnection(t *testing.T) (scheme.Project, scheme.Connection) {
	p := f.newProject()
	c := f.newConnection(p)

	f.createProjectsAndAssert(t, p)
	f.createConnectionsAndAssert(t, c)

	return p, c
}

func preTriggerTest(t *testing.T) *dbFixture {
	f := newDBFixture()
	findAndAssertCount[scheme.Trigger](t, f, 0, "") // no triggers
	return f
}

func TestCreateTrigger(t *testing.T) {
	f := preTriggerTest(t)

	p, c := f.createProjectConnection(t)
	tr := f.newTrigger(p, c)

	// test createTrigger
	f.createTriggersAndAssert(t, tr)
}

func TestGetTrigger(t *testing.T) {
	f := preTriggerTest(t)

	p, c := f.createProjectConnection(t)
	t1 := f.newTrigger(p, c)
	f.createTriggersAndAssert(t, t1)

	// test getTrigger
	t2, err := f.gormdb.getTriggerByID(f.ctx, t1.TriggerID)
	assert.NoError(t, err)
	resetTimes(t2)
	assert.Equal(t, t1, *t2)

	assert.NoError(t, f.gormdb.deleteTrigger(f.ctx, t1.TriggerID))
	_, err = f.gormdb.getTriggerByID(f.ctx, t1.TriggerID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestCreateTriggerForeignKeys(t *testing.T) {
	f := preTriggerTest(t)

	p, c := f.createProjectConnection(t)
	b := f.newBuild(p)
	f.saveBuildsAndAssert(t, b)

	// negative test with non-existing assets

	// zero ProjectID and ConnectionID
	t1 := f.newTrigger()
	assert.Equal(t, t1.ProjectID, uuid.Nil)
	assert.Nil(t, t1.ConnectionID)
	assert.ErrorIs(t, f.gormdb.createTrigger(f.ctx, &t1), gorm.ErrForeignKeyViolated)

	// change connection to builtin scheduler connection (now user owned) - should be allowed

	// use buildID (owned by user to pass user ownership test) to fake unexisting IDs
	t2 := f.newTrigger(p, c)
	t2.ProjectID = b.BuildID // no such projectID, since it's a buildID
	assert.ErrorIs(t, f.gormdb.createTrigger(f.ctx, &t2), gorm.ErrForeignKeyViolated)

	t2.ProjectID = p.ProjectID
	t2.ConnectionID = &b.BuildID // no such connectionID, since it's a buildID
	assert.ErrorIs(t, f.gormdb.createTrigger(f.ctx, &t2), gorm.ErrForeignKeyViolated)
	t2.ConnectionID = &c.ConnectionID

	// test with existing assets
	f.createTriggersAndAssert(t, t2)
}

func TestDeleteTriggerForeignKeys(t *testing.T) {
	f := preTriggerTest(t)
	p, c := f.createProjectConnection(t)

	trg := f.newTrigger(p, c)
	f.createTriggersAndAssert(t, trg)
	evt := f.newEvent(trg, p)
	f.createEventsAndAssert(t, evt)

	// trigger could be deleted, even if it referenced by non-deleted event
	findAndAssertOne(t, f, evt, "trigger_id = ?", trg.TriggerID)    // non-deleted event
	assert.NoError(t, f.gormdb.deleteTrigger(f.ctx, trg.TriggerID)) // deleted trigger
}

func TestListTriggers(t *testing.T) {
	f := preTriggerTest(t)

	p, c := f.createProjectConnection(t)
	t1 := f.newTrigger(p, c)
	f.createTriggersAndAssert(t, t1)

	// test listTriggers
	triggers, err := f.gormdb.listTriggers(f.ctx, sdkservices.ListTriggersFilter{})
	assert.NoError(t, err)
	assert.Len(t, triggers, 1)
	resetTimes(&triggers[0])
	assert.Equal(t, t1, triggers[0])

	// test listTriggers after delete
	assert.NoError(t, f.gormdb.deleteTrigger(f.ctx, t1.TriggerID))
	triggers, err = f.gormdb.listTriggers(f.ctx, sdkservices.ListTriggersFilter{})
	assert.NoError(t, err)
	assert.Len(t, triggers, 0)
}

func TestDeleteTrigger(t *testing.T) {
	f := preTriggerTest(t)

	p, c := f.createProjectConnection(t)
	tr := f.newTrigger(p, c)
	f.createTriggersAndAssert(t, tr)

	// test deleteTrigger
	assert.NoError(t, f.gormdb.deleteTrigger(f.ctx, tr.TriggerID))
	f.assertTriggersDeleted(t, tr)
}

func TestDuplicatedTrigger(t *testing.T) {
	f := preTriggerTest(t)

	p1 := f.newProject()
	p2 := f.newProject()
	c := f.newConnection(p1)
	f.createProjectsAndAssert(t, p1, p2)
	f.createConnectionsAndAssert(t, c)

	// test create two triggers with the same name for the same project and same environment
	tr1 := f.newTrigger(p1, c, "trg")
	f.createTriggersAndAssert(t, tr1)

	tr2 := f.newTrigger(p1, c, "trg")
	assert.ErrorIs(t, f.gormdb.createTrigger(f.ctx, &tr2), gorm.ErrDuplicatedKey)

	// could create the same named trigger for the different project
	tr3 := f.newTrigger(p2, c, "trg")
	f.createTriggersAndAssert(t, tr3)
}

func TestGetTriggerWithActiveDeploymentByID(t *testing.T) {
	f := preTriggerTest(t)

	// test non-existing trigger.
	nonExistingID := sdktypes.NewTriggerID()
	_, _, err := f.gormdb.GetTriggerWithActiveDeploymentByID(f.ctx, nonExistingID.UUIDValue())
	assert.ErrorIs(t, err, sdkerrors.ErrNotFound)

	p, c := f.createProjectConnection(t)
	tr := f.newTrigger(p, c)
	f.createTriggersAndAssert(t, tr)

	// test without active deployment.
	trigger, hasActiveDeployment, err := f.gormdb.GetTriggerWithActiveDeploymentByID(f.ctx, tr.TriggerID)
	assert.NoError(t, err)
	assert.False(t, hasActiveDeployment)
	assert.NotEqual(t, sdktypes.InvalidTrigger, trigger)

	// create active deployment.
	b := f.newBuild(p)
	f.saveBuildsAndAssert(t, b)
	d := f.newDeployment(b, p)
	d.State = int32(sdktypes.DeploymentStateActive.ToProto())
	f.createDeploymentsAndAssert(t, d)

	// test with active deployment.
	trigger, hasActiveDeployment, err = f.gormdb.GetTriggerWithActiveDeploymentByID(f.ctx, tr.TriggerID)
	assert.NoError(t, err)
	assert.True(t, hasActiveDeployment)
	assert.NotEqual(t, sdktypes.InvalidTrigger, trigger)
}

func TestGetTriggerWithActiveDeploymentByWebhookSlug(t *testing.T) {
	f := preTriggerTest(t)

	p, c := f.createProjectConnection(t)
	tr := f.newTrigger(p, c)
	tr.WebhookSlug = "test-webhook"
	f.createTriggersAndAssert(t, tr)

	// test without active deployment
	_, err := f.gormdb.GetTriggerWithActiveDeploymentByWebhookSlug(f.ctx, "test-webhook")
	assert.ErrorIs(t, err, sdkerrors.ErrNotFound)

	// create active deployment
	b := f.newBuild(p)
	f.saveBuildsAndAssert(t, b)
	d := f.newDeployment(b, p)
	d.State = int32(sdktypes.DeploymentStateActive.ToProto())
	f.createDeploymentsAndAssert(t, d)

	// test with active deployment
	trigger, err := f.gormdb.GetTriggerWithActiveDeploymentByWebhookSlug(f.ctx, "test-webhook")
	assert.NoError(t, err)
	assert.NotEqual(t, sdktypes.InvalidTrigger, trigger)
}

func TestCreateTriggerWithTimezone(t *testing.T) {
	f := preTriggerTest(t)

	p, _ := f.createProjectConnection(t)
	tr := f.newTrigger(p)
	tr.SourceType = sdktypes.TriggerSourceTypeSchedule.String()
	tr.Schedule = "0 9 * * *"
	tr.Timezone = "Asia/Jerusalem"

	// Create trigger with timezone.
	f.createTriggersAndAssert(t, tr)

	// Retrieve and verify timezone is stored.
	retrieved, err := f.gormdb.getTriggerByID(f.ctx, tr.TriggerID)
	assert.NoError(t, err)
	assert.Equal(t, "Asia/Jerusalem", retrieved.Timezone)
}

func TestUpdateTriggerTimezone(t *testing.T) {
	f := preTriggerTest(t)

	p, _ := f.createProjectConnection(t)
	tr := f.newTrigger(p)
	tr.SourceType = sdktypes.TriggerSourceTypeSchedule.String()
	tr.Schedule = "0 9 * * *"
	tr.Timezone = "UTC"

	f.createTriggersAndAssert(t, tr)

	// Update timezone.
	tr.Timezone = "Europe/London"
	assert.NoError(t, f.gormdb.updateTrigger(f.ctx, &tr))

	// Verify update.
	retrieved, err := f.gormdb.getTriggerByID(f.ctx, tr.TriggerID)
	assert.NoError(t, err)
	assert.Equal(t, "Europe/London", retrieved.Timezone)
}
