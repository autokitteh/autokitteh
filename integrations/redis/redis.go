package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// TODO: make redis work with structs (need to convert them to stringdicts).

type module struct {
	// set only for external clients
	vars sdkservices.Vars

	// set only for internal clients
	internalClient *redis.Client
	keyfn          func(string) string
}

func (m *module) client(ctx context.Context) (*redis.Client, func(string) string, error) {
	if m.internalClient != nil {
		return m.internalClient, m.keyfn, nil
	}

	c, err := m.externalClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	return c, func(k string) string { return k }, nil
}

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: sdktypes.NewIntegrationIDFromName("redis").String(),
	UniqueName:    "redis",
	DisplayName:   "Redis",
	Description:   "Redis is an open-source, in-memory data structure store, used as a database, cache, and message broker.",
	LogoUrl:       "/static/images/redis.svg",
	UserLinks: map[string]string{
		"1 Redis command reference": "https://redis.io/commands/",
		"2 Go client API":           "https://pkg.go.dev/github.com/go-redis/redis/v9",
	},
}))

func New(vars sdkservices.Vars) sdkservices.Integration {
	m := &module{vars: vars}

	opts := makeOpts(m)

	// Allow these only for non-internal clients. Otherwise we cannot guarantee proper prefixes
	// for command keys.
	opts = append(opts, sdkmodule.ExportFunction("do", m.do, sdkmodule.WithFuncDoc("run an arbitrary command")))

	return sdkintegrations.NewIntegration(desc, sdkmodule.New(opts...), sdkintegrations.WithConnectionConfigFromVars(vars))
}

func NewInternalModule(name string, xid sdktypes.ExecutorID, client *redis.Client, keyfn func(string) string) sdkmodule.Module {
	opts := makeOpts(&module{internalClient: client, keyfn: keyfn})
	return sdkmodule.New(opts...)
}

func makeOpts(m *module) []sdkmodule.Optfn {
	// TODO(ENG-131): Implement more redis commands.
	return []sdkmodule.Optfn{
		sdkmodule.ExportValue(
			"nil",
			sdkmodule.WithValue(sdktypes.NewStringValue(string(redis.Nil)))),
		sdkmodule.ExportFunction(
			"delete", // Starlark doesn't allow "del" as a function name.
			m.del,
			sdkmodule.WithFuncDoc("https://redis.io/commands/del/"),
			sdkmodule.WithArgs("*keys")),
		sdkmodule.ExportFunction(
			"expire",
			m.expire,
			sdkmodule.WithFuncDoc("https://redis.io/commands/expire/"),
			sdkmodule.WithArgs("key", "expiration", "nx?", "xx?", "gt?", "lt?")),
		// TODO: Format the lines below as multi-line blocks like all the other integrations.
		sdkmodule.ExportFunction("set", m.set, sdkmodule.WithFuncDoc("https://redis.io/commands/set/"), sdkmodule.WithArgs("key", "value", "ttl?")),
		sdkmodule.ExportFunction("get", m.get, sdkmodule.WithFuncDoc("https://redis.io/commands/get/"), sdkmodule.WithArgs("key")),
		sdkmodule.ExportFunction("incr", m.incr, sdkmodule.WithFuncDoc("https://redis.io/commands/incr/"), sdkmodule.WithArgs("key")),
		sdkmodule.ExportFunction("decr", m.decr, sdkmodule.WithFuncDoc("https://redis.io/commands/decr/"), sdkmodule.WithArgs("key")),
		sdkmodule.ExportFunction("incrby", m.incr, sdkmodule.WithFuncDoc("https://redis.io/commands/incrby/"), sdkmodule.WithArgs("key", "by?")),
		sdkmodule.ExportFunction("decrby", m.decr, sdkmodule.WithFuncDoc("https://redis.io/commands/decrby/"), sdkmodule.WithArgs("key", "by?")),
		sdkmodule.ExportFunction("lrange", m.lrange, sdkmodule.WithFuncDoc("https://redis.io/commands/lrange/"), sdkmodule.WithArgs("key", "start", "stop")),
		sdkmodule.ExportFunction("ltrim", m.ltrim, sdkmodule.WithFuncDoc("https://redis.io/commands/ltrim/"), sdkmodule.WithArgs("key", "start", "stop")),
		sdkmodule.ExportFunction("lpush", m.lpush, sdkmodule.WithFuncDoc("https://redis.io/commands/lpush/"), sdkmodule.WithArgs("key", "*vs")),
		sdkmodule.ExportFunction("lpushx", m.lpushx, sdkmodule.WithFuncDoc("https://redis.io/commands/lpushx/"), sdkmodule.WithArgs("key", "*vs")),
		sdkmodule.ExportFunction("lpos", m.lpos, sdkmodule.WithFuncDoc("https://redis.io/commands/lpos/"), sdkmodule.WithArgs("key", "value", "rank?", "max_len?")),
		sdkmodule.ExportFunction("llen", m.llen, sdkmodule.WithFuncDoc("https://redis.io/commands/llen/"), sdkmodule.WithArgs("key")),
		sdkmodule.ExportFunction("lset", m.lset, sdkmodule.WithFuncDoc("https://redis.io/commands/lset/"), sdkmodule.WithArgs("key", "index", "value")),
		sdkmodule.ExportFunction("lpop", m.lpop, sdkmodule.WithFuncDoc("https://redis.io/commands/lpop/"), sdkmodule.WithArgs("key")),
		sdkmodule.ExportFunction("lindex", m.lindex, sdkmodule.WithFuncDoc("https://redis.io/commands/lindex/"), sdkmodule.WithArgs("key", "index")),
		sdkmodule.ExportFunction("linsert", m.linsert, sdkmodule.WithFuncDoc("https://redis.io/commands/linsert/"), sdkmodule.WithArgs("key", "op", "pivot", "value")),
		sdkmodule.ExportFunction("lrem", m.lrem, sdkmodule.WithFuncDoc("https://redis.io/commands/lrem/"), sdkmodule.WithArgs("key")),
		sdkmodule.ExportFunction("rpush", m.rpush, sdkmodule.WithFuncDoc("https://redis.io/commands/rpush/"), sdkmodule.WithArgs("key", "*vs")),
		sdkmodule.ExportFunction("rpushx", m.rpushx, sdkmodule.WithFuncDoc("https://redis.io/commands/rpushx/"), sdkmodule.WithArgs("key", "*vs")),
		sdkmodule.ExportFunction("rpop", m.rpop, sdkmodule.WithFuncDoc("https://redis.io/commands/rpop/"), sdkmodule.WithArgs("key")),
		sdkmodule.ExportFunction("rpoplpush", m.rpoplpush, sdkmodule.WithFuncDoc("https://redis.io/commands/rpoplpush/"), sdkmodule.WithArgs("src", "dst")),
		sdkmodule.ExportFunction("brpoplpush", m.brpoplpush, sdkmodule.WithFuncDoc("https://redis.io/commands/rpoplpush/"), sdkmodule.WithArgs("src", "dst")),
		sdkmodule.ExportFunction("brpop", m.brpop, sdkmodule.WithFuncDoc("https://redis.io/commands/brpop/"), sdkmodule.WithArgs("timeout", "*keys")),
		sdkmodule.ExportFunction("blpop", m.blpop, sdkmodule.WithFuncDoc("https://redis.io/commands/blpop/"), sdkmodule.WithArgs("timeout", "*keys")),
		sdkmodule.ExportFunction("dbsize", m.dbsize, sdkmodule.WithFuncDoc("https://redis.io/commands/dbsize/")),
	}
}

// This function is not included in the integration when used as an internal client. See `NewModule`.
func (m *module) do(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	if len(kwargs) != 0 {
		return sdktypes.InvalidValue, errors.New("not expecting kwargs")
	}

	unwrappedArgs := make([]any, len(args))
	for i, v := range args {
		var err error
		if unwrappedArgs[i], err = unwrap(v); err != nil {
			return sdktypes.InvalidValue, fmt.Errorf("arg %d: %w", i, err)
		}
	}

	client, err := m.externalClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return returnCmd(client.Do(ctx, unwrappedArgs...))
}

// https://redis.io/commands/del/
func (m *module) del(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var ks []string
	err := sdkmodule.UnpackArgs(args, kwargs,
		"*keys", &ks,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return returnCmd(client.Del(ctx, kittehs.Transform(ks, keyfn)...))
}

// https://redis.io/commands/expire/
func (m *module) expire(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		key string
		ttl time.Duration

		nx, xx, gt, lt bool
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"key", &key,
		"expiration", &ttl,
		"nx?", &nx,
		"xx?", &xx,
		"gt?", &gt,
		"lt?", &lt,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	switch {
	case nx:
		return returnCmd(client.ExpireNX(ctx, keyfn(key), ttl))
	case xx:
		return returnCmd(client.ExpireXX(ctx, keyfn(key), ttl))
	case gt:
		return returnCmd(client.ExpireGT(ctx, keyfn(key), ttl))
	case lt:
		return returnCmd(client.ExpireLT(ctx, keyfn(key), ttl))
	default:
		return returnCmd(client.Expire(ctx, keyfn(key), ttl))
	}
}

func (m *module) set(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		k   string
		v   sdktypes.Value
		ttl time.Duration
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "key", &k, "value", &v, "ttl?", &ttl); err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	u, err := unwrap(v)
	if err != nil {
		return sdktypes.InvalidValue, fmt.Errorf("value: %w", err)
	}

	return returnCmd(client.Set(ctx, keyfn(k), u, ttl))
}

func (m *module) get(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var k string

	if err := sdkmodule.UnpackArgs(args, kwargs, "key", &k); err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return returnCmd(client.Get(ctx, keyfn(k)))
}

func (m *module) incr(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		k  string
		by int64
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "key", &k, "by?", &by); err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return returnCmd(client.IncrBy(ctx, keyfn(k), by))
}

func (m *module) decr(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		k  string
		by int64
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "key", &k, "by?", &by); err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return returnCmd(client.DecrBy(ctx, keyfn(k), by))
}

func (m *module) lpos(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		k, v         string
		rank, maxLen int64
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "key", &k, "value", &v, "rank?", &rank, "max_len", &maxLen); err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return returnCmd(client.LPos(ctx, keyfn(k), v, redis.LPosArgs{Rank: rank, MaxLen: maxLen}))
}

func (m *module) llen(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var k string

	if err := sdkmodule.UnpackArgs(args, kwargs, "key", &k); err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return returnCmd(client.LLen(ctx, keyfn(k)))
}

func (m *module) lset(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		k     string
		v     any
		index int64
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "key", &k, "index", &index, "value", &v); err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return returnCmd(client.LSet(ctx, keyfn(k), index, v))
}

func (m *module) lrem(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		k     string
		v     any
		count int64
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "key", &k, "value", &v); err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return returnCmd(client.LRem(ctx, keyfn(k), count, v))
}

func (m *module) rpoplpush(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var src, dst string

	if err := sdkmodule.UnpackArgs(args, kwargs, "src", &src, "dst", &dst); err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return returnCmd(client.RPopLPush(ctx, keyfn(src), keyfn(dst)))
}

func (m *module) brpoplpush(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		src, dst string
		t        time.Duration
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "src", &src, "dst", &dst, "timeout", &t); err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return returnCmd(client.BRPopLPush(ctx, keyfn(src), keyfn(dst), t))
}

func (m *module) lpush(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		k  string
		vs []any
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "key", &k, "*vs", &vs); err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return returnCmd(client.LPush(ctx, keyfn(k), vs...))
}

func (m *module) lpushx(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		k  string
		vs []any
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "key", &k, "*vs", &vs); err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return returnCmd(client.LPushX(ctx, keyfn(k), vs...))
}

func (m *module) lpop(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var k string

	if err := sdkmodule.UnpackArgs(args, kwargs, "key", &k); err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return returnCmd(client.LPop(ctx, keyfn(k)))
}

func (m *module) lrange(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		k           string
		start, stop int64
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "key", &k, "start", &start, "stop", &stop); err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return returnCmd(client.LRange(ctx, keyfn(k), start, stop))
}

func (m *module) ltrim(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		k           string
		start, stop int64
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "key", &k, "start", &start, "stop", &stop); err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return returnCmd(client.LTrim(ctx, keyfn(k), start, stop))
}

func (m *module) rpush(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		k  string
		vs []any
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "key", &k, "*vs", &vs); err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return returnCmd(client.RPush(ctx, keyfn(k), vs...))
}

func (m *module) rpushx(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		k  string
		vs []any
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "key", &k, "*vs", &vs); err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return returnCmd(client.RPushX(ctx, keyfn(k), vs...))
}

func (m *module) rpop(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var k string

	if err := sdkmodule.UnpackArgs(args, kwargs, "key", &k); err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return returnCmd(client.RPop(ctx, keyfn(k)))
}

func (m *module) brpop(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		ks []string
		t  time.Duration
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "timeout", &t, "*keys", &ks); err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return returnCmd(client.BRPop(ctx, t, kittehs.Transform(ks, keyfn)...))
}

func (m *module) blpop(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		ks []string
		t  time.Duration
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "timeout", &t, "*keys", &ks); err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return returnCmd(client.BLPop(ctx, t, kittehs.Transform(ks, keyfn)...))
}

func (m *module) lindex(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		k     string
		index int64
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "key", &k, "index", &index); err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return returnCmd(client.LIndex(ctx, keyfn(k), index))
}

func (m *module) linsert(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var (
		k, op    string
		pivot, v any
	)

	if err := sdkmodule.UnpackArgs(args, kwargs, "key", &k, "op", &op, "pivot", &pivot, "value", &v); err != nil {
		return sdktypes.InvalidValue, err
	}

	client, keyfn, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return returnCmd(client.LInsert(ctx, keyfn(k), op, pivot, v))
}

func (m *module) dbsize(ctx context.Context, _ []sdktypes.Value, _ map[string]sdktypes.Value) (sdktypes.Value, error) {
	client, _, err := m.client(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	return returnCmd(client.DBSize(ctx))
}
