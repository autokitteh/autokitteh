package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
	"github.com/redis/go-redis/v9"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
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
)

func loadConfig() *Config {
	const prefix = "AKREDIS_"

	k := koanf.New(".")

	// See https://github.com/knadh/koanf#reading-environment-variables.
	kittehs.Must0(k.Load(env.Provider(prefix, ".", func(s string) string {
		return strings.Replace(strings.ToLower(strings.TrimPrefix(s, prefix)), "_", ".", -1)
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

func externalClient(ctx context.Context) (*redis.Client, error) {
	addr := string(sdkmodule.FunctionDataFromContext(ctx))

	if c, ok := clients.Get(addr); ok {
		return c, nil
	}

	opts, err := redis.ParseURL(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid redis url: %w", err)
	}

	// In rare occasions (where multiple clients of the same url are first introduced)
	// this can happen more than once, but that shouldn't be a problem.
	c := redis.NewClient(opts)

	_ = clients.Add(addr, c)

	return c, nil
}
