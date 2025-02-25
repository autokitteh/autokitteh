package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"go.autokitteh.dev/autokitteh/internal/kvstore"
	"go.autokitteh.dev/autokitteh/internal/kvstore/encoding"
	"go.autokitteh.dev/autokitteh/internal/kvstore/util"
)

var defaultTimeout = 2 * time.Second

// Client is a gokv.Store implementation for Redis.
type client struct {
	c       *redis.Client
	timeOut time.Duration
	codec   encoding.Codec
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c client) Set(ctx context.Context, k string, v any) error {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that Redis can handle
	// (the Set method takes an interface{}, but the Get method only returns a string,
	// so it can be assumed that the interface{} parameter type is only for convenience
	// for a couple of builtin types like int etc.).
	data, err := c.codec.Marshal(v)
	if err != nil {
		return err
	}

	tctx, cancel := context.WithTimeout(ctx, c.timeOut)
	defer cancel()

	err = c.c.Set(tctx, k, string(data), 0).Err()
	if err != nil {
		return err
	}
	return nil
}

// Get retrieves the stored value for the given key.
// You need to pass a pointer to the value, so in case of a struct
// the automatic unmarshalling can populate the fields of the object
// that v points to with the values of the retrieved object's values.
// If no value is found it returns (false, nil).
// The key must not be "" and the pointer must not be nil.
func (c client) Get(ctx context.Context, k string, v any) (found bool, err error) {
	if err := util.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	tctx, cancel := context.WithTimeout(ctx, c.timeOut)
	defer cancel()

	dataString, err := c.c.Get(tctx, k).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	return true, c.codec.Unmarshal([]byte(dataString), v)
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (c client) Delete(ctx context.Context, k string) error {
	if err := util.CheckKey(k); err != nil {
		return err
	}

	tctx, cancel := context.WithTimeout(ctx, c.timeOut)
	defer cancel()

	_, err := c.c.Del(tctx, k).Result()
	return err
}

// Close closes the client.
// It must be called to release any open resources.
func (c client) Close() error {
	return c.c.Close()
}

// Options are the options for the Redis client.
type Options struct {
	// Address of the Redis server, including the port.
	// Optional ("localhost:6379" by default).
	Address string
	// Password for the Redis server.
	// Optional ("" by default).
	Password string
	// DB to use.
	// Optional (0 by default).
	DB int
	// The timeout for operations.
	// Optional (2 * time.Second by default).
	Timeout *time.Duration
	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
// Address: "localhost:6379", Password: "", DB: 0, Timeout: 2 * time.Second, Codec: encoding.JSON
var DefaultOptions = Options{
	Address: "localhost:6379",
	Timeout: &defaultTimeout,
	Codec:   encoding.JSON,
	// No need to set Password or DB because their Go zero values are fine for that.
}

// NewClient creates a new Redis client.
//
// You must call the Close() method on the client when you're done working with it.
func NewClient(options Options) (kvstore.Store, error) {
	result := client{}

	// Set default values
	if options.Address == "" {
		options.Address = DefaultOptions.Address
	}
	if options.Timeout == nil {
		options.Timeout = DefaultOptions.Timeout
	}
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	client := redis.NewClient(&redis.Options{
		Addr:     options.Address,
		Password: options.Password,
		DB:       options.DB,
	})

	tctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := client.Ping(tctx).Err()
	if err != nil {
		return result, err
	}

	result.c = client
	result.timeOut = *options.Timeout
	result.codec = options.Codec

	return result, nil
}
