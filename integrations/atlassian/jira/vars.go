package jira

import (
	"fmt"

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

	WebhookKeySymbol  = sdktypes.NewSymbol("WebhookKey")
	WebhookID         = sdktypes.NewSymbol("WebhookID")
	WebhookExpiration = sdktypes.NewSymbol("WebhookExpiration")
)

// webhookKey combines a Jira domain with its webhook ID to create a unique identifier.
// This is used to find all the relevant AutoKitteh connections for an incoming event.
func webhookKey(domain, webhookID string) string {
	return fmt.Sprintf("%s/%s", domain, webhookID)
}
