package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
)

func (f *dbFixture) createIntegrationsAndAssert(t *testing.T, integrations ...scheme.Integration) {
	for _, i := range integrations {
		assert.NoError(t, f.gormdb.createIntegration(f.ctx, &i))
		findAndAssertOne(t, f, i, "integration_id = ?", i.IntegrationID)
	}
}

func (f *dbFixture) assertIntegrationsDeleted(t *testing.T, integrations ...scheme.Integration) {
	for _, integration := range integrations {
		assertDeleted(t, f, integration)
	}
}

func TestCreateIntegration(t *testing.T) {
	f := newDBFixture(true)                               // no foreign keys
	findAndAssertCount(t, f, scheme.Integration{}, 0, "") // no integrations

	i := f.newIntegration()
	// test createIntegration
	f.createIntegrationsAndAssert(t, i)
}

func TestDeleteIntegration(t *testing.T) {
	f := newDBFixture(true)                               // no foreign keys
	findAndAssertCount(t, f, scheme.Integration{}, 0, "") // no integrations

	i := f.newIntegration()
	f.createIntegrationsAndAssert(t, i)

	// test deleteIntegration
	assert.NoError(t, f.gormdb.deleteIntegration(f.ctx, i.IntegrationID))
	f.assertIntegrationsDeleted(t, i)
}
