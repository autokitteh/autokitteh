package vars

import (
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
    // Bot token provided by BotFather
    BotTokenVar = sdktypes.NewSymbol("bot_token")
    
    // Optional webhook secret for additional security
    WebhookSecretVar = sdktypes.NewSymbol("webhook_secret")
    
    // Bot information stored after successful authentication
    BotIDVar                     = sdktypes.NewSymbol("bot_id")
    BotUsernameVar               = sdktypes.NewSymbol("bot_username")
    BotFirstNameVar              = sdktypes.NewSymbol("bot_first_name")
    BotLastNameVar               = sdktypes.NewSymbol("bot_last_name")
    BotCanJoinGroupsVar          = sdktypes.NewSymbol("bot_can_join_groups")
    BotCanReadAllGroupMessagesVar = sdktypes.NewSymbol("bot_can_read_all_group_messages")
    BotSupportsInlineQueriesVar  = sdktypes.NewSymbol("bot_supports_inline_queries")
)

// BotConfig contains the user-provided bot configuration.
type BotConfig struct {
    BotToken      string `var:"bot_token,secret"`
    WebhookSecret string `var:"webhook_secret,secret"`
}

// BotInfo contains the bot information retrieved from Telegram.
type BotInfo struct {
    ID                      int64  `var:"bot_id"`
    Username                string `var:"bot_username"`
    FirstName               string `var:"bot_first_name"`
    LastName                string `var:"bot_last_name"`
    CanJoinGroups           bool   `var:"bot_can_join_groups"`
    CanReadAllGroupMessages bool   `var:"bot_can_read_all_group_messages"`
    SupportsInlineQueries   bool   `var:"bot_supports_inline_queries"`
}