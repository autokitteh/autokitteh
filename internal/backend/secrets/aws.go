package secrets

import (
	"context"
	"encoding/json"

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
func newAWSSecrets(l *zap.Logger, _ *Config) (Secrets, error) {
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
func (s *awsSecrets) Set(ctx context.Context, key string, data map[string]string) error {
	d, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	_, err = s.client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(key),
		SecretString: aws.String(string(d)),
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *awsSecrets) Get(ctx context.Context, key string) (map[string]string, error) {
	sec, err := s.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	data := make(map[string]string)
	if err := json.Unmarshal([]byte(*sec.SecretString), &data); err != nil {
		return nil, err
	}
	return data, nil
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
