package forms

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/forms/v1"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	"go.autokitteh.dev/autokitteh/integrations/google/connections"
	"go.autokitteh.dev/autokitteh/integrations/google/vars"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type api struct {
	vars sdkservices.Vars
	cid  sdktypes.ConnectionID
}

var IntegrationID = sdktypes.NewIntegrationIDFromName("googleforms")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: IntegrationID.String(),
	UniqueName:    "googleforms",
	DisplayName:   "Google Forms",
	Description:   "Google Forms is a survey administration software that part of the Google Workspace office suite.",
	LogoUrl:       "/static/images/google_forms.svg",
	ConnectionUrl: "/googleforms/connect",
	ConnectionCapabilities: &sdktypes.ConnectionCapabilitiesPB{
		RequiresConnectionInit: true,
	},
}))

const (
	// pubsubTopicEnvVar is the name of an environment variable that
	// contains the GCP Pub/Sub topic name for Google Forms notifications.
	pubsubTopicEnvVar = "GOOGLE_FORMS_PUBSUB_TOPIC"
)

func New(cvars sdkservices.Vars) sdkservices.Integration {
	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New( /* No exported functions for Starlark */ ),
		connections.ConnStatus(cvars),
		connections.ConnTest(cvars),
		sdkintegrations.WithConnectionConfigFromVars(cvars),
	)
}

// Extract the form ID from the connection's vars.
// Return an empty string if the form ID wasn't set (i.e. do nothing).
func (a api) formID(ctx context.Context) (string, error) {
	data, err := a.connectionData(ctx)
	if err != nil {
		return "", err
	}

	return data.FormID, nil
}

func (a api) formsIDAndClient(ctx context.Context) (string, *forms.Service, error) {
	id, err := a.formID(ctx)
	if err != nil {
		return "", nil, err
	}

	client, err := a.formsClient(ctx)
	if err != nil {
		return "", nil, err
	}

	return id, client, nil
}

func (a api) formsClient(ctx context.Context) (*forms.Service, error) {
	data, err := a.connectionData(ctx)
	if err != nil {
		return nil, err
	}

	var src oauth2.TokenSource
	if data.OAuthData != "" {
		if src, err = oauthTokenSource(ctx, data.OAuthData); err != nil {
			return nil, err
		}
	} else {
		src, err = jwtTokenSource(ctx, data.JSON)
		if err != nil {
			return nil, err
		}
	}

	svc, err := forms.NewService(ctx, option.WithTokenSource(src))
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func (a api) connectionData(ctx context.Context) (*vars.Vars, error) {
	cid, err := sdkmodule.FunctionConnectionIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if !cid.IsValid() {
		cid = a.cid // Fallback during authentication flows.
	}

	vs, err := a.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
	if err != nil {
		return nil, err
	}

	var decoded vars.Vars
	vs.Decode(&decoded)
	return &decoded, nil
}

func oauthTokenSource(ctx context.Context, data string) (oauth2.TokenSource, error) {
	d, err := sdkintegrations.DecodeOAuthData(data)
	if err != nil {
		return nil, err
	}

	return oauthConfig().TokenSource(ctx, d.Token), nil
}

// TODO(ENG-112): Use OAuth().Get() instead of calling this function.
func oauthConfig() *oauth2.Config {
	addr := os.Getenv("WEBHOOK_ADDRESS")
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
		RedirectURL:  fmt.Sprintf("https://%s/oauth/redirect/google", addr),
		// https://developers.google.com/identity/protocols/oauth2/scopes#script
		Scopes: []string{
			// Non-sensitive.
			googleoauth2.OpenIDScope,
			googleoauth2.UserinfoEmailScope,
			googleoauth2.UserinfoProfileScope,
			// Sensitive.
			forms.FormsBodyScope,
			forms.FormsResponsesReadonlyScope,
		},
	}
}

func jwtTokenSource(ctx context.Context, data string) (oauth2.TokenSource, error) {
	scopes := oauthConfig().Scopes

	cfg, err := google.JWTConfigFromJSON([]byte(data), scopes...)
	if err != nil {
		return nil, err
	}

	return cfg.TokenSource(ctx), nil
}

func (a api) getForm(ctx context.Context) (*forms.Form, error) {
	formID, client, err := a.formsIDAndClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := client.Forms.Get(formID).Do()
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (a api) listResponses(ctx context.Context) ([]*forms.FormResponse, error) {
	formID, client, err := a.formsIDAndClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := client.Forms.Responses.List(formID).Do()
	if err != nil {
		return nil, err
	}

	return resp.Responses, nil
}

// To receive notifications, the topic must grant publish privileges to the
// Forms service account `forms-notifications@system.gserviceaccount.com`.
// Only the GCP project that owns a topic may create a watch with it.
// Pub/Sub delivery guarantees should be considered.
func (a api) watchesCreate(ctx context.Context, e WatchEventType) (*forms.Watch, error) {
	formID, client, err := a.formsIDAndClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := client.Forms.Watches.Create(formID, &forms.CreateWatchRequest{
		Watch: &forms.Watch{
			EventType: string(e),
			Target: &forms.WatchTarget{
				Topic: &forms.CloudPubsubTopic{TopicName: os.Getenv(pubsubTopicEnvVar)},
			},
		},
	}).Do()
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (a api) watchesList(ctx context.Context) ([]*forms.Watch, error) {
	formID, client, err := a.formsIDAndClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := client.Forms.Watches.List(formID).Do()
	if err != nil {
		return nil, err
	}

	return resp.Watches, nil
}

func (a api) watchesRenew(ctx context.Context, watchID string) (*forms.Watch, error) {
	formID, client, err := a.formsIDAndClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := client.Forms.Watches.Renew(formID, watchID, &forms.RenewWatchRequest{}).Do()
	if err != nil {
		return nil, err
	}

	return resp, nil
}
