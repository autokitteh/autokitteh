package common

import (
	"go.autokitteh.dev/autokitteh/sdk/sdkclients"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

var client sdkservices.Services

func InitRPCClient(url string, authToken string) {
	client = sdkclients.New(sdkclient.Params{URL: url, AuthToken: authToken})
}

func Client() sdkservices.Services { return client }
