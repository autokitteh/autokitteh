package internal

import (
	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
)

func New[
	T any,
	F func(connect.HTTPClient, string, ...connect.ClientOption) T,
](
	f F,
	p sdkclient.Params,
) T {
	p = p.Safe()

	if p.AuthToken != "" {
		p.Options = append(
			p.Options,
			connect.WithInterceptors(
				newClientAuthUnaryInterceptor(p.AuthToken),
			),
		)
	}

	return f(p.HTTPClient, p.URL, p.Options...)
}
