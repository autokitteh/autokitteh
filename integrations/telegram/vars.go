package telegram

import "go.autokitteh.dev/autokitteh/sdk/sdktypes"

const (
	// HTTP header that contains the secret token for webhook verification.
	headerSecretToken = "X-Telegram-Bot-Api-Secret-Token"
)

var (
	BotTokenVar    = sdktypes.NewSymbol("BotToken")
	SecretTokenVar = sdktypes.NewSymbol("SecretToken")
	BotIDVar       = sdktypes.NewSymbol("BotID")
	WebhookURLVar  = sdktypes.NewSymbol("WebhookURL")
)

// List of possible Telegram event types
var telegramEventTypes = []string{
	"message",
	"callback_query",
	"edited_message",
	"inline_query",
	"chosen_inline_result",
	"channel_post",
	"edited_channel_post",
	"shipping_query",
	"pre_checkout_query",
	"poll",
	"poll_answer",
	"my_chat_member",
	"chat_member",
	"chat_join_request",
	"message_reaction",
	"message_reaction_count",
}

// TelegramVars represents the variables stored for a Telegram connection
type TelegramVars struct {
	BotToken    string `vars:"secret"`
	SecretToken string `vars:"secret"`
	BotID       string
	WebhookURL  string
}

// TelegramUser represents a Telegram user (bot) response from getMe API
type TelegramUser struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

// TelegramResponse represents the standard Telegram API response format
type TelegramResponse struct {
	OK     bool         `json:"ok"`
	Result TelegramUser `json:"result"`
}
