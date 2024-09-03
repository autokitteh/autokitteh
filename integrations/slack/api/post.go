package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/integrations/slack/internal/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	HeaderContentType    = "Content-Type"
	HeaderSlackTimestamp = "X-Slack-Request-Timestamp"
	HeaderSlackSignature = "X-Slack-Signature"
	HeaderAuthorization  = "Authorization"

	ContentTypeForm            = "application/x-www-form-urlencoded"
	ContentTypeJSONCharsetUTF8 = "application/json; charset=utf-8"
	ContentTypeJSON            = "application/json"

	// Timeout for short-lived outbound HTTP POST requests.
	Timeout = 3 * time.Second
)

// slackURL is a var and not a const for unit-testing purposes.
var slackURL = "https://slack.com/api/"

type ctxKey string

var OAuthTokenContextKey = ctxKey("OAuthTokenContext")

// PostForm sends a short-lived HTTP POST request with an OAuth bearer token and
// URL-encoded key/value payload, and then receives and parses the JSON response.
func PostForm(ctx context.Context, vars sdkservices.Vars, kv url.Values, resp any, slackMethod string) error {
	l := extrazap.ExtractLoggerFromContext(ctx).With(
		zap.String("httpContent", "form"),
		zap.String("slackMethod", slackMethod),
	)
	ctx = extrazap.AttachLoggerToContext(l, ctx)

	// Construct the request URL.
	u, err := url.JoinPath(slackURL, slackMethod)
	if err != nil {
		l.Error("Failed to construct Slack API URL",
			zap.Error(err),
			zap.String("base", slackURL),
		)
		return err
	}

	// Send an HTTP POST request with the URL-encoded payload.
	body, err := post(ctx, vars, u, kv.Encode(), ContentTypeForm)
	if err != nil {
		return err
	}

	// Parse and return the JSON in the HTTP response.
	if err := json.Unmarshal(body, resp); err != nil {
		l.Error("Failed to parse JSON payload",
			zap.ByteString("json", body),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// PostJSON sends a short-lived HTTP POST request with an OAuth bearer token
// and JSON payload, and then receives and parses the JSON response.
func PostJSON(ctx context.Context, vars sdkservices.Vars, req, resp any, slackMethod string) error {
	l := extrazap.ExtractLoggerFromContext(ctx).With(
		zap.String("httpContent", "json"),
		zap.String("slackMethod", slackMethod),
	)
	ctx = extrazap.AttachLoggerToContext(l, ctx)

	// Construct the request URL.
	// If slackMethod is a full URL (a callback URL provided by a Slack event)
	// then obviously don't prepend the Slack API's URL base to it.
	u := slackMethod
	if !strings.HasPrefix(u, "https://") {
		var err error
		u, err = url.JoinPath(slackURL, slackMethod)
		if err != nil {
			l.Error("Failed to construct Slack API URL",
				zap.Error(err),
				zap.String("base", slackURL),
				zap.String("slackMethod", slackMethod),
			)
			return err
		}
	}

	// Construct the request body.
	b, err := json.Marshal(req)
	if err != nil {
		l.Error("Failed to serialize JSON payload",
			zap.Error(err),
			zap.Any("json", req),
		)
		return err
	}

	// Send an HTTP POST request with the JSON payload.
	body, err := post(ctx, vars, u, string(b), ContentTypeJSONCharsetUTF8)
	if err != nil {
		return err
	}

	// Parse and return the JSON in the HTTP response.
	if err := json.Unmarshal(body, resp); err != nil {
		l.Error("Failed to parse JSON payload",
			zap.ByteString("json", body),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func post(ctx context.Context, vars sdkservices.Vars, url, body, contentType string) ([]byte, error) {
	l := extrazap.ExtractLoggerFromContext(ctx)

	// Retrieve the AutoKitteh connection's OAuth user access token.
	oauthToken, err := getConnection(ctx, vars)
	if err != nil {
		return nil, err
	}

	// Special cases where we don't have/need it from the connection:
	// 1. Response to a user interaction webhook (no need for a token)
	// 2. Unit tests
	if oauthToken.AccessToken == "" {
		// 3. OAuth redirect handler is testing a new OAuth token,
		//    while initializing a new AK connection (so it passes
		//    the OAuth token via the context up to this point)
		var ok bool
		oauthToken.AccessToken, ok = ctx.Value(OAuthTokenContextKey).(string)
		if !ok {
			oauthToken.AccessToken = ""
		}
	}

	// Construct HTTP POST request.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		l.Error("Failed to construct HTTP request",
			zap.String("httpMethod", http.MethodPost),
			zap.String("url", url),
			zap.String("body", body),
			zap.Error(err),
		)
		return nil, err
	}
	req.Header.Add(HeaderContentType, contentType)
	if oauthToken.AccessToken != "" {
		req.Header.Add(HeaderAuthorization, "Bearer "+oauthToken.AccessToken)
	}

	// Send request to server.
	c := &http.Client{Timeout: Timeout}
	resp, err := c.Do(req)
	if err != nil {
		l.Error("Failed to send HTTP request",
			zap.Error(err),
		)
		return nil, err
	}
	defer resp.Body.Close()

	// Parse HTTP response.
	if resp.StatusCode != http.StatusOK {
		l.Error("Received unsuccessful HTTP response",
			zap.Int("code", resp.StatusCode),
			zap.String("status", resp.Status),
		)
		return nil, errors.New(resp.Status)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Error("Failed to read HTTP response body",
			zap.Error(err),
		)
		return nil, err
	}
	return b, nil
}

func getConnection(ctx context.Context, varsSvc sdkservices.Vars) (*oauth2.Token, error) {
	token, ok := ctx.Value(OAuthTokenContextKey).(string)
	if ok {
		return &oauth2.Token{AccessToken: token}, nil
	}

	if varsSvc == nil {
		return &oauth2.Token{}, nil
	}

	// Extract the connection ID from the given context.
	cid, err := sdkmodule.FunctionConnectionIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	vs, err := varsSvc.Get(ctx, sdktypes.NewVarScopeID(cid))
	if err != nil {
		return nil, err
	}

	// Socket mode connection.
	if bt := vs.GetValue(vars.BotTokenName); bt != "" {
		return &oauth2.Token{AccessToken: bt}, nil
	}

	// Not Socket mode, so maybe OAuth?
	oauthData, err := sdkintegrations.DecodeOAuthData(vs.GetValue(vars.OAuthDataName))
	if err != nil {
		// No, we are using a temporary URL which was provided by Slack
		// (https://hooks.slack.com/actions/...), so we don't have to attach
		// an AutoKitteh connection's OAuth token to our outgoing request.
		return &oauth2.Token{}, nil
	}

	// Yes, we are sending a regular Slack API request
	// on behalf of an OAuth-based AutoKitteh connection.
	return oauthData.Token, nil
}
