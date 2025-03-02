// Package common provides common utilities for integrations.
package common

import (
	"embed"
	"fmt"
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func Descriptor(uniqueName, displayName, logoURL string) sdktypes.Integration {
	return kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
		IntegrationId: sdktypes.NewIntegrationIDFromName(uniqueName).String(),
		UniqueName:    uniqueName,
		DisplayName:   displayName,
		LogoUrl:       logoURL,
		ConnectionUrl: "/" + uniqueName,
		ConnectionCapabilities: &sdktypes.ConnectionCapabilitiesPB{
			RequiresConnectionInit: true,
			SupportsConnectionTest: true,
		},
	}))
}

// ServeStaticUI registers an integration's static web content to
// AutoKitteh's internal user-authenticated HTTP server. User auth
// isn't actually required for serving these files, but it's impossible
// to use them (i.e. to initialize a connection) without authentication.
func ServeStaticUI(m *muxes.Muxes, i sdktypes.Integration, fs embed.FS) {
	pattern := fmt.Sprintf("GET %s/", i.ConnectionURL().Path)
	m.Auth.Handle(pattern, http.FileServer(http.FS(fs)))
}

// RegisterSaveHandler registers a webhook to handle saving connection variables.
// This is always the first step in initializing a connection, regardless of its
// auth type. It's also the last step for non-OAuth authentication types. The
// handler function requires an authenticated user context for database access.
func RegisterSaveHandler(m *muxes.Muxes, i sdktypes.Integration, h http.HandlerFunc) {
	pattern := fmt.Sprintf(" %s/save", i.ConnectionURL().Path)
	m.Auth.HandleFunc(http.MethodGet+pattern, h)
	m.Auth.HandleFunc(http.MethodPost+pattern, h)
}

// RegisterOAuthHandler registers a webhook to handle the last step in a 3-legged
// OAuth 2.0 flow. This is the only step that requires an authenticated user context
// for database access. The handler function receives an incoming redirect request
// from AutoKitteh's generic OAuth service, which contains an OAuth token (if the
// OAuth flow was successful) and form parameters for debugging and validation.
func RegisterOAuthHandler(m *muxes.Muxes, i sdktypes.Integration, h http.HandlerFunc) {
	m.Auth.HandleFunc(fmt.Sprintf("GET %s/oauth", i.ConnectionURL().Path), h)
}
