package temporalclient

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/converter"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const keySize = 32

func initConverter(t *testing.T, r io.Reader, keyNames []string) (converter.DataConverter, error) {
	keys := make(map[string]string)

	for _, k := range keyNames {
		bs := make([]byte, keySize)
		n, err := io.ReadFull(r, bs)
		require.NoError(t, err)
		require.Equal(t, keySize, n)
		keys[k] = hex.EncodeToString(bs)
		t.Logf("key %q: %s", k, keys[k])
	}

	return NewDataConverter(
		zap.NewNop(),
		&DataConverterConfig{
			Compress: true,
			Encryption: dataConverterEncryptionConfig{
				Encrypt: true,
				Keys:    strings.Join(kittehs.Transform(keyNames, func(k string) string { return k + "=" + keys[k] }), ","),
			},
		},
		converter.GetDefaultDataConverter(),
	)
}

func TestNoKeys(t *testing.T) {
	_, err := initConverter(t, rand.Reader, nil)
	assert.EqualError(t, err, ErrNoKeys.Error())
}

func TestSameKey(t *testing.T) {
	cvt := kittehs.Must1(initConverter(t, rand.Reader, []string{"key1", "key2"}))

	v1 := sdktypes.NewStringValue("meow, world!")
	v2 := "woof, world"

	encoded, err := cvt.ToPayloads(v1, v2)
	if assert.NoError(t, err) && assert.Len(t, encoded.GetPayloads(), 2) {
		for _, p := range encoded.GetPayloads() {
			md := p.GetMetadata()
			assert.Equal(t, metadataEncryptionEncoding, string(md[converter.MetadataEncoding]))
			assert.Equal(t, "key1", string(md[metadataEncryptionKeyID]))
		}
	}

	var (
		vv1 sdktypes.Value
		vv2 string
	)

	if assert.NoError(t, cvt.FromPayloads(encoded, &vv1, &vv2)) {
		assert.Equal(t, v1, vv1)
		assert.Equal(t, v2, vv2)
	}
}

func TestOldKey(t *testing.T) {
	buf := make([]byte, 2*keySize)
	_, _ = io.CopyN(bytes.NewBuffer(buf), rand.Reader, int64(cap(buf)))

	cvt1 := kittehs.Must1(initConverter(t, bytes.NewReader(buf), []string{"key2"}))

	v1 := sdktypes.NewStringValue("meow, world!")
	v2 := "woof, world"

	encoded, err := cvt1.ToPayloads(v1, v2)
	require.NoError(t, err)

	cvt2 := kittehs.Must1(initConverter(t, bytes.NewReader(buf), []string{"key1", "key2"}))

	var (
		vv1 sdktypes.Value
		vv2 string
	)

	if assert.NoError(t, cvt2.FromPayloads(encoded, &vv1, &vv2)) {
		assert.Equal(t, v1, vv1)
		assert.Equal(t, v2, vv2)
	}
}
