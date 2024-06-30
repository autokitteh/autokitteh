package jira

import (
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	baseURL = sdktypes.NewSymbol("BaseURL")
	token   = sdktypes.NewSymbol("Token")
	email   = sdktypes.NewSymbol("Email")

	oauthAccessToken = sdktypes.NewSymbol("oauth_AccessToken")
	accessID         = sdktypes.NewSymbol("AccessID")
	accessURL        = sdktypes.NewSymbol("AccessURL")
	accessName       = sdktypes.NewSymbol("AccessName")
	accessScope      = sdktypes.NewSymbol("AccessScope")
	accessAvatarURL  = sdktypes.NewSymbol("AccessAvatarURL")

	webhookID         = sdktypes.NewSymbol("WebhookID")
	webhookExpiration = sdktypes.NewSymbol("WebhookExpiration")
)
