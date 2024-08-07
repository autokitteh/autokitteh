package webhooks

import (
	"context"
	"net/http"
	"strings"

	"github.com/twilio/twilio-go"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	// AuthPath is the URL path for our webhook to save a new autokitteh
	// connection, after the user submits their Twilio secrets.
	AuthPath = "/twilio/save"

	headerContentType = "Content-Type"
	contentTypeForm   = "application/x-www-form-urlencoded"
)

type Vars struct {
	AccountSID string
	Username   string `vars:"secret"`
	Password   string `vars:"secret"`
}
type TwilioTest struct {
	Vars Vars
}

// HandleAuth saves a new autokitteh connection with user-submitted Twilio secrets.
func (h handler) HandleAuth(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, h.integration)

	// Check "Content-Type" header.
	contentType := r.Header.Get(headerContentType)
	if !strings.HasPrefix(contentType, contentTypeForm) {
		c.AbortBadRequest("unexpected content type")
		return
	}

	// Read and parse POST request body.
	if err := r.ParseForm(); err != nil {
		l.Warn("Failed to parse incoming HTTP request", zap.Error(err))
		c.AbortBadRequest("form parsing error")
		return
	}

	accountSID := r.Form.Get("account_sid")
	username := accountSID
	password := r.Form.Get("auth_token")
	if password == "" {
		username = r.Form.Get("api_key")
		password = r.Form.Get("api_secret")
	}

	// TODO(ENG-1156): Test the authentication details.
	api := TwilioTest{Vars: Vars{
		AccountSID: accountSID,
		Username:   username,
		Password:   password,
	}}
	ctx := context.Background()

	_, err := api.Test(ctx)
	if err != nil {
		l.Warn("Twilio authentication test failed", zap.Error(err))
		c.AbortBadRequest("Twilio authentication test failed: " + err.Error())
		return
	}

	c.Finalize(sdktypes.EncodeVars(Vars{
		AccountSID: accountSID,
		Username:   username,
		Password:   password,
	}))
}

type TestResponse struct {
	AccountSID   string `json:"AccountSID"`
	FriendlyName string `json:"FriendlyName"`
	Status       string `json:"Status"`
}

func (a TwilioTest) Test(ctx context.Context) (*TestResponse, error) {
	// Create a new Twilio client using the provided username and password
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: a.Vars.Username,
		Password: a.Vars.Password,
	})

	// Fetch the account details
	resp, err := client.Api.FetchAccount(a.Vars.AccountSID)
	if err != nil {
		return nil, err
	}

	// Parse and return the response
	return &TestResponse{
		AccountSID:   *resp.Sid,
		FriendlyName: *resp.FriendlyName,
		Status:       *resp.Status,
	}, nil
}
