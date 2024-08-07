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

func preIntegrationTest(t *testing.T) *dbFixture {
	f := newDBFixture()
	findAndAssertCount[scheme.Integration](t, f, 0, "") // no integrations
	return f
}

func TestCreateIntegration(t *testing.T) {
	f := preIntegrationTest(t)

	i := f.newIntegration("test")
	// test createIntegration
	f.createIntegrationsAndAssert(t, i)
}

func TestDeleteIntegration(t *testing.T) {
	f := preIntegrationTest(t)

	i := f.newIntegration("test")
	f.createIntegrationsAndAssert(t, i)

	// test deleteIntegration
	assert.NoError(t, f.gormdb.deleteIntegration(f.ctx, i.IntegrationID))
	f.assertIntegrationsDeleted(t, i)
}
