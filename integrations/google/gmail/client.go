package gmail

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
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

const (
	googleScope = "google"
)

type api struct {
	vars sdkservices.Vars
	cid  sdktypes.ConnectionID
}

var IntegrationID = sdktypes.NewIntegrationIDFromName("gmail")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: IntegrationID.String(),
	UniqueName:    "gmail",
	DisplayName:   "Gmail",
	Description:   "Gmail is an email service provided by Google.",
	LogoUrl:       "/static/images/gmail.svg",
	ConnectionUrl: "/gmail/connect",
	ConnectionCapabilities: &sdktypes.ConnectionCapabilitiesPB{
		RequiresConnectionInit: true,
	},
}))

func New(cvars sdkservices.Vars) sdkservices.Integration {
	scope := googleScope

	opts := ExportedFunctions(cvars, scope, false)

	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New(opts...),
		connections.ConnStatus(cvars),
		connections.ConnTest(cvars),
		sdkintegrations.WithConnectionConfigFromVars(cvars))
}

func ExportedFunctions(cvars sdkservices.Vars, scope string, prefix bool) []sdkmodule.Optfn {
	a := api{vars: cvars}
	return []sdkmodule.Optfn{
		// Users.
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "get_profile"),
			a.getProfile,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users/getProfile")),

		// Drafts.
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "drafts_create"),
			a.draftsCreate,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.drafts/create"),
			sdkmodule.WithArgs("raw")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "drafts_delete"),
			a.draftsDelete,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.drafts/delete"),
			sdkmodule.WithArgs("id")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "drafts_get"),
			a.draftsGet,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.drafts/get"),
			sdkmodule.WithArgs("id", "format?")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "drafts_list"),
			a.draftsList,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.drafts/list"),
			sdkmodule.WithArgs("max_results?", "page_token?", "q?", "include_spam_trash?")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "drafts_send"),
			a.draftsSend,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.drafts/send"),
			sdkmodule.WithArgs("raw")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "drafts_update"),
			a.draftsUpdate,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.drafts/update"),
			sdkmodule.WithArgs("id", "raw")),

		// History.
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "history_list"),
			a.historyList,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.history/list"),
			sdkmodule.WithArgs("start_history_id", "max_results?", "page_token?", "label_id?", "history_types?")),

		// Labels.
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "labels_create"),
			a.labelsCreate,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.labels/create"),
			sdkmodule.WithArgs("label")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "labels_delete"),
			a.labelsDelete,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.labels/delete"),
			sdkmodule.WithArgs("id")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "labels_get"),
			a.labelsGet,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.labels/get"),
			sdkmodule.WithArgs("id")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "labels_list"),
			a.labelsList,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.labels/list")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "labels_patch"),
			a.labelsPatch,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.labels/patch"),
			sdkmodule.WithArgs("id", "label")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "labels_update"),
			a.labelsUpdate,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.labels/update"),
			sdkmodule.WithArgs("id", "label")),

		// Messages.
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "messages_batch_modify"),
			a.messagesBatchModify,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.messages/batchModify"),
			sdkmodule.WithArgs("ids", "add_label_ids?", "remove_label_ids?")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "messages_get"),
			a.messagesGet,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.messages/get"),
			sdkmodule.WithArgs("id", "format?", "metadata_headers?")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "messages_import"),
			a.messagesImport,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.messages/import"),
			sdkmodule.WithArgs("raw", "internal_date_source?", "never_mark_spam?", "processForCalendar?", "deleted?")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "messages_insert"),
			a.messagesInsert,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.messages/insert"),
			sdkmodule.WithArgs("raw", "internal_date_source?", "deleted?")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "messages_list"),
			a.messagesList,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.messages/list"),
			sdkmodule.WithArgs("max_results?", "page_token?", "q?", "label_ids?", "include_spam_trash?")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "messages_modify"),
			a.messagesModify,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.messages/modify"),
			sdkmodule.WithArgs("id", "add_label_ids?", "remove_label_ids?")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "messages_send"),
			a.messagesSend,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.messages/send"),
			sdkmodule.WithArgs("raw")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "messages_trash"),
			a.messagesTrash,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.messages/trash"),
			sdkmodule.WithArgs("id")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "messages_untrash"),
			a.messagesUntrash,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.messages/untrash"),
			sdkmodule.WithArgs("id")),

		// Attachments.
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "messages_attachments_get"),
			a.messagesAttachmentsGet,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.messages.attachments/get"),
			sdkmodule.WithArgs("message_id", "id")),

		// TODO: Settings
		// https://developers.google.com/gmail/api/reference/rest/v1/users.settings

		// Threads.
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "threads_get"),
			a.threadsGet,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.threads/get"),
			sdkmodule.WithArgs("id", "format?", "metadata_headers?")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "threads_list"),
			a.threadsList,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.threads/list"),
			sdkmodule.WithArgs("max_results?", "page_token?", "q?", "label_ids?", "include_spam_trash?")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "threads_modify"),
			a.threadsModify,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.threads/modify"),
			sdkmodule.WithArgs("id", "add_label_ids?", "remove_label_ids?")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "threads_trash"),
			a.threadsTrash,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.threads/trash"),
			sdkmodule.WithArgs("id")),
		sdkmodule.ExportFunction(
			withOrWithout(prefix, "threads_untrash"),
			a.threadsUntrash,
			sdkmodule.WithFuncDoc("https://developers.google.com/gmail/api/reference/rest/v1/users.threads/untrash"),
			sdkmodule.WithArgs("id")),
	}
}

func withOrWithout(prefix bool, s string) string {
	if prefix {
		return "gmail_" + s
	}
	return s
}

func (a api) gmailClient(ctx context.Context) (*gmail.Service, error) {
	data, err := a.connectionData(ctx)
	if err != nil {
		return nil, err
	}

	var src oauth2.TokenSource
	if data.OAuthData != "" {
		if src, err = a.oauthTokenSource(ctx, data.OAuthData); err != nil {
			return nil, err
		}
	} else {
		src, err = a.jwtTokenSource(ctx, data.JSON)
		if err != nil {
			return nil, err
		}
	}

	svc, err := gmail.NewService(ctx, option.WithTokenSource(src))
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

	var vars vars.Vars
	vs.Decode(&vars)
	return &vars, nil
}

func (a api) oauthTokenSource(ctx context.Context, data string) (oauth2.TokenSource, error) {
	tok, err := sdkintegrations.DecodeOAuthData(data)
	if err != nil {
		return nil, err
	}

	return oauthConfig().TokenSource(ctx, tok.Token), nil
}

// TODO(ENG-112): Use OAuth().Get() instead of calling this function.
func oauthConfig() *oauth2.Config {
	addr := os.Getenv("WEBHOOK_ADDRESS")
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
		RedirectURL:  fmt.Sprintf("https://%s/oauth/redirect/google", addr),
		// https://developers.google.com/gmail/api/auth/scopes
		Scopes: []string{
			// Non-sensitive.
			googleoauth2.OpenIDScope,
			googleoauth2.UserinfoEmailScope,
			googleoauth2.UserinfoProfileScope,
			// Restricted.
			gmail.GmailModifyScope,
			gmail.GmailSettingsBasicScope,
		},
	}
}

func (a api) jwtTokenSource(ctx context.Context, data string) (oauth2.TokenSource, error) {
	scopes := oauthConfig().Scopes

	cfg, err := google.JWTConfigFromJSON([]byte(data), scopes...)
	if err != nil {
		return nil, err
	}

	return cfg.TokenSource(ctx), nil
}
