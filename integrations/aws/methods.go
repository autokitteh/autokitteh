package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type AWSVars struct {
	AccessKeyID  string
	Region       string
	SecretKey    string
	SessionToken string
}

type AWSAPI struct {
	Vars AWSVars
}

type TestResponse struct {
	Buckets []string `json:"Buckets"`
}

func (a AWSAPI) Test(ctx context.Context) (*TestResponse, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(a.Vars.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				a.Vars.AccessKeyID,
				a.Vars.SecretKey,
				a.Vars.SessionToken)),
	)
	if err != nil {
		return nil, err
	}

	// Create an S3 client
	s3Client := s3.NewFromConfig(cfg)

	resp, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, fmt.Errorf("invalid response from ListBuckets")
	}

	// Parse and return the response
	bucketNames := []string{}
	for _, bucket := range resp.Buckets {
		bucketNames = append(bucketNames, aws.ToString(bucket.Name))
	}

	return &TestResponse{
		Buckets: bucketNames,
	}, nil
}
