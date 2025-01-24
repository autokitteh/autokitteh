package dbgorm

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
)

func (f *dbFixture) saveSignalsAndAssert(t *testing.T, signals ...scheme.Signal) {
	for _, signal := range signals {
		assert.NoError(t, f.gormdb.saveSignal(f.ctx, &signal))
		findAndAssertOne(t, f, signal, "signal_id = ?", signal.SignalID)
	}
}

func (f *dbFixture) assertSignalsDeleted(t *testing.T, signals ...scheme.Signal) {
	for _, signal := range signals {
		assertDeleted(t, f, signal)
	}
}

func preSignalTest(t *testing.T) *dbFixture {
	f := newDBFixture()
	findAndAssertCount[scheme.Signal](t, f, 0, "") // no signals
	return f
}

func TestSaveSignal(t *testing.T) {
	f := preSignalTest(t)
	require.NoError(t, foreignKeys(f.gormdb, false)) // no foreign keys

	sig := f.newSignal()
	// test createSignal
	f.saveSignalsAndAssert(t, sig)
}

func TestSaveSignelForeignKeys(t *testing.T) {
	f := preSignalTest(t)

	// prepare
	sig := f.newSignal()
	p := f.newProject()
	conn := f.newConnection(p)

	sig.DestinationID = conn.ConnectionID
	sig.ConnectionID = &conn.ConnectionID

	f.createProjectsAndAssert(t, p)
	f.createConnectionsAndAssert(t, conn)

	// negative test with non-existing assets

	sig.DestinationID = uuid.New()
	sig.ConnectionID = &sig.DestinationID
	assert.ErrorIs(t, f.gormdb.saveSignal(f.ctx, &sig), gorm.ErrForeignKeyViolated)
	sig.ConnectionID = &conn.ConnectionID

	// test with existing assets
	f.saveSignalsAndAssert(t, sig)
}

func TestDeleteSignal(t *testing.T) {
	f := preSignalTest(t)
	require.NoError(t, foreignKeys(f.gormdb, false)) // no foreign keys

	sig := f.newSignal()
	f.saveSignalsAndAssert(t, sig)

	// test deleteSignal
	assert.NoError(t, f.gormdb.RemoveSignal(f.ctx, sig.SignalID))
	f.assertSignalsDeleted(t, sig)
}
