package internal

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
)

func newClientAuthInterceptor(token string) connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(
			func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
				if !req.Spec().IsClient {
					msg := "client auth interceptor used in server"
					sdklogger.DPanic(msg)
					return nil, errors.New(msg)
				}

				req.Header().Set(sdkclient.AuthorizationHeader, token)

				return next(ctx, req)
			},
		)
	}

	return connect.UnaryInterceptorFunc(interceptor)
}
