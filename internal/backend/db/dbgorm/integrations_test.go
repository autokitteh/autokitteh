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

func TestCreateIntegration(t *testing.T) {
	f := newDBFixture(true)                               // no foreign keys
	findAndAssertCount(t, f, scheme.Integration{}, 0, "") // no integrations

	i := newIntegration()
	// test createIntegration
	f.createIntegrationsAndAssert(t, i)
}
