package slack

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/auth"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/bookmarks"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/bots"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/chat"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/conversations"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/reactions"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/users"
	"go.autokitteh.dev/autokitteh/integrations/slack/internal/vars"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var integrationID = sdktypes.NewIntegrationIDFromName("slack")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: integrationID.String(),
	UniqueName:    "slack",
	DisplayName:   "Slack",
	Description:   "Slack is a cloud-based team communication platform.",
	LogoUrl:       "/static/images/slack.svg",
	UserLinks: map[string]string{
		"1 Web API reference":    "https://api.slack.com/methods",
		"2 Events API reference": "https://api.slack.com/events?filter=Events",
		"3 Python client API":    "https://slack.dev/python-slack-sdk/api-docs/slack_sdk/",
	},
	ConnectionUrl: "/slack/connect",
	ConnectionCapabilities: &sdktypes.ConnectionCapabilitiesPB{
		RequiresConnectionInit: true,
	},
}))

func New(vs sdkservices.Vars) sdkservices.Integration {
	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New(exportFuncs(vs)...),
		connStatus(vs),
		connTest(vs),
		sdkintegrations.WithConnectionConfigFromVars(vs),
	)
}

// connStatus is an optional connection status check provided by
// the integration to AutoKitteh. The possible results are "Init
// required" (the connection is not usable yet) and "Using X".
func connStatus(cvars sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		vs, err := cvars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			zap.L().Error("failed to read connection vars", zap.String("connection_id", cid.String()), zap.Error(err))
			return sdktypes.InvalidStatus, err
		}

		at := vs.Get(vars.AuthType)
		if !at.IsValid() || at.Value() == "" {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		switch at.Value() {
		case integrations.OAuth, integrations.OAuthDefault:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using OAuth v2"), nil
		case integrations.SocketMode:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using Socket Mode"), nil
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}
	})
}

// connTest is an optional connection test provided by the integration
// to AutoKitteh. It is used to verify that the connection is working
// as expected. The possible results are "OK" and "error".
func connTest(cvars sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Init required"), nil
		}

		vs, err := cvars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			zap.L().Error("failed to read connection vars", zap.String("connection_id", cid.String()), zap.Error(err))
			return sdktypes.InvalidStatus, err
		}

		at := vs.Get(vars.AuthType)
		if !at.IsValid() || at.Value() == "" {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		token := ""

		switch at.Value() {
		case integrations.OAuth:
			token = vs.GetValueByString("oauth_AccessToken")
		case integrations.SocketMode:
			token = vs.Get(vars.BotTokenName).Value()
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}

		_, err = auth.TestWithToken(ctx, token)
		if err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, ""), nil
	})
}

func exportFuncs(vs sdkservices.Vars) []sdkmodule.Optfn {
	authAPI := auth.API{Vars: vs}
	bookmarksAPI := bookmarks.API{Vars: vs}
	botsAPI := bots.API{Vars: vs}
	chatAPI := chat.API{Vars: vs}
	conversationsAPI := conversations.API{Vars: vs}
	reactionsAPI := reactions.API{Vars: vs}
	usersAPI := users.API{Vars: vs}

	return []sdkmodule.Optfn{
		// Auth.
		sdkmodule.ExportFunction(
			"auth_test",
			authAPI.Test,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/auth.test"),
		),

		// Bookmarks.
		sdkmodule.ExportFunction(
			"bookmarks_add",
			bookmarksAPI.Add,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/bookmarks.add"),
			sdkmodule.WithArgs("channel_id", "title", "type?", "link?", "emoji?", "entity_id?", "parent_id?"),
		),
		sdkmodule.ExportFunction(
			"bookmarks_edit",
			bookmarksAPI.Edit,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/bookmarks.edit"),
			sdkmodule.WithArgs("bookmark_id", "channel_id", "emoji?", "link?", "title?"),
		),
		sdkmodule.ExportFunction(
			"bookmarks_list",
			bookmarksAPI.List,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/bookmarks.list"),
			sdkmodule.WithArgs("channel_id"),
		),
		sdkmodule.ExportFunction(
			"bookmarks_remove",
			bookmarksAPI.Remove,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/bookmarks.remove"),
			sdkmodule.WithArgs("bookmark_id", "channel_id", "quip_section_id?"),
		),

		// Bots.
		sdkmodule.ExportFunction(
			"bots_info",
			botsAPI.Info,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/bots.info"),
			sdkmodule.WithArgs("bot", "team_id?"),
		),

		// Chat.
		sdkmodule.ExportFunction(
			"chat_delete",
			chatAPI.Delete,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/chat.delete"),
			sdkmodule.WithArgs("channel", "ts"),
		),
		sdkmodule.ExportFunction(
			"chat_get_permalink",
			chatAPI.GetPermalink,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/chat.getPermalink"),
			sdkmodule.WithArgs("channel", "message_ts"),
		),
		sdkmodule.ExportFunction(
			"chat_post_ephemeral",
			chatAPI.PostEphemeral,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/chat.postEphemeral"),
			sdkmodule.WithArgs("channel", "user", "text", "blocks?", "thread_ts?"),
		),
		sdkmodule.ExportFunction(
			"chat_post_message",
			chatAPI.PostMessage,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/chat.postMessage"),
			sdkmodule.WithArgs("channel", "text?", "blocks?", "thread_ts?", "reply_broadcast?", "username?", "icon_url?"),
		),
		sdkmodule.ExportFunction(
			"chat_update",
			chatAPI.Update,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/chat.update"),
			sdkmodule.WithArgs("channel", "ts", "text?", "blocks?", "reply_broadcast?"),
		),

		// Convenience wrappers for "chat.postMessage".
		sdkmodule.ExportFunction(
			"send_text_message",
			chatAPI.SendTextMessage,
			sdkmodule.WithFuncDoc("convenience wrapper for chat.postMessage"),
			sdkmodule.WithArgs("target", "text", "thread_ts?", "reply_broadcast?"),
		),
		sdkmodule.ExportFunction(
			"send_approval_message",
			chatAPI.SendApprovalMessage,
			sdkmodule.WithFuncDoc("convenience wrapper for chat.postMessage"),
			sdkmodule.WithArgs("target", "header", "message", "green_button?", "red_button?", "thread_ts?", "reply_broadcast?"),
		),

		// Conversations.
		sdkmodule.ExportFunction(
			"conversations_archive",
			conversationsAPI.Archive,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/conversations.archive"),
			sdkmodule.WithArgs("channel"),
		),
		sdkmodule.ExportFunction(
			"conversations_close",
			conversationsAPI.Close,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/conversations.close"),
			sdkmodule.WithArgs("close"),
		),
		sdkmodule.ExportFunction(
			"conversations_create",
			conversationsAPI.Create,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/conversations.create"),
			sdkmodule.WithArgs("name", "is_private?", "team_id?"),
		),
		sdkmodule.ExportFunction(
			"conversations_history",
			conversationsAPI.History,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/conversations.history"),
			sdkmodule.WithArgs("channel", "cursor?", "limit?", "include_all_metadata?", "inclusive?", "oldest?", "latest?"),
		),
		sdkmodule.ExportFunction(
			"conversations_info",
			conversationsAPI.Info,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/conversations.info"),
			sdkmodule.WithArgs("channel", "include_locale?", "include_num_members?"),
		),
		sdkmodule.ExportFunction(
			"conversations_invite",
			conversationsAPI.Invite,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/conversations.invite"),
			sdkmodule.WithArgs("channel", "users", "force?"),
		),
		sdkmodule.ExportFunction(
			"conversations_list",
			conversationsAPI.List,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/conversations.list"),
			sdkmodule.WithArgs("cursor?", "limit?", "exclude_archived?", "team_id?", "types?"),
		),
		sdkmodule.ExportFunction(
			"conversations_members",
			conversationsAPI.Members,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/conversations.members"),
			sdkmodule.WithArgs("channel", "cursor?", "limit?"),
		),
		sdkmodule.ExportFunction(
			"conversations_open",
			conversationsAPI.Open,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/conversations.open"),
			sdkmodule.WithArgs("channel?", "users?", "prevent_creation?"),
		),
		sdkmodule.ExportFunction(
			"conversations_rename",
			conversationsAPI.Rename,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/conversations.rename"),
			sdkmodule.WithArgs("channel", "name"),
		),
		sdkmodule.ExportFunction(
			"conversations_replies",
			conversationsAPI.Replies,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/conversations.replies"),
			sdkmodule.WithArgs("channel", "ts", "cursor?", "limit?", "include_all_metadata?", "inclusive?", "oldest?", "latest?"),
		),
		sdkmodule.ExportFunction(
			"conversations_set_purpose",
			conversationsAPI.SetPurpose,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/conversations.setPurpose"),
			sdkmodule.WithArgs("channel", "purpose"),
		),
		sdkmodule.ExportFunction(
			"conversations_set_topic",
			conversationsAPI.SetTopic,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/conversations.setTopic"),
			sdkmodule.WithArgs("channel", "topic"),
		),
		sdkmodule.ExportFunction(
			"conversations_unarchive",
			conversationsAPI.Unarchive,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/conversations.unarchive"),
			sdkmodule.WithArgs("channel"),
		),

		// Reactions.
		sdkmodule.ExportFunction(
			"reactions_add",
			reactionsAPI.Add,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/reactions.add"),
			sdkmodule.WithArgs("channel", "name", "timestamp"),
		),

		// Users.
		// TODO(ENG-1057): sdkmodule.ExportFunction(
		// "users_conversations",
		// 	sdkmodule.WithFuncDoc("https://api.slack.com/methods/users.conversations"),
		// 	sdkmodule.WithArgs(...TODO...),
		// 	usersAPI.GetPresence),
		sdkmodule.ExportFunction(
			"users_get_presence",
			usersAPI.GetPresence,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/users.getPresence"),
			sdkmodule.WithArgs("user?"),
		),
		sdkmodule.ExportFunction(
			"users_info",
			usersAPI.Info,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/users.info"),
			sdkmodule.WithArgs("user", "include_locale?"),
		),
		sdkmodule.ExportFunction(
			"users_list",
			usersAPI.List,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/users.list"),
			sdkmodule.WithArgs("cursor?", "limit?", "include_locale?", "team_id?"),
		),
		sdkmodule.ExportFunction(
			"users_lookup_by_email",
			usersAPI.LookupByEmail,
			sdkmodule.WithFuncDoc("https://api.slack.com/methods/users.lookupByEmail"),
			sdkmodule.WithArgs("email"),
		),
	}
}
