package sdkruntimesclient

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"strconv"

	"connectrpc.com/connect"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	runtimesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1/runtimesv1connect"
	"go.autokitteh.dev/autokitteh/sdk/internal/rpcerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkbuildfile"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var bidi, _ = strconv.ParseBool(os.Getenv("BIDI_RUN"))

type client struct {
	client runtimesv1connect.RuntimesServiceClient
	sl     *zap.SugaredLogger
}

func New(p sdkclient.Params) sdkservices.Runtimes {
	return &client{client: internal.New(runtimesv1connect.NewRuntimesServiceClient, p), sl: p.Safe().L.Sugar()}
}

func (c *client) Run(
	ctx context.Context,
	rid sdktypes.RunID,
	path string,
	build *sdkbuildfile.BuildFile,
	globals map[string]sdktypes.Value,
	cbs *sdkservices.RunCallbacks,
) (sdkservices.Run, error) {
	run := c.run

	if bidi {
		run = c.bidiRun
	}

	return run(ctx, rid, path, build, globals, cbs)
}

func (c *client) Build(ctx context.Context, fs fs.FS, symbols []sdktypes.Symbol, memo map[string]string) (*sdkbuildfile.BuildFile, error) {
	resources, err := kittehs.FSToMap(fs)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Build(ctx, connect.NewRequest(&runtimesv1.BuildRequest{
		Resources: resources,
		Symbols:   kittehs.TransformToStrings(symbols),
		Memo:      memo,
	}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	if resp.Msg.Error != nil {
		perr, err := sdktypes.ProgramErrorFromProto(resp.Msg.Error)
		if err != nil {
			return nil, fmt.Errorf("invalid error: %w", err)
		}
		return nil, perr.ToError()
	}

	return sdkbuildfile.Read(bytes.NewReader(resp.Msg.BuildFile))
}

func (c *client) List(ctx context.Context) ([]sdktypes.Runtime, error) {
	resp, err := c.client.List(ctx, connect.NewRequest(&runtimesv1.ListRequest{}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return kittehs.TransformError(resp.Msg.Runtimes, sdktypes.StrictRuntimeFromProto)
}

func (c *client) New(ctx context.Context, name sdktypes.Symbol) (sdkservices.Runtime, error) {
	resp, err := c.client.Describe(ctx, connect.NewRequest(&runtimesv1.DescribeRequest{Name: name.String()}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	if resp.Msg.Runtime == nil {
		return nil, nil
	}

	desc, err := sdktypes.StrictRuntimeFromProto(resp.Msg.Runtime)
	if err != nil {
		return nil, fmt.Errorf("invalid runtime: %w", err)
	}

	return &runtime{desc: desc, client: c}, nil
}
