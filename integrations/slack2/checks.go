package slack2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/slack/api"
	"go.autokitteh.dev/autokitteh/integrations/slack2/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// status checks the connection's initialization status (is it
// initialized? what type of authentication is configured?). This
// ensures that the connection is at least theoretically usable.
func status(v sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		vs, errStatus, err := common.ReadConnectionVars(ctx, v, cid)
		if errStatus.IsValid() || err != nil {
			return errStatus, err
		}

		switch common.ReadAuthType(vs) {
		case "":
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		// TODO(INT-267): Remove [integrations.OAuth] once the web UI is migrated too.
		case integrations.OAuth, integrations.OAuthDefault, integrations.OAuthPrivate:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using OAuth 2.0"), nil
		case integrations.SocketMode:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using Socket Mode"), nil
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}
	})
}

// test checks whether the connection is actually usable, i.e. the configured
// authentication credentials are valid and can be used to make API calls.
func test(v sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		vs, errStatus, err := common.ReadConnectionVars(ctx, v, cid)
		if errStatus.IsValid() || err != nil {
			return errStatus, err
		}

		var token string
		switch common.ReadAuthType(vs) {
		case "":
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		// TODO(INT-267): Remove [integrations.OAuth] once the web UI is migrated too.
		case integrations.OAuth, integrations.OAuthDefault, integrations.OAuthPrivate:
			token = vs.GetValue(common.OAuthAccessTokenVar)
		case integrations.SocketMode:
			token = vs.GetValue(vars.BotTokenVar)
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}

		if _, err = testBotToken(ctx, token); err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, ""), nil
	})
}

// testBotToken checks a bot token's authentication & identity.
// Based on: https://api.slack.com/methods/auth.test (no scopes required).
func testBotToken(ctx context.Context, botToken string) (*authTestResponse, error) {
	ctx = context.WithValue(ctx, api.OAuthTokenContextKey, botToken)
	resp := &authTestResponse{}
	err := api.PostJSON(ctx, nil, struct{}{}, resp, "auth.test")
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}
	return resp, nil
}

type authTestResponse struct {
	api.SlackResponse

	URL                 string `json:"url"`
	Team                string `json:"team"`
	User                string `json:"user"`
	TeamID              string `json:"team_id"`
	UserID              string `json:"user_id"`
	BotID               string `json:"bot_id"`
	EnterpriseID        string `json:"enterprise_id"`
	IsEnterpriseInstall bool   `json:"is_enterprise_install"`
}

// getBotInfo gets information about a bot user.
// Based on: https://api.slack.com/methods/bots.info
// Required Slack app scope: https://api.slack.com/scopes/users:read
func getBotInfo(ctx context.Context, botToken string, authTest *authTestResponse) (*botInfo, error) {
	// Construct HTTP GET request.
	u := "https://slack.com/api/bots.info"
	u = fmt.Sprintf("%s?bot=%s&team_id=%s", u, authTest.BotID, authTest.TeamID)
	if authTest.IsEnterpriseInstall {
		u = fmt.Sprintf("%s&enterprise_id=%s", u, authTest.EnterpriseID)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+botToken)

	// Send request to server.
	c := &http.Client{Timeout: api.Timeout}
	httpResp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	// Parse HTTP response.
	if httpResp.StatusCode != http.StatusOK {
		return nil, errors.New(httpResp.Status)
	}
	b, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	jsonResp := &botInfoResponse{}
	if err := json.Unmarshal(b, jsonResp); err != nil {
		return nil, err
	}

	if !jsonResp.OK {
		return nil, errors.New(jsonResp.Error)
	}
	if jsonResp.Bot.AppID == "" {
		return nil, errors.New("empty response")
	}
	return &jsonResp.Bot, nil
}

type botInfoResponse struct {
	api.SlackResponse

	Bot botInfo `json:"bot"`
}

type botInfo struct {
	ID      string            `json:"id"`
	Deleted bool              `json:"deleted"`
	Name    string            `json:"name"`
	Updated int               `json:"updated"`
	AppID   string            `json:"app_id"`
	UserID  string            `json:"user_id"`
	Icons   map[string]string `json:"icons"`
}

// tempWebSocketURL generates a temporary WebSocket URL for a Socket Mode app.
// Based on: https://api.slack.com/methods/apps.connections.open
// Required Slack app scope: https://api.slack.com/scopes/connections:write
func tempWebSocketURL(ctx context.Context, appToken string) (*openConnectionResponse, error) {
	ctx = context.WithValue(ctx, api.OAuthTokenContextKey, appToken)
	resp := &openConnectionResponse{}
	err := api.PostJSON(ctx, nil, struct{}{}, resp, "apps.connections.open")
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}
	return resp, nil
}

type openConnectionResponse struct {
	api.SlackResponse

	URL string `json:"url"`
}

// encodeInstallInfo encodes a Slack app's installation into AutoKitteh connection variables.
func encodeInstallInfo(auth *authTestResponse, bot *botInfo) vars.InstallInfo {
	return vars.InstallInfo{
		EnterpriseID: auth.EnterpriseID,
		Team:         auth.Team,
		TeamID:       auth.TeamID,
		User:         auth.User,
		UserID:       auth.UserID,

		BotName:    bot.Name,
		BotID:      bot.ID,
		BotUpdated: time.Unix(int64(bot.Updated), 0).UTC().Format(time.RFC3339),
		AppID:      bot.AppID,

		InstallIDs: vars.InstallIDs(bot.AppID, auth.EnterpriseID, auth.TeamID),
	}
}
