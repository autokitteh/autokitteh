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

type amazonSecrets struct {
	client  *secretsmanager.Client
	logger  *zap.Logger
	timeout time.Duration
}

// NewAmazonSecrets initializes a client connection to Amazon
// Secrets Manager (https://aws.amazon.com/secrets-manager/).
func NewAmazonSecrets(l *zap.Logger, c *Config) (Secrets, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		l.Error("AWS config initialization", zap.Error(err))
		return nil, err
	}

	client := secretsmanager.NewFromConfig(cfg)
	timeout := parseTimeout(l, c.TimeoutDuration)
	return &amazonSecrets{client: client, logger: l, timeout: timeout}, nil
}

// The data size limit is 64 KiB, according to this link:
// https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_CreateSecret.html
func (s *amazonSecrets) Set(scope, name string, data map[string]string) error {
	ctx, cancel := limitedContext(s.timeout)
	defer cancel()

	d, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		s.logger.Error("JSON marshal",
			zap.String("name", secretPath(scope, name)),
			zap.Error(err),
		)
		return err
	}

	_, err = s.client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(secretPath(scope, name)),
		SecretString: aws.String(string(d)),
	})
	if err != nil {
		s.logger.Error("AWS Secrets Manager CreateSecret",
			zap.String("name", secretPath(scope, name)),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (s *amazonSecrets) Get(scope, name string) (map[string]string, error) {
	ctx, cancel := limitedContext(s.timeout)
	defer cancel()

	sec, err := s.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretPath(scope, name)),
	})
	if err != nil {
		s.logger.Error("AWS Secrets Manager GetSecretValue",
			zap.String("name", secretPath(scope, name)),
			zap.Error(err),
		)
		return nil, err
	}

	data := make(map[string]string)
	if err := json.Unmarshal([]byte(*sec.SecretString), &data); err != nil {
		s.logger.Error("JSON unmarshal",
			zap.String("name", secretPath(scope, name)),
			zap.Error(err),
		)
		return nil, err
	}

	return data, nil
}

func (s *amazonSecrets) Append(scope, name, token string) error {
	ctx, cancel := limitedContext(s.timeout)
	defer cancel()

	data, err := s.Get(scope, name)
	if err != nil {
		return err
	}

	data[token] = time.Now().UTC().Format(time.RFC3339)
	d, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		s.logger.Error("JSON marshal",
			zap.String("name", secretPath(scope, name)),
			zap.Error(err),
		)
		return err
	}

	_, err = s.client.PutSecretValue(ctx, &secretsmanager.PutSecretValueInput{
		SecretId:     aws.String(secretPath(scope, name)),
		SecretString: aws.String(string(d)),
	})
	if err != nil {
		s.logger.Error("AWS Secrets Manager PutSecretValue",
			zap.String("name", secretPath(scope, name)),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (s *amazonSecrets) Delete(scope, name string) error {
	ctx, cancel := limitedContext(s.timeout)
	defer cancel()

	_, err := s.client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(secretPath(scope, name)),
		ForceDeleteWithoutRecovery: aws.Bool(true),
	})
	if err != nil {
		s.logger.Error("AWS Secrets Manager DeleteSecret",
			zap.String("name", secretPath(scope, name)),
			zap.Error(err),
		)
		return err
	}
	return nil
}
