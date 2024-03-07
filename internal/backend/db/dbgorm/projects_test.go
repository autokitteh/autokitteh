package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
)

func createProjectAndAssert(t *testing.T, f *dbFixture, project scheme.Project) {
	assert.NoError(t, f.gormdb.createProject(f.ctx, project))
	findAndAssertOne(t, f, project, "project_id = ?", project.ProjectID)
}
