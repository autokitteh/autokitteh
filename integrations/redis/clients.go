package redis

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
	"github.com/redis/go-redis/v9"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Config struct {
	LRUSize   int           `koanf:"lru_size"`
	LRUExpiry time.Duration `koanf:"lru_expiry"`
}

var (
	defaultConfig = Config{
		LRUSize:   128,
		LRUExpiry: time.Hour,
	}

	clients *expirable.LRU[string, *redis.Client]

	urlVarName = sdktypes.NewSymbol("URL")
)

func loadConfig() *Config {
	const prefix = "AKREDIS_"

	k := koanf.New(".")

	// See https://github.com/knadh/koanf#reading-environment-variables.
	kittehs.Must0(k.Load(env.Provider(prefix, ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(s, prefix)), "_", ".")
	}), nil))

	config := defaultConfig

	kittehs.Must0(k.Unmarshal("", &config))

	return &config
}

func init() {
	config := loadConfig()

	clients = expirable.NewLRU(
		config.LRUSize,
		func(_ string, c *redis.Client) { _ = c.Close() },
		config.LRUExpiry,
	)
}

type Token struct {
	Client *redis.Client

	// A function to manipulate the key. Usually used to add constant prefix to a key.
	KeyFunc func(string) string
}

func (m *module) externalClient(ctx context.Context) (*redis.Client, error) {
	cid, err := sdkmodule.FunctionConnectionIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	vars, err := m.vars.Get(ctx, sdktypes.NewVarScopeID(cid), urlVarName)
	if err != nil {
		return nil, err
	}

	urlVar := vars.Get(urlVarName)
	if !urlVar.IsValid() {
		return nil, errors.New("missing URL var")
	}

	url := urlVar.Value()

	if c, ok := clients.Get(url); ok {
		return c, nil
	}

	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("invalid redis url: %w", err)
	}

	// In rare occasions (where multiple clients of the same url are first introduced)
	// this can happen more than once, but that shouldn't be a problem.
	c := redis.NewClient(opts)

	_ = clients.Add(url, c)

	return c, nil
}
