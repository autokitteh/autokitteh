package jira

import (
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	baseURL = sdktypes.NewSymbol("BaseURL")
	token   = sdktypes.NewSymbol("Token")
	email   = sdktypes.NewSymbol("Email")

	oauthAccessToken = sdktypes.NewSymbol("oauth_AccessToken")
	accessID         = sdktypes.NewSymbol("access_ID")
	accessURL        = sdktypes.NewSymbol("access_URL")
	accessName       = sdktypes.NewSymbol("access_Name")
	accessScope      = sdktypes.NewSymbol("access_Scope")
	accessAvatarURL  = sdktypes.NewSymbol("access_AvatarURL")

	webhookID         = sdktypes.NewSymbol("WebhookID")
	webhookExpiration = sdktypes.NewSymbol("WebhookExpiration")
)
