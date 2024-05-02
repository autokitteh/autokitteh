package sessionsvcs

import (
	"github.com/redis/go-redis/v9"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type Svcs struct {
	fx.In

	DB db.DB

	Envs         sdkservices.Envs
	Deployments  sdkservices.Deployments
	Builds       sdkservices.Builds
	Runtimes     sdkservices.Runtimes
	Integrations sdkservices.Integrations
	Triggers     sdkservices.Triggers
	Connections  sdkservices.Connections
	Vars         sdkservices.Vars

	RedisClient *redis.Client
	Temporal    client.Client
}
