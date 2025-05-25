package events

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// Update represents a Telegram update object
// https://core.telegram.org/bots/api#update
type Update struct {
	UpdateID           int64               `json:"update_id"`
	Message            *Message            `json:"message,omitempty"`
	EditedMessage      *Message            `json:"edited_message,omitempty"`
	ChannelPost        *Message            `json:"channel_post,omitempty"`
	EditedChannelPost  *Message            `json:"edited_channel_post,omitempty"`
	CallbackQuery      *CallbackQuery      `json:"callback_query,omitempty"`
	InlineQuery        *InlineQuery        `json:"inline_query,omitempty"`
	ChosenInlineResult *ChosenInlineResult `json:"chosen_inline_result,omitempty"`
}

// Message represents a Telegram message
// https://core.telegram.org/bots/api#message
type Message struct {
	MessageID       int64           `json:"message_id"`
	From            *User           `json:"from,omitempty"`
	Date            int64           `json:"date"`
	Chat            Chat            `json:"chat"`
	ForwardFrom     *User           `json:"forward_from,omitempty"`
	ForwardFromChat *Chat           `json:"forward_from_chat,omitempty"`
	ForwardDate     int64           `json:"forward_date,omitempty"`
	ReplyToMessage  *Message        `json:"reply_to_message,omitempty"`
	EditDate        int64           `json:"edit_date,omitempty"`
	Text            string          `json:"text,omitempty"`
	Entities        []MessageEntity `json:"entities,omitempty"`
	Photo           []PhotoSize     `json:"photo,omitempty"`
	Document        *Document       `json:"document,omitempty"`
	Video           *Video          `json:"video,omitempty"`
	Voice           *Voice          `json:"voice,omitempty"`
	Caption         string          `json:"caption,omitempty"`
	Contact         *Contact        `json:"contact,omitempty"`
	Location        *Location       `json:"location,omitempty"`
}

// User represents a Telegram user
// https://core.telegram.org/bots/api#user
type User struct {
	ID           int64  `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

// Chat represents a Telegram chat
// https://core.telegram.org/bots/api#chat
type Chat struct {
	ID          int64  `json:"id"`
	Type        string `json:"type"`
	Title       string `json:"title,omitempty"`
	Username    string `json:"username,omitempty"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	Description string `json:"description,omitempty"`
}

// MessageEntity represents a special entity in a text message
// https://core.telegram.org/bots/api#messageentity
type MessageEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	URL    string `json:"url,omitempty"`
	User   *User  `json:"user,omitempty"`
}

// PhotoSize represents one size of a photo or a file/sticker thumbnail
// https://core.telegram.org/bots/api#photosize
type PhotoSize struct {
	FileID   string `json:"file_id"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	FileSize int    `json:"file_size,omitempty"`
}

// Document represents a general file
// https://core.telegram.org/bots/api#document
type Document struct {
	FileID   string     `json:"file_id"`
	Thumb    *PhotoSize `json:"thumb,omitempty"`
	FileName string     `json:"file_name,omitempty"`
	MimeType string     `json:"mime_type,omitempty"`
	FileSize int        `json:"file_size,omitempty"`
}

// Video represents a video file
// https://core.telegram.org/bots/api#video
type Video struct {
	FileID   string     `json:"file_id"`
	Width    int        `json:"width"`
	Height   int        `json:"height"`
	Duration int        `json:"duration"`
	Thumb    *PhotoSize `json:"thumb,omitempty"`
	MimeType string     `json:"mime_type,omitempty"`
	FileSize int        `json:"file_size,omitempty"`
}

// Voice represents a voice note
// https://core.telegram.org/bots/api#voice
type Voice struct {
	FileID   string `json:"file_id"`
	Duration int    `json:"duration"`
	MimeType string `json:"mime_type,omitempty"`
	FileSize int    `json:"file_size,omitempty"`
}

// Contact represents a phone contact
// https://core.telegram.org/bots/api#contact
type Contact struct {
	PhoneNumber string `json:"phone_number"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name,omitempty"`
	UserID      int64  `json:"user_id,omitempty"`
}

// Location represents a point on the map
// https://core.telegram.org/bots/api#location
type Location struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// CallbackQuery represents an incoming callback query from a callback button
// https://core.telegram.org/bots/api#callbackquery
type CallbackQuery struct {
	ID              string   `json:"id"`
	From            User     `json:"from"`
	Message         *Message `json:"message,omitempty"`
	InlineMessageID string   `json:"inline_message_id,omitempty"`
	Data            string   `json:"data,omitempty"`
}

// InlineQuery represents an incoming inline query
// https://core.telegram.org/bots/api#inlinequery
type InlineQuery struct {
	ID       string    `json:"id"`
	From     User      `json:"from"`
	Query    string    `json:"query"`
	Offset   string    `json:"offset"`
	Location *Location `json:"location,omitempty"`
}

// ChosenInlineResult represents a result of an inline query chosen by a user
// https://core.telegram.org/bots/api#choseninlineresult
type ChosenInlineResult struct {
	ResultID        string    `json:"result_id"`
	From            User      `json:"from"`
	Location        *Location `json:"location,omitempty"`
	InlineMessageID string    `json:"inline_message_id,omitempty"`
	Query           string    `json:"query"`
}

// WrapUpdate wraps a Telegram update in an AutoKitteh event
func WrapUpdate(update Update, cid sdktypes.ConnectionID, iid sdktypes.IntegrationID) (sdktypes.Event, error) {
	eventType := determineEventType(update)

	// Convert the update using AutoKitteh's value wrapper
	wrapped, err := sdktypes.WrapValue(update)
	if err != nil {
		return sdktypes.InvalidEvent, fmt.Errorf("failed to wrap Telegram update: %w", err)
	}

	data, err := wrapped.ToStringValuesMap()
	if err != nil {
		return sdktypes.InvalidEvent, fmt.Errorf("failed to convert wrapped Telegram update: %w", err)
	}

	return sdktypes.EventFromProto(&sdktypes.EventPB{
		EventType: eventType,
		Data:      kittehs.TransformMapValues(data, sdktypes.ToProto),
	})
}

// determineEventType determines the event type based on the update content
func determineEventType(update Update) string {
	switch {
	case update.Message != nil:
		return "message"
	case update.EditedMessage != nil:
		return "edited_message"
	case update.ChannelPost != nil:
		return "channel_post"
	case update.EditedChannelPost != nil:
		return "edited_channel_post"
	case update.CallbackQuery != nil:
		return "callback_query"
	case update.InlineQuery != nil:
		return "inline_query"
	case update.ChosenInlineResult != nil:
		return "chosen_inline_result"
	default:
		return "unknown_update"
	}
}
