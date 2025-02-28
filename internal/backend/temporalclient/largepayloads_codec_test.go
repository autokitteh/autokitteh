package temporalclient

import (
	"context"
	"crypto/sha256"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/api/common/v1"
	"go.uber.org/zap/zaptest"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var cfg = LargePayloadConfig{
	ThresholdSize: 1024,
	StoreType:     "inmem",
}

func TestBelowThreshold(t *testing.T) {
	cvt, err := newLargePayloadCodec(zaptest.NewLogger(t), cfg)
	require.NoError(t, err)

	in := []*common.Payload{
		{
			Data: []byte("meow"),
		},
	}

	encoded, err := cvt.Encode(in)
	require.NoError(t, err)
	require.Equal(t, []*common.Payload{
		{
			Data: []byte("meow"),
		},
	}, encoded)

	decoded, err := cvt.Decode(encoded)
	require.NoError(t, err)
	require.Equal(t, in, decoded)
}

func TestAboveThreshold(t *testing.T) {
	sdktypes.SetIDGenerator(sdktypes.NewSequentialIDGeneratorForTesting(0))
	expectedID := []byte(sdktypes.SequentialIDForTesting(1).String())
	expectedData := []byte(strings.Repeat("meow", 1024))
	expectedSHA := sha256.Sum256(expectedData)

	cvt, err := newLargePayloadCodec(zaptest.NewLogger(t), cfg)
	require.NoError(t, err)

	out, err := cvt.Encode([]*common.Payload{
		{
			Data: expectedData,
		},
	})
	if assert.NoError(t, err) {
		assert.Equal(t, []*common.Payload{
			{
				Metadata: map[string][]byte{
					largeObjectHashMetadataKey: expectedSHA[:],
					largeObjectKeyMetadataKey:  expectedID,
					largeObjectSizeMetadataKey: []byte(strconv.Itoa(len(expectedData))),
				},
				Data: expectedID,
			},
		}, out)
	}

	codec := cvt.(*largePayloadCodec)

	data := make([]byte, len(expectedData))
	found, err := codec.store.Get(context.TODO(), string(expectedID), &data)
	if assert.True(t, found) && assert.NoError(t, err) {
		assert.Equal(t, expectedData, data)
	}

	decoded, err := cvt.Decode(out)
	if assert.NoError(t, err) {
		assert.Equal(t, []*common.Payload{
			{
				Metadata: map[string][]byte{},
				Data:     expectedData,
			},
		}, decoded)
	}
}
