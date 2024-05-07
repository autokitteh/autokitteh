package store

import (
	"context"

	"github.com/redis/go-redis/v9"

	redisint "go.autokitteh.dev/autokitteh/integrations/redis"
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/store"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const moduleName = "store"

var ExecutorID = sdktypes.NewExecutorID(fixtures.NewBuiltinIntegrationID("store"))

func New(envID sdktypes.EnvID, projectID sdktypes.ProjectID, client *redis.Client) sdkexecutor.Executor {
	mod := redisint.NewInternalModule(
		moduleName,
		ExecutorID,
		client,
		func(s string) string { return store.Prefix(projectID, envID) + s },
	)

	// context.TODO() is supplied here as the redis integration does not require
	// use of context in Configure.
	vs := kittehs.Must1(mod.Configure(context.TODO(), ExecutorID, sdktypes.InvalidConnectionID))

	return sdkexecutor.NewExecutor(mod, ExecutorID, vs)
}
