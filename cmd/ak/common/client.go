package common

import (
	"go.autokitteh.dev/autokitteh/sdk/sdkclients"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

var client sdkservices.Services

func InitRPCClient(authToken string) (err error) {
	if authToken == "" {
		if authToken, err = GetToken(); err != nil {
			return
		}
	}

	client = sdkclients.New(sdkclient.Params{URL: serverURL.String(), AuthToken: authToken})
	return
}

func Client() sdkservices.Services { return client }
