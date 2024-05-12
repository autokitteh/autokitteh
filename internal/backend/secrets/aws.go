package secrets

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"go.uber.org/zap"
)

type awsSecrets struct {
	client *secretsmanager.Client
	logger *zap.Logger
	// timeout time.Duration
}

// NewAWSSecrets initializes a client connection to AWS Secrets Manager.
func newAWSSecrets(l *zap.Logger, _ *awsSecretManagerConfig) (Secrets, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		l.Error("AWS config initialization", zap.Error(err))
		return nil, err
	}

	client := secretsmanager.NewFromConfig(cfg)
	return &awsSecrets{client: client, logger: l}, nil
}

// The data size limit is 64 KiB, according to this link:
// https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_CreateSecret.html
func (s *awsSecrets) Set(ctx context.Context, key string, data string) error {
	_, err := s.client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:                        aws.String(key),
		SecretString:                aws.String(data),
		ForceOverwriteReplicaSecret: true,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *awsSecrets) Get(ctx context.Context, key string) (string, error) {
	sec, err := s.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(key),
	})
	if err != nil {
		return "", err
	}

	return *sec.SecretString, nil
}

func (s *awsSecrets) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(key),
		ForceDeleteWithoutRecovery: aws.Bool(true),
	})
	if err != nil {
		return err
	}
	return nil
}
