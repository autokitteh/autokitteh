package vars

import (
	"fmt"
	"strings"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	UserAppID     = varKey("app_id")
	UserInstallID = varKey("install_id")

	PATKey    = sdktypes.NewSymbol("pat_key")
	PATSecret = sdktypes.NewSymbol("pat_secret")
	PAT       = sdktypes.NewSymbol("pat")
	PATUser   = sdktypes.NewSymbol("pat_user")
)

func varKey(kind string) func(string) sdktypes.Symbol {
	return func(user string) sdktypes.Symbol {
		return sdktypes.NewSymbol(fmt.Sprintf("%s__%s", kind, encodeUser(user)))
	}
}

func encodeUser(user string) string {
	// Username may only contain alphanumeric characters or single hyphens, and cannot begin or end with a hyphen.
	return strings.ReplaceAll(user, "-", "_")
}

func InstallKey(appID, installID string) sdktypes.Symbol {
	return sdktypes.NewSymbol(fmt.Sprintf("install_key__%s__%s", appID, installID))
}
