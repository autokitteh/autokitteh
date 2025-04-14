package vars

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	AuthType = sdktypes.NewSymbol("auth_type")

	// GitHub app (OAuth)
	AppID     = sdktypes.NewSymbol("app_id")
	AppName   = sdktypes.NewSymbol("app_name")
	InstallID = sdktypes.NewSymbol("install_id")

	TargetID   = sdktypes.NewSymbol("target_id")
	TargetName = sdktypes.NewSymbol("target_name")
	TargetType = sdktypes.NewSymbol("target_type")

	RepoSelection = sdktypes.NewSymbol("repositories")
	Permissions   = sdktypes.NewSymbol("permissions")
	Events        = sdktypes.NewSymbol("events")
	UpdatedAt     = sdktypes.NewSymbol("updated_at")

	// PAT + webhook
	PATKey    = sdktypes.NewSymbol("pat_key")
	PATSecret = sdktypes.NewSymbol("pat_secret")
	PAT       = sdktypes.NewSymbol("pat")
	PATUser   = sdktypes.NewSymbol("pat_user")

	// Custom OAuth
	ClientID      = sdktypes.NewSymbol("client_id")
	ClientSecret  = sdktypes.NewSymbol("client_secret")
	WebhookSecret = sdktypes.NewSymbol("webhook_secret")
	EnterpriseURL = sdktypes.NewSymbol("enterprise_url")
	PrivateKey    = sdktypes.NewSymbol("private_key")
)

func InstallKey(appID, installID string) sdktypes.Symbol {
	return sdktypes.NewSymbol(fmt.Sprintf("install_key__%s__%s", appID, installID))
}
