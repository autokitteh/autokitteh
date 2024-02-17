package gmail

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type api struct {
	Secrets sdkservices.Secrets
	Scope   string
}

var integrationID = sdktypes.IntegrationIDFromName("gmail")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: integrationID.String(),
	UniqueName:    "gmail",
	DisplayName:   "Gmail",
	Description:   "Gmail is an email service provided by Google.",
	LogoUrl:       "/static/images/gmail.svg",
	UserLinks: map[string]string{
		"1 API overview":       "https://developers.google.com/gmail/api/guides",
		"2 REST API reference": "https://developers.google.com/gmail/api/reference/rest",
		"3 Go client API":      "https://pkg.go.dev/google.golang.org/api/gmail/v1",
	},
	ConnectionUrl: "/gmail/connect",
}))

func New(sec sdkservices.Secrets) sdkservices.Integration {
	scope := sdktypes.GetIntegrationUniqueName(desc).String()

	opts := []sdkmodule.Optfn{sdkmodule.WithConfigAsData()}
	opts = append(opts, ExportedFunctions(sec, scope, false)...)

	return sdkintegrations.NewIntegration(desc, sdkmodule.New(opts...))
}

func ExportedFunctions(sec sdkservices.Secrets, scope string, prefix bool) []sdkmodule.Optfn {
	a := api{Secrets: sec, Scope: scope}
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

// getConnection calls the Get method in SecretsService.
func (a api) getConnection(ctx context.Context) (*oauth2.Token, error) {
	// Extract the connection token from the given context.
	connToken := sdkmodule.FunctionDataFromContext(ctx)

	oauthToken, err := a.Secrets.Get(ctx, a.Scope, string(connToken))
	if err != nil {
		return nil, err
	}

	exp, err := time.Parse(time.RFC3339, oauthToken["expiry"])
	if err != nil {
		exp = time.Unix(0, 0)
	}
	return &oauth2.Token{
		AccessToken:  oauthToken["accessToken"],
		TokenType:    oauthToken["tokenType"],
		RefreshToken: oauthToken["refreshToken"],
		Expiry:       exp,
	}, nil
}

// TODO(ENG-112): Use OAuth().Get() instead of defining oauth2.Config.
func tokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource {
	addr := os.Getenv("WEBHOOK_ADDRESS")
	cfg := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
		RedirectURL:  fmt.Sprintf("https://%s/oauth/redirect/google", addr),
		Scopes: []string{
			googleoauth2.OpenIDScope,
			googleoauth2.UserinfoEmailScope,
			googleoauth2.UserinfoProfileScope,
			gmail.GmailModifyScope,
		},
	}
	return cfg.TokenSource(ctx, t)
}

func (a api) gmailClient(ctx context.Context) (*gmail.Service, error) {
	oauthToken, err := a.getConnection(ctx)
	if err != nil {
		return nil, err
	}

	src := tokenSource(ctx, oauthToken)
	svc, err := gmail.NewService(ctx, option.WithTokenSource(src))
	if err != nil {
		return nil, err
	}
	return svc, nil
}
