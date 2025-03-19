package zoom

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/oauth"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// status checks the connection's initialization status (is it
// initialized? what type of authentication is configured?). This
// ensures that the connection is at least theoretically usable.
func status(v sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		vs, errStatus, err := common.ReadVarsWithStatus(ctx, v, cid)
		if errStatus.IsValid() || err != nil {
			return errStatus, err
		}

		switch common.ReadAuthType(vs) {
		case "":
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		case integrations.OAuthDefault, integrations.OAuthPrivate:
			return common.CheckOAuthToken(vs)
		case integrations.ServerToServer:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using S2S"), nil
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}
	})
}

// test checks whether the connection is actually usable, i.e. the configured
// authentication credentials are valid and can be used to make API calls.
func test(v sdkservices.Vars, o *oauth.OAuth) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		vs, errStatus, err := common.ReadVarsWithStatus(ctx, v, cid)
		if errStatus.IsValid() || err != nil {
			return errStatus, err
		}

		switch common.ReadAuthType(vs) {
		case "":
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		case integrations.OAuthDefault, integrations.OAuthPrivate:
			return oauthTest(ctx, o, vs)
		case integrations.ServerToServer:
			return serverToServerTest(ctx, vs)
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}
	})
}

// oauthTest verifies the OAuth authentication with Zoom API using the "me" context.
// (Based on: https://developers.zoom.us/docs/integrations/oauth/#the-me-context).
func oauthTest(ctx context.Context, o *oauth.OAuth, vs sdktypes.Vars) (sdktypes.Status, error) {
	// TODO(INT-338): Support private OAuth.
	// Check if the token is expired and refresh it.
	l := zap.L().With(zap.String("integration", desc.UniqueName().String()))
	token := o.FreshToken(ctx, l, desc, vs)

	url := "https://api.zoom.us/v2/users/me"
	auth := "Bearer " + token.AccessToken
	if _, err := common.HTTPGet(ctx, url, auth); err != nil {
		return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
	}

	return sdktypes.NewStatus(sdktypes.StatusCodeOK, "OAuth connection successful"), nil
}

// serverToServerTest validates the Server-to-Server authentication test for Zoom.
// (Based on: https://developers.zoom.us/docs/internal-apps/s2s-oauth/).
func serverToServerTest(ctx context.Context, vs sdktypes.Vars) (sdktypes.Status, error) {
	var app privateApp
	vs.Decode(&app)

	// Get a server-to-server token.
	token, err := serverToken(ctx, app)
	if err != nil {
		return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
	}

	url := "https://api.zoom.us/v2/users"
	auth := "Bearer " + token
	_, err = common.HTTPGet(ctx, url, auth)
	if err != nil {
		return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
	}

	return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Server-to-Server connection successful"), nil
}

// httpPost sends an HTTP POST request with a URL-encoded body and basic authentication.
func httpPost(ctx context.Context, data url.Values, clientID, clientSecret, url string) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	if clientID != "" && clientSecret != "" {
		req.SetBasicAuth(clientID, clientSecret)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}

	return resp, nil
}
