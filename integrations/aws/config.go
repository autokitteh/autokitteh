package aws

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type authData struct {
	Region      string
	AccessKeyID string `var:"secret"`
	SecretKey   string `var:"secret"`
	Token       string `var:"secret"`
}

var ErrNotInit = errors.New("not initialized")

func getAWSConfig(ctx context.Context, vars sdkservices.Vars) (*aws.Config, error) {
	cid, err := sdkmodule.FunctionConnectionIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if !cid.IsValid() {
		if defaultAWSConfig == nil {
			return nil, sdkerrors.ErrNotInitialized
		}

		return defaultAWSConfig, nil
	}

	cvars, err := vars.Reveal(ctx, sdktypes.NewVarScopeID(cid))
	if err != nil {
		return nil, err
	}

	if len(cvars) == 0 {
		if defaultAWSConfig == nil {
			return nil, sdkerrors.ErrNotInitialized
		}

		return defaultAWSConfig, nil
	}

	var authData authData
	cvars.Decode(&authData)

	awsCfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(authData.AccessKeyID, authData.SecretKey, authData.Token)),
		config.WithRegion(authData.Region),
	)

	return &awsCfg, err
}
