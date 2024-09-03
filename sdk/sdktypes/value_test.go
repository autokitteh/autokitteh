package sdktypes_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestValueProtoToJSONStringValue(t *testing.T) {
	pb := kittehs.Must1(sdktypes.NewListValue([]sdktypes.Value{
		sdktypes.NewBooleanValue(true),
		sdktypes.NewIntegerValue(42),
	})).ToProto()

	pb1, err := sdktypes.ValueProtoToJSONStringValue(pb)
	if !assert.NoError(t, err) {
		assert.Equal(t, pb1, sdktypes.NewStringValue("[true,42]").ToProto())
	}
}
