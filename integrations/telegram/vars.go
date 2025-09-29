package telegram

import "go.autokitteh.dev/autokitteh/sdk/sdktypes"

var (
	BotToken    = sdktypes.NewSymbol("BotToken")
	SecretToken = sdktypes.NewSymbol("SecretToken")
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

// TelegramUpdate represents a Telegram webhook update
type TelegramUpdate struct {
	UpdateID int               `json:"update_id"`
	Message  *TelegramMessage  `json:"message,omitempty"`
	Callback *TelegramCallback `json:"callback_query,omitempty"`
	Edited   *TelegramMessage  `json:"edited_message,omitempty"`
}

// TelegramMessage represents a Telegram message
type TelegramMessage struct {
	MessageID int             `json:"message_id"`
	From      *TelegramUser   `json:"from,omitempty"`
	Chat      *TelegramChat   `json:"chat"`
	Date      int64           `json:"date"`
	Text      string          `json:"text,omitempty"`
	Photo     []TelegramPhoto `json:"photo,omitempty"`
	Document  *TelegramDoc    `json:"document,omitempty"`
}

// TelegramChat represents a Telegram chat
type TelegramChat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// TelegramPhoto represents a photo in different sizes
type TelegramPhoto struct {
	FileID   string `json:"file_id"`
	FileSize int    `json:"file_size,omitempty"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

// TelegramDoc represents a document
type TelegramDoc struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
	FileSize int    `json:"file_size,omitempty"`
}

// TelegramCallback represents a callback query
type TelegramCallback struct {
	ID      string           `json:"id"`
	From    *TelegramUser    `json:"from"`
	Message *TelegramMessage `json:"message,omitempty"`
	Data    string           `json:"data,omitempty"`
}
