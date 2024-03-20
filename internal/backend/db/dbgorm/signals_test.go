package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

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
	f := newDBFixture(true)                          // no foreign keys
	findAndAssertCount(t, f, scheme.Signal{}, 0, "") // no signals

	sig := f.newSignal()
	// test createSignal
	f.saveSignalsAndAssert(t, sig)
}

func TestDeleteSignal(t *testing.T) {
	f := newDBFixture(true)                          // no foreign keys
	findAndAssertCount(t, f, scheme.Signal{}, 0, "") // no Signal

	sig := f.newSignal()
	f.saveSignalsAndAssert(t, sig)

	// test deleteSignal
	assert.NoError(t, f.gormdb.RemoveSignal(f.ctx, sig.SignalID))
	f.assertSignalsDeleted(t, sig)
}
