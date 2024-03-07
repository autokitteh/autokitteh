package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
)

func createEnvAndAssert(t *testing.T, f *dbFixture, env scheme.Env) {
	assert.NoError(t, f.gormdb.createEnv(f.ctx, env))
	findAndAssertOne(t, f, env, "env_id = ?", env.EnvID)
}

func TestCreateEnv(t *testing.T) {
	f := newDbFixture(true)                       // no foreign keys
	findAndAssertCount(t, f, scheme.Env{}, 0, "") // no envs

	e := newEnv()
	// test createEnv
	createEnvAndAssert(t, f, e)
}

func TestDeleteEnv(t *testing.T) {
	f := newDbFixture(true)                       // no foreign keys
	findAndAssertCount(t, f, scheme.Env{}, 0, "") // no envs

	e := newEnv()
	createEnvAndAssert(t, f, e)

	// test deleteEnv
	assert.NoError(t, f.gormdb.deleteEnv(f.ctx, e.EnvID))
	findAndAssertCount(t, f, scheme.Env{}, 0, "") // no envs
}
