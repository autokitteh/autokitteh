package sdkclient

import (
	"net/http"

	"connectrpc.com/connect"
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
}

func (p Params) Safe() Params {
	if p.HTTPClient == nil {
		p.HTTPClient = &http.Client{}
	}

	return p
}
