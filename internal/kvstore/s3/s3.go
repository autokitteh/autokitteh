package s3

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/aws/aws-sdk-go-v2/config"
	s3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"go.autokitteh.dev/autokitteh/internal/kvstore/checks"
	"go.autokitteh.dev/autokitteh/internal/kvstore/encoding"
)

// Client is a kvstore.Store implementation for S3.
type Client struct {
	c          *s3.Client
	bucketName string
	codec      encoding.Codec
}

// Set stores the given value for the given key.
// Values are automatically marshalled to JSON or gob (depending on the configuration).
// The key must not be "" and the value must not be nil.
func (c Client) Set(ctx context.Context, k string, v any) error {
	if err := checks.CheckKeyAndValue(k, v); err != nil {
		return err
	}

	// First turn the passed object into something that S3 can handle.
	data, err := c.codec.Marshal(v)
	if err != nil {
		return err
	}

	pubObjectInput := s3.PutObjectInput{
		Body:   bytes.NewReader(data),
		Bucket: &c.bucketName,
		Key:    &k,
	}
	_, err = c.c.PutObject(ctx, &pubObjectInput)
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
func (c Client) Get(ctx context.Context, k string, v any) (found bool, err error) {
	if err := checks.CheckKeyAndValue(k, v); err != nil {
		return false, err
	}

	getObjectInput := s3.GetObjectInput{
		Bucket: &c.bucketName,
		Key:    &k,
	}
	getObjectOutput, err := c.c.GetObject(ctx, &getObjectInput)
	if err != nil {
		var nf *types.NotFound
		if errors.As(err, &nf) {
			return false, nil
		}
		return false, err
	}
	if getObjectOutput.Body == nil {
		// Return false if there's no value
		// TODO: Maybe return an error? Behaviour should be consistent across all implementations.
		return false, nil
	}
	data, err := io.ReadAll(getObjectOutput.Body)
	if err != nil {
		return true, err
	}

	return true, c.codec.Unmarshal(data, v)
}

// Delete deletes the stored value for the given key.
// Deleting a non-existing key-value pair does NOT lead to an error.
// The key must not be "".
func (c Client) Delete(ctx context.Context, k string) error {
	if err := checks.CheckKey(k); err != nil {
		return err
	}

	deleteObjectInput := s3.DeleteObjectInput{
		Bucket: &c.bucketName,
		Key:    &k,
	}
	_, err := c.c.DeleteObject(ctx, &deleteObjectInput)
	return err
}

// Close closes the client.
// In the S3 implementation this doesn't have any effect.
func (c Client) Close() error {
	return nil
}

// Options are the options for the S3 client.
type Options struct {
	// Name of the S3 bucket.
	// The bucket is automatically created if it doesn't exist yet.
	BucketName string

	// Encoding format.
	// Optional (encoding.JSON by default).
	Codec encoding.Codec
}

// DefaultOptions is an Options object with default values.
var DefaultOptions = Options{
	Codec: encoding.JSON,
	// No need to set Region, AWSaccessKeyID, AWSsecretAccessKey
	// CustomEndpoint or UsePathStyleAddressing because their Go zero values are fine.
}

// NewClient creates a new S3 client.
//
// Credentials can be set in the options, but it's recommended to either use the shared credentials file
// (Linux: "~/.aws/credentials", Windows: "%UserProfile%\.aws\credentials")
// or environment variables (AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY).
// See https://github.com/awsdocs/aws-go-developer-guide/blob/0ae5712d120d43867cf81de875cb7505f62f2d71/doc_source/configuring-sdk.rst#specifying-credentials.
func NewClient(options Options) (Client, error) {
	result := Client{}

	// Precondition check
	if options.BucketName == "" {
		return result, errors.New("The BucketName in the options must not be empty")
	}

	// Set default values
	if options.Codec == nil {
		options.Codec = DefaultOptions.Codec
	}

	// Set credentials only if set in the options.
	// If not set, the SDK uses the shared credentials file or environment variables, which is the preferred way.
	// Return an error if only one of the values is set.
	awsCfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return result, err
	}

	return Client{
		c:          s3.NewFromConfig(awsCfg),
		bucketName: options.BucketName,
		codec:      options.Codec,
	}, nil
}
