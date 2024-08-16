package internal

import (
	"crypto/tls"
	"net"
	"net/http"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"

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
		p.Options = append(p.Options, connect.WithInterceptors(newClientAuthInterceptor(p.AuthToken)))
	}

	httpClient := p.HTTPClient

	if httpClient == nil || http.DefaultClient == httpClient {
		// see https://connectrpc.com/docs/go/deployment#h2c.

		httpClient = &http.Client{
			Transport: &http2.Transport{
				AllowHTTP: true,
				DialTLS: func(network, addr string, _ *tls.Config) (net.Conn, error) {
					// If you're also using this client for non-h2c traffic, you may want
					// to delegate to tls.Dial if the network isn't TCP or the addr isn't
					// in an allowlist.
					return net.Dial(network, addr)
				},
				// TODO: Don't forget timeouts!
			},
		}
	}

	return f(httpClient, p.URL, p.Options...)
}
