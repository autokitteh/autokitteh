package dbgorm

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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

func TestSaveSignal(t *testing.T) {
	f := newDBFixture()
	foreignKeys(f.gormdb, false)                   // no foreign keys
	findAndAssertCount[scheme.Signal](t, f, 0, "") // no signals

	sig := f.newSignal()
	// test createSignal
	f.saveSignalsAndAssert(t, sig)
}

func TestSaveSignelForeignKeys(t *testing.T) {
	f := newDBFixture()
	findAndAssertCount[scheme.Signal](t, f, 0, "") // no signals

	// prepare
	sig := f.newSignal()
	conn := f.newConnection()

	sig.ConnectionID = conn.ConnectionID

	f.createConnectionsAndAssert(t, conn)

	// negative test with non-existing assets

	sig.ConnectionID = uuid.New()
	assert.ErrorIs(t, f.gormdb.saveSignal(f.ctx, &sig), gorm.ErrForeignKeyViolated)
	sig.ConnectionID = conn.ConnectionID

	// test with existing assets
	f.saveSignalsAndAssert(t, sig)
}

func TestDeleteSignal(t *testing.T) {
	f := newDBFixture()
	foreignKeys(f.gormdb, false)                   // no foreign keys
	findAndAssertCount[scheme.Signal](t, f, 0, "") // no signals

	sig := f.newSignal()
	f.saveSignalsAndAssert(t, sig)

	// test deleteSignal
	assert.NoError(t, f.gormdb.RemoveSignal(f.ctx, sig.SignalID))
	f.assertSignalsDeleted(t, sig)
}
