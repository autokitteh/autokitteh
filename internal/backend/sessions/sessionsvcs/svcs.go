package sessionsvcs

import (
	"github.com/redis/go-redis/v9"
	"go.temporal.io/sdk/client"
	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type Svcs struct {
	fx.In

	DB db.DB

	Builds       sdkservices.Builds
	Connections  sdkservices.Connections
	Deployments  sdkservices.Deployments
	Envs         sdkservices.Envs
	Integrations sdkservices.Integrations
	Runtimes     sdkservices.Runtimes
	Triggers     sdkservices.Triggers
	Vars         sdkservices.Vars

	RedisClient *redis.Client
	Temporal    temporalclient.Client
}

func (s *Svcs) TemporalClient() client.Client { return s.Temporal.Temporal() }
