package confluence

import (
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	authType = sdktypes.NewSymbol("authType")

	baseURL = sdktypes.NewSymbol("BaseURL")
	token   = sdktypes.NewSymbol("Token")
	email   = sdktypes.NewSymbol("Email")

	accessID        = sdktypes.NewSymbol("AccessID")
	accessURL       = sdktypes.NewSymbol("AccessURL")
	accessName      = sdktypes.NewSymbol("AccessName")
	accessScope     = sdktypes.NewSymbol("AccessScope")
	accessAvatarURL = sdktypes.NewSymbol("AccessAvatarURL")
)

func webhookID(category string) sdktypes.Symbol {
	return sdktypes.NewSymbol("WebhookID_" + category)
}

func webhookSecret(category string) sdktypes.Symbol {
	return sdktypes.NewSymbol("WebhookSecret_" + category)
}
