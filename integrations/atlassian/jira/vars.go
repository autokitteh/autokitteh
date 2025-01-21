package jira

import (
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	authType = sdktypes.NewSymbol("authType")

	baseURL = sdktypes.NewSymbol("BaseURL")
	token   = sdktypes.NewSymbol("Token")
	email   = sdktypes.NewSymbol("Email")

	AccessID        = sdktypes.NewSymbol("AccessID")
	accessURL       = sdktypes.NewSymbol("AccessURL")
	accessName      = sdktypes.NewSymbol("AccessName")
	accessScope     = sdktypes.NewSymbol("AccessScope")
	accessAvatarURL = sdktypes.NewSymbol("AccessAvatarURL")

	WebhookID         = sdktypes.NewSymbol("WebhookID")
	WebhookExpiration = sdktypes.NewSymbol("WebhookExpiration")
)
