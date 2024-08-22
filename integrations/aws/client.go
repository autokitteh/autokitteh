package aws

import (
	"context"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rdsdata"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type integration struct {
	vars sdkservices.Vars
}

var (
	svcs = []struct {
		name string
		fn   any
	}{
		{"ec2", ec2.NewFromConfig},
		{"eventbridge", eventbridge.NewFromConfig},
		{"iam", iam.NewFromConfig},
		{"rds", rds.NewFromConfig},
		{"rdsdata", rdsdata.NewFromConfig},
		{"s3", s3.NewFromConfig},
		{"sns", sns.NewFromConfig},
		{"sqs", sqs.NewFromConfig},
	}

	useDefaultConfig, _ = strconv.ParseBool(os.Getenv("AWS_USE_DEFAULT_CONFIG"))

	defaultAWSConfig *aws.Config

	authType = sdktypes.NewSymbol("authType")
)

func init() {
	initDefaultConfig()
}

func initDefaultConfig() {
	if !useDefaultConfig {
		return
	}

	sdklogger.Warn("aws: using default global config")

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		sdklogger.Panic(err)
		return
	}

	defaultAWSConfig = &cfg
}

func initOpts(vars sdkservices.Vars) (opts []sdkmodule.Optfn) {
	for _, svc := range svcs {
		opts = append(opts, kittehs.Must1(importServiceMethods(vars, svc.name, svc.fn))...)
	}
	return
}

var integrationID = sdktypes.NewIntegrationIDFromName("aws")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: integrationID.String(),
	UniqueName:    "aws",
	DisplayName:   "AWS (All APIs)",
	Description:   "Aggregation of all available Amazon Web Services (AWS) APIs.",
	LogoUrl:       "/static/images/aws.svg",
	UserLinks: map[string]string{
		"1 API documentation": "https://docs.aws.amazon.com/",
		"2 Service console":   "https://console.aws.amazon.com/",
	},
	ConnectionUrl: "/aws/connect",
}))

func New(cvars sdkservices.Vars) sdkservices.Integration {
	i := &integration{vars: cvars}
	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New(initOpts(cvars)...),
		connStatus(i),
		sdkintegrations.WithConnectionConfigFromVars(cvars),
	)
}

// connStatus is an optional connection status check provided by the
// integration with AutoKitteh. The possible results are "init required" 
// (indicating the connection is not yet usable), "using X" (indicating 
// one of multiple available authentication methods is in use), or 
// "initialized" when only one authentication method is available.
func connStatus(i *integration) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "init required"), nil
		}

		vs, err := i.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			return sdktypes.InvalidStatus, err
		}

		at := vs.Get(authType)
		if !at.IsValid() || at.Value() == "" {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "init required"), nil
		}

		// Align with:
		// https://github.com/autokitteh/web-platform/blob/main/src/enums/connections/connectionTypes.enum.ts
		switch at.Value() {
		case "awsConfig":
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "initialized"), nil
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "bad auth type"), nil
		}
	})
}
