package redis

import (
	"encoding"
	"fmt"
	"net"
	"time"

	"github.com/redis/go-redis/v9"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func unwrap(v sdktypes.Value) (any, error) {
	u, err := sdktypes.DefaultValueWrapper.Unwrap(v)
	if err != nil {
		return nil, err
	}

	// All supported redis types, adapted from https://github.com/redis/go-redis/blob/master/internal/proto/writer.go#L62.
	switch u.(type) {
	case struct{}: // Nothing unwraps to struct{}{}
		u = nil
	case nil, string, []byte,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		bool,
		time.Time, time.Duration:
		// just using unwrapped is fine, no need to modify it.
	case encoding.BinaryMarshaler:
		// TODO: figure out how it's used in the redis client.
		return nil, fmt.Errorf("unhandled type")
	case net.IP:
		// TODO: return w.bytes(v)
		return nil, fmt.Errorf("unhandled type")
	default:
		return nil, fmt.Errorf("unhandled type")
	}

	return u, nil
}

type resulter[R any] interface {
	Result() (R, error)
}

func returnCmd[R any, C resulter[R]](cmd C) (sdktypes.Value, error) {
	ret, err := cmd.Result()
	switch {
	case err == redis.Nil:
		return sdktypes.Nothing, nil
	case err != nil:
		return sdktypes.InvalidValue, err
	}

	wrapped, err := sdktypes.WrapValue(ret)
	if err != nil {
		return sdktypes.InvalidValue, fmt.Errorf("wrap: %w", err)
	}

	return wrapped, nil
}
