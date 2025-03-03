package temporalclient

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"strconv"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kvstore"
	"go.autokitteh.dev/autokitteh/internal/kvstore/encoding"
	"go.autokitteh.dev/autokitteh/internal/kvstore/file"
	"go.autokitteh.dev/autokitteh/internal/kvstore/gomap"
	"go.autokitteh.dev/autokitteh/internal/kvstore/s3"
	"go.autokitteh.dev/autokitteh/internal/xdg"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	prefix                     = "ak_blob_"
	largeObjectKeyMetadataKey  = prefix + "key"
	largeObjectSizeMetadataKey = prefix + "size"
	largeObjectHashMetadataKey = prefix + "hash"
)

type LargePayloadConfig struct {
	StoreType string `koanf:"store_type"` // "none" to disable.

	// ThresholdSize is the maximum size of a payload that will be stored in Temporal.
	// If a payload is larger than this size, it will be stored in a separate store.
	ThresholdSize int `koanf:"threshold_size"`
	MaxSize       int `koanf:"max_size"` // maximum size of a payload.

	FileRootPath string `koanf:"file_root_path"` // if store_type == "file".
	S3BucketName string `koanf:"s3_bucket_name"` // if store_type == "s3".
}

type StoreType string

const (
	StoreTypeNone  StoreType = "none"
	StoreTypeInMem StoreType = "inmem"
	StoreTypeFile  StoreType = "file"
	StoreTypeS3    StoreType = "s3"
)

type largePayloadCodec struct {
	l     *zap.Logger
	cfg   LargePayloadConfig
	store kvstore.Store
}

func newLargePayloadCodec(l *zap.Logger, cfg LargePayloadConfig) (converter.PayloadCodec, error) {
	var store kvstore.Store

	switch StoreType(cfg.StoreType) {
	case "":
		return nil, errors.New("store type not specified")
	case StoreTypeNone:
		l.Info("large payloads store is disabled")
		return nil, nil
	case StoreTypeInMem:
		store = gomap.NewStore(gomap.Options{Codec: encoding.Gob})
		l.Warn("using in-memory store for large payloads")
	case StoreTypeFile:
		var (
			err error
			ext = "gob"
		)

		dir := cfg.FileRootPath
		if dir == "" {
			dir = filepath.Join(xdg.DataHomeDir(), "temporal-large-payloads")
		}

		if store, err = file.NewStore(file.Options{
			Codec:             encoding.Gob,
			FilenameExtension: &ext,
			Directory:         dir,
		}); err != nil {
			return nil, err
		}

		l.Info("using file store for large payloads", zap.String("root", dir))
	case StoreTypeS3:
		var err error
		if store, err = s3.NewClient(s3.Options{
			BucketName: cfg.S3BucketName,
			Codec:      encoding.Gob,
		}); err != nil {
			return nil, err
		}

		l.Info("using S3 store for large payloads", zap.String("bucket", cfg.S3BucketName))
	default:
		return nil, fmt.Errorf("unknown store type  %q", cfg.StoreType)
	}

	return &largePayloadCodec{cfg: cfg, l: l, store: store}, nil
}

func (c *largePayloadCodec) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	out := make([]*commonpb.Payload, 0, len(payloads))

	for _, in := range payloads {
		data := in.GetData()
		size := len(data)

		l := c.l.With(zap.Int("size", size), zap.Any("metadata", in.GetMetadata()))

		if size < c.cfg.ThresholdSize {
			l.Debug("payload is below threshold size, storing in Temporal", zap.Int("size", size), zap.Int("threshold", c.cfg.ThresholdSize))
			out = append(out, in)
			continue
		}

		if c.cfg.MaxSize != 0 && size > c.cfg.MaxSize {
			l.Error("payload is above maximum size")
			return nil, fmt.Errorf("payload is above maximum size: %d > %d bytes.", size, c.cfg.MaxSize)
		}

		key := sdktypes.NewUUID().String()
		l = l.With(zap.String("key", key))

		l.Debug("payload is above threshold size, storing in external store", zap.Int("size", size), zap.Int("threshold", c.cfg.ThresholdSize))

		if err := c.store.Set(context.Background(), key, in.GetData()); err != nil {
			l.Error("failed to store payload in external store", zap.Error(err))
			return nil, err
		}

		md := maps.Clone(in.Metadata)
		if md == nil {
			md = make(map[string][]byte, 2)
		}

		sha := sha256.Sum256(data)

		md[largeObjectKeyMetadataKey] = []byte(key)
		md[largeObjectSizeMetadataKey] = []byte(strconv.Itoa(size))
		md[largeObjectHashMetadataKey] = sha[:]

		out = append(out, &commonpb.Payload{
			Metadata: md,
			Data:     []byte(key),
		})
	}

	return out, nil
}

func (c *largePayloadCodec) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	out := make([]*commonpb.Payload, 0, len(payloads))

	for _, in := range payloads {
		md := maps.Clone(in.GetMetadata())

		key, ok := md[largeObjectKeyMetadataKey]
		if !ok {
			out = append(out, in)
			continue
		}

		expectedSHA, ok := md[largeObjectHashMetadataKey]
		if !ok {
			c.l.Error("payload hash not found in metadata")
			return nil, errors.New("payload hash not found in metadata")
		}

		l := c.l.With(zap.String("key", string(key)))

		l.Debug("payload is stored in external store, fetching")

		var data []byte
		found, err := c.store.Get(context.Background(), string(key), &data)
		if err != nil {
			l.Error("failed to fetch payload from external store", zap.Error(err))
			return nil, err
		}

		if !found {
			l.Error("payload not found in external store")
			return nil, err
		}

		sha := sha256.Sum256(data)
		if !bytes.Equal(sha[:], expectedSHA) {
			l.Error("payload hash mismatch")
			return nil, errors.New("payload hash mismatch")
		}

		delete(md, largeObjectHashMetadataKey)
		delete(md, largeObjectKeyMetadataKey)
		delete(md, largeObjectSizeMetadataKey)

		out = append(out, &commonpb.Payload{
			Metadata: md,
			Data:     data,
		})
	}

	return out, nil
}
