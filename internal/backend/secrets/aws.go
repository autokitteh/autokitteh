package secrets

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"go.uber.org/zap"
)

type awsSecrets struct {
	client  *secretsmanager.Client
	logger  *zap.Logger
	timeout time.Duration
}

// NewAWSSecrets initializes a client connection to AWS Secrets Manager.
func NewAWSSecrets(l *zap.Logger, c *Config) (Secrets, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		l.Error("AWS config initialization", zap.Error(err))
		return nil, err
	}

	client := secretsmanager.NewFromConfig(cfg)
	return &awsSecrets{client: client, logger: l, timeout: c.Timeout}, nil
}

// The data size limit is 64 KiB, according to this link:
// https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_CreateSecret.html
func (s *awsSecrets) Set(ctx context.Context, scope, name string, data map[string]string) error {
	ctx, cancel := limitContext(ctx, s.timeout)
	defer cancel()

	d, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	_, err = s.client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretPath(scope, name)),
		SecretString: aws.String(string(d)),
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *awsSecrets) Get(ctx context.Context, scope, name string) (map[string]string, error) {
	ctx, cancel := limitContext(ctx, s.timeout)
	defer cancel()

	sec, err := s.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretPath(scope, name)),
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

func (s *awsSecrets) Append(ctx context.Context, scope, name, token string) error {
	ctx, cancel := limitContext(ctx, s.timeout)
	defer cancel()

	data, err := s.Get(ctx, scope, name)
	if err != nil {
		return err
	}

	data[token] = time.Now().UTC().Format(time.RFC3339)
	d, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	_, err = s.client.PutSecretValue(ctx, &secretsmanager.PutSecretValueInput{
		SecretId:     aws.String(secretPath(scope, name)),
		SecretString: aws.String(string(d)),
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *awsSecrets) Delete(ctx context.Context, scope, name string) error {
	ctx, cancel := limitContext(ctx, s.timeout)
	defer cancel()

	_, err := s.client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(secretPath(scope, name)),
		ForceDeleteWithoutRecovery: aws.Bool(true),
	})
	if err != nil {
		return err
	}
	return nil
}
