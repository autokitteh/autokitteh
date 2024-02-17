package sdkmodule

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestModule(t *testing.T) {
	mod := New(
		ExportFunction("say", func(ctx context.Context, _ []sdktypes.Value, _ map[string]sdktypes.Value) (sdktypes.Value, error) {
			data := FunctionDataFromContext(ctx)
			sound := "meow"
			if data != nil {
				sound = string(data)
			}
			return sdktypes.NewStringValue(sound), nil
		}),
		ExportValue("dog", WithNewValue(func(xid sdktypes.ExecutorID, _ []byte) (sdktypes.Value, error) {
			return sdktypes.NewStructValue(
				sdktypes.NewStringValue("dog"),
				map[string]sdktypes.Value{
					"say": sdktypes.NewFunctionValue(xid, "say", []byte("woof"), nil, nil),
				},
			), nil
		})),
	)

	require.NotNil(t, mod)

	vs, err := mod.Configure(context.Background(), sdktypes.NewExecutorID(sdktypes.NewIntegrationID()), "")
	require.NoError(t, err)

	require.Contains(t, vs, "say")

	sayv := vs["say"]

	v, err := mod.Call(context.Background(), sayv, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, v)

	assert.Equal(t, "meow", sdktypes.GetStringValue(v))

	_, dog := sdktypes.GetStructValue(vs["dog"])
	sayv = dog["say"]

	v, err = mod.Call(context.Background(), sayv, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, v)

	assert.Equal(t, "woof", sdktypes.GetStringValue(v))
}
