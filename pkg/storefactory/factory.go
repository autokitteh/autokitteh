package storefactory

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"

	"github.com/autokitteh/autokitteh/pkg/gormfactory"
	L "github.com/autokitteh/autokitteh/pkg/l"
)

type GRPCConfig struct {
	HostPort string `envconfig:"HOSTPORT" default:"127.0.0.1:20001" json:"hostport"`
}

type RedisConfig struct {
	URL string `envconfig:"URL" default:"redis://localhost:6789/0" json:"url"`
}

type InMemConfig struct {
	Config string `envconfig:"CONFIG" json:"config"`
}

type FSConfig struct {
	Options  map[string]string `envconfig:"OPTIONS" json:"options"`
	RootPath string            `envconfig:"ROOT_PATH" json:"root_path"`
}

type OtherConfig struct {
	Config string `envconfig:"CONFIG" json:"config"`
}

type Config struct {
	// If not empty, used from OpenString. All else ignored.
	Spec string `envconfig:"SPEC" json:"spec"`

	// If Type is empty, will try to use FromDefault.
	Type string `envconfig:"TYPE" json:"type"`

	GRPC  GRPCConfig         `envconfig:"GRPC" json:"grpc"`   // type="grpc"
	GORM  gormfactory.Config `envconfig:"GORM" json:"gorm"`   // type="gorm"
	Redis RedisConfig        `envconfig:"REDIS" json:"redis"` // type="redis"
	InMem InMemConfig        `envconfig:"INMEM" json:"inmem"` // type="inmem"
	FS    FSConfig           `envconfig:"FS" json:"fs"`       // type="fs"
	Other OtherConfig        `envconfig:"OTHER" json:"other"` // otherwise
}

func (c Config) IsSet() bool { return c.Spec != "" || c.Type != "" }

type Factory struct {
	FromDefault  func(context.Context, L.L) (interface{}, error)
	FromGORM     func(_ context.Context, _ L.L, _ *gorm.DB) (interface{}, error)
	FromGRPCConn func(context.Context, L.L, *grpc.ClientConn) (interface{}, error)
	FromRedis    func(context.Context, L.L, *redis.Client) (interface{}, error)
	FromInMem    func(context.Context, L.L, *InMemConfig) (interface{}, error)
	FromFS       func(context.Context, L.L, *FSConfig) (interface{}, error)
	FromOther    func(context.Context, L.L, string, *OtherConfig) (interface{}, error)
}

func (f *Factory) withDefaults() *Factory {
	ff := *f

	if ff.FromDefault == nil {
		ff.FromDefault = func(context.Context, L.L) (interface{}, error) {
			return nil, fmt.Errorf("no default set")
		}
	}

	if ff.FromGORM == nil {
		ff.FromGORM = func(context.Context, L.L, *gorm.DB) (interface{}, error) {
			return nil, fmt.Errorf("gorm not supported")
		}
	}

	if ff.FromGRPCConn == nil {
		ff.FromGRPCConn = func(context.Context, L.L, *grpc.ClientConn) (interface{}, error) {
			return nil, fmt.Errorf("grpc not supported")
		}
	}

	if ff.FromRedis == nil {
		ff.FromRedis = func(context.Context, L.L, *redis.Client) (interface{}, error) {
			return nil, fmt.Errorf("redis not supported")
		}
	}

	if ff.FromInMem == nil {
		ff.FromInMem = func(context.Context, L.L, *InMemConfig) (interface{}, error) {
			return nil, fmt.Errorf("inmem not supported")
		}
	}

	if ff.FromFS == nil {
		ff.FromFS = func(context.Context, L.L, *FSConfig) (interface{}, error) {
			return nil, fmt.Errorf("fs not supported")
		}
	}

	if ff.FromOther == nil {
		ff.FromOther = func(_ context.Context, _ L.L, typ string, _ *OtherConfig) (interface{}, error) {
			return nil, fmt.Errorf("unrecognized db type %q", typ)
		}
	}

	return &ff
}

func ParseConfigString(text string /* "type:rest" */) *Config {
	if len(text) == 0 {
		return &Config{}
	}

	typ, rest, ok := strings.Cut(text, ":")

	cfg := Config{Type: typ}

	switch cfg.Type {
	case "gorm":
		if ok {
			cfg.GORM.Type, cfg.GORM.DSN, _ = strings.Cut(rest, ":")
		}

	case "redis":
		cfg.Redis.URL = rest

	case "grpc":
		cfg.GRPC.HostPort = rest

	case "fs":
		if before, after, ok := strings.Cut(rest, ","); ok {
			cfg.FS.RootPath = before

			parts := strings.Split(after, ",")
			cfg.FS.Options = make(map[string]string, len(parts))

			for _, p := range parts {
				a, b, _ := strings.Cut(p, "=")
				cfg.FS.Options[a] = b
			}
		} else {
			cfg.FS.RootPath = rest
		}

	case "inmem":
		cfg.InMem.Config = rest

	case "other":
		cfg.Other.Config = rest
	}

	return &cfg
}

func (f *Factory) MustOpenString(ctx context.Context, l L.L, text string) interface{} {
	return f.MustOpen(ctx, l, ParseConfigString(text))
}

func (f *Factory) OpenString(ctx context.Context, l L.L, text string) (interface{}, error) {
	return f.Open(ctx, l, ParseConfigString(text))
}

func (f *Factory) MustOpen(ctx context.Context, l L.L, cfg *Config) interface{} {
	s, err := f.Open(ctx, l, cfg)
	if err != nil {
		panic(err)
	}
	return s
}

func (f *Factory) Open(ctx context.Context, l L.L, cfg *Config) (interface{}, error) {
	if cfg.Spec != "" {
		return f.OpenString(ctx, l, cfg.Spec)
	}

	fwd := f.withDefaults()

	if cfg == nil {
		cfg = &Config{}
	}

	var out interface{}

	switch cfg.Type {
	case "", "default":
		var err error
		if out, err = fwd.FromDefault(ctx, l); err != nil {
			return nil, fmt.Errorf("default: %w", err)
		}

	case "gorm":
		db, err := gormfactory.Open(cfg.GORM)
		if err != nil {
			return nil, err
		}

		if out, err = fwd.FromGORM(ctx, l, db); err != nil {
			return nil, fmt.Errorf("gorm: %w", err)
		}

	case "inmem":
		var err error
		if out, err = fwd.FromInMem(ctx, l, &cfg.InMem); err != nil {
			return nil, fmt.Errorf("inmem: %w", err)
		}

	case "fs":
		var err error
		if out, err = f.FromFS(ctx, l, &cfg.FS); err != nil {
			return nil, fmt.Errorf("fs: %w", err)
		}

	case "grpc":
		conn, err := grpc.Dial(cfg.GRPC.HostPort, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err != nil {
			return nil, fmt.Errorf("grpc dial: %w", err)
		}

		if out, err = f.FromGRPCConn(ctx, l, conn); err != nil {
			return nil, fmt.Errorf("grpc: %w", err)
		}

	case "redis":
		url := cfg.Redis.URL

		if strings.HasPrefix(cfg.Redis.URL, "//") {
			url = "redis:" + cfg.Redis.URL
		} else if !strings.HasPrefix(cfg.Redis.URL, "redis://") {
			url = "redis://" + cfg.Redis.URL
		}

		opts, err := redis.ParseURL(url)
		if err != nil {
			return nil, fmt.Errorf("redis url: %w", err)
		}

		client := redis.NewClient(opts)

		if out, err = f.FromRedis(ctx, l, client); err != nil {
			return nil, fmt.Errorf("redis: %w", err)
		}

	default:
		var err error
		if out, err = fwd.FromOther(ctx, l, cfg.Type, &cfg.Other); err != nil {
			return nil, err
		}
	}

	return out, nil
}
