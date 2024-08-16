package sdkclient

import (
	"net/http"

	"connectrpc.com/connect"
	"go.uber.org/zap"
)

const (
	DefaultPort = "9980"

	DefaultLocalURL = "http://localhost:" + DefaultPort
	DefaultCloudURL = "https://api.autokitteh.cloud"
)

const (
	AuthorizationHeader = "Authorization"
)

type Params struct {
	HTTPClient *http.Client
	URL        string
	Options    []connect.ClientOption
	AuthToken  string
	L          *zap.Logger
}

func (p Params) Safe() Params {
	if p.HTTPClient == nil {
		p.HTTPClient = http.DefaultClient
	}

	if p.L == nil {
		p.L = zap.NewNop()
	}

	return p
}
