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

	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/integrations/slack/internal/vars"
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

type OAuthTokenContextKey struct{}

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
	// This is usually a struct, but *may* be a string: when the OAuth redirect
	// webhook tests the connection, it doesn't have an AK connection token yet,
	// but it still needs to use its new OAuth token, so we pass it as the body.
	// See the underlying post() method below for how it handles this case.
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

	// Convert the autokitteh connection token into an OAuth user access token.
	oauthToken, err := getConnection(ctx, vars)
	if err != nil {
		return nil, err
	}
	if oauthToken.AccessToken == "" {
		// Special cases where we don't have/need a connection token:
		// 1. OAuth redirect handler is testing a new OAuth token, before
		//    creating a new AK connection token (so it passes the OAuth
		//    user access token as a fake request body up to this point)
		var ok bool
		oauthToken.AccessToken, ok = ctx.Value(OAuthTokenContextKey{}).(string)
		if !ok {
			l.Warn("Unexpected non-string OAuth access token after OAuth exchange")
		}
		// 2. Response to a user interaction webhook (with a JSON body)
		// 3. Unit tests (body is an empty string)
		// --> No need to do anything in these cases.
	}

	// Construct HTTP POST request.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		l.Error("Failed to construct HTTP request",
			zap.Error(err),
			zap.String("httpMethod", http.MethodPost),
			zap.String("url", url),
			zap.String("body", body),
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

// getConnection calls the Get method in SecretsService.
func getConnection(ctx context.Context, varsSvc sdkservices.Vars) (*oauth2.Token, error) {
	if varsSvc == nil {
		// test.
		return &oauth2.Token{}, nil
	}

	// Extract the connection token from the given context.
	cid, err := sdkmodule.FunctionConnectionIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	vs, err := varsSvc.Get(context.Background(), sdktypes.NewVarScopeID(cid))
	if err != nil {
		return nil, err
	}

	// Socket mode connection.
	if bt := vs.GetValue(vars.BotTokenName); bt != "" {
		return &oauth2.Token{
			AccessToken: bt,
		}, nil
	}

	// OAuth connection.
	oauthData, err := sdkintegrations.DecodeOAuthData(vs.GetValue(vars.OAuthDataName))
	if err != nil {
		return nil, err
	}

	return oauthData.Token, nil
}
