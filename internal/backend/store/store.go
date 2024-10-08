package store

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/config"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const miniredisURL = "miniredis"

type Config struct {
	ServerURL string `koanf:"server_url"`
}

func (c Config) Validate() error {
	if c.ServerURL == "" || c.ServerURL == miniredisURL {
		return nil
	}

	_, err := url.Parse(c.ServerURL)
	return err
}

var Configs = config.Set[Config]{
	Default: &Config{},
	Dev: &Config{
		ServerURL: miniredisURL,
	},
}

type store struct {
	z *zap.Logger

	client *redis.Client
}

func New(z *zap.Logger, cfg *Config) (sdkservices.Store, *redis.Client, error) {
	var client *redis.Client

	if cfg.ServerURL == "" {
		z.Warn("redis not configured, will not be available")
		return nil, nil, nil
	}

	if cfg.ServerURL == miniredisURL {
		mr, err := miniredis.Run()
		if err != nil {
			return nil, nil, fmt.Errorf("miniredis: %w", err)
		}

		addr := mr.Addr()

		client = redis.NewClient(&redis.Options{Addr: addr})
	} else {
		opts, err := redis.ParseURL(cfg.ServerURL)
		if err != nil {
			return nil, nil, err
		}

		client = redis.NewClient(opts)
	}

	return &store{z: z, client: client}, client, nil
}

func Prefix(projectID sdktypes.ProjectID, envID sdktypes.EnvID) string {
	return fmt.Sprintf("%v:%v:", projectID.Value(), envID.Value())
}

func (s *store) List(ctx context.Context, envID sdktypes.EnvID, projectID sdktypes.ProjectID) ([]string, error) {
	if s.client == nil {
		// return nil in order not to fail any query that doesn't check if redis enabled or not.
		return nil, nil
	}

	var (
		prefix = Prefix(projectID, envID)
		cursor uint64
		ks     []string
	)

	for {
		chunk, cursor, err := s.client.Scan(ctx, cursor, "*", 1000).Result()
		if err != nil {
			return nil, err
		}

		if cursor == 0 {
			break
		}

		ks = append(ks, chunk...)
	}

	ks = kittehs.Transform(ks, func(s string) string { return strings.TrimPrefix(s, prefix) })

	sort.Strings(ks)

	return ks, nil
}

func (s *store) Get(ctx context.Context, envID sdktypes.EnvID, projectID sdktypes.ProjectID, keys []string) (map[string]sdktypes.Value, error) {
	if s.client == nil {
		return nil, sdkerrors.ErrNotImplemented
	}

	prefix := Prefix(projectID, envID)

	vs, err := s.client.MGet(ctx, kittehs.Transform(keys, func(k string) string { return prefix + k })...).Result()
	if err != nil {
		return nil, err
	}

	if len(vs) != len(keys) {
		return nil, fmt.Errorf("number of returned values %d != number of keys %d", len(keys), len(vs))
	}

	m := make(map[string]sdktypes.Value, len(vs))
	for i, k := range keys {
		if m[k], err = sdktypes.WrapValue(vs[i]); err != nil {
			return nil, fmt.Errorf("wrap #%d: %w", i, err)
		}
	}

	return m, nil
}
