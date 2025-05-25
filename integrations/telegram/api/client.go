package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	telegramAPIURL = "https://api.telegram.org/bot"
	defaultTimeout = 30 * time.Second
)

// Client is a Telegram Bot API client
type Client struct {
    token      string
    baseURL    string  // ADD THIS LINE
    httpClient *http.Client
}

// NewClient creates a new Telegram API client
func NewClient(token string) *Client {
    return &Client{
        token:   token,
        baseURL: telegramAPIURL + token,  // ADD THIS LINE
        httpClient: &http.Client{
            Timeout: defaultTimeout,
        },
    }
}

// Add test helper method
func (c *Client) setBaseURLForTesting(url string) {
    c.baseURL = url
}

// APIResponse represents a response from the Telegram Bot API
type APIResponse struct {
	OK          bool            `json:"ok"`
	Result      json.RawMessage `json:"result,omitempty"`
	ErrorCode   int             `json:"error_code,omitempty"`
	Description string          `json:"description,omitempty"`
}

// User represents a Telegram user or bot
type User struct {
	ID                      int64  `json:"id"`
	IsBot                   bool   `json:"is_bot"`
	FirstName               string `json:"first_name"`
	LastName                string `json:"last_name,omitempty"`
	Username                string `json:"username,omitempty"`
	LanguageCode            string `json:"language_code,omitempty"`
	CanJoinGroups           bool   `json:"can_join_groups,omitempty"`
	CanReadAllGroupMessages bool   `json:"can_read_all_group_messages,omitempty"`
	SupportsInlineQueries   bool   `json:"supports_inline_queries,omitempty"`
}

// Chat represents a Telegram chat
type Chat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// Message represents a Telegram message
type Message struct {
	MessageID int64  `json:"message_id"`
	From      *User  `json:"from,omitempty"`
	Date      int64  `json:"date"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text,omitempty"`
	Caption   string `json:"caption,omitempty"`
	Entities  []MessageEntity `json:"entities,omitempty"`        // Add this
	CaptionEntities []MessageEntity `json:"caption_entities,omitempty"` // Optional: for caption formatting
}

type MessageEntity struct {
    Type     string `json:"type"`
    Offset   int    `json:"offset"`
    Length   int    `json:"length"`
    URL      string `json:"url,omitempty"`
    User     *User  `json:"user,omitempty"`
    Language string `json:"language,omitempty"`
}

// SendMessageParams represents parameters for sending a message
type SendMessageParams struct {
	ChatID                int64       `json:"chat_id"`
	Text                  string      `json:"text"`
	ParseMode             string      `json:"parse_mode,omitempty"`
	Entities              []MessageEntity `json:"entities,omitempty"`
	DisableWebPagePreview bool        `json:"disable_web_page_preview,omitempty"`
	DisableNotification   bool        `json:"disable_notification,omitempty"`
	ReplyToMessageID      int64       `json:"reply_to_message_id,omitempty"`
	ReplyMarkup           interface{} `json:"reply_markup,omitempty"`
}

// WebhookInfo represents information about a webhook
type WebhookInfo struct {
	URL                    string   `json:"url"`
	HasCustomCertificate   bool     `json:"has_custom_certificate"`
	PendingUpdateCount     int      `json:"pending_update_count"`
	LastErrorDate          int64    `json:"last_error_date,omitempty"`
	LastErrorMessage       string   `json:"last_error_message,omitempty"`
	MaxConnections         int      `json:"max_connections,omitempty"`
	AllowedUpdates         []string `json:"allowed_updates,omitempty"`
}

// SetWebhookParams represents parameters for setting a webhook
type SetWebhookParams struct {
	URL            string   `json:"url"`
	Certificate    string   `json:"certificate,omitempty"`
	MaxConnections int      `json:"max_connections,omitempty"`
	AllowedUpdates []string `json:"allowed_updates,omitempty"`
	DropPendingUpdates bool `json:"drop_pending_updates,omitempty"`
	SecretToken    string   `json:"secret_token,omitempty"`
}

// InlineKeyboardMarkup represents an inline keyboard
type InlineKeyboardMarkup struct {
    InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

// InlineKeyboardButton represents one button of an inline keyboard
type InlineKeyboardButton struct {
    Text                         string `json:"text"`
    URL                          string `json:"url,omitempty"`
    CallbackData                 string `json:"callback_data,omitempty"`
    WebApp                       *WebApp `json:"web_app,omitempty"`
    SwitchInlineQuery            string `json:"switch_inline_query,omitempty"`
    SwitchInlineQueryCurrentChat string `json:"switch_inline_query_current_chat,omitempty"`
}

// ReplyKeyboardMarkup represents a custom keyboard with reply options
type ReplyKeyboardMarkup struct {
    Keyboard              [][]KeyboardButton `json:"keyboard"`
    ResizeKeyboard        bool               `json:"resize_keyboard,omitempty"`
    OneTimeKeyboard       bool               `json:"one_time_keyboard,omitempty"`
    InputFieldPlaceholder string             `json:"input_field_placeholder,omitempty"`
    Selective             bool               `json:"selective,omitempty"`
}

// KeyboardButton represents one button of the reply keyboard
type KeyboardButton struct {
    Text            string                  `json:"text"`
    RequestContact  bool                    `json:"request_contact,omitempty"`
    RequestLocation bool                    `json:"request_location,omitempty"`
    RequestPoll     *KeyboardButtonPollType `json:"request_poll,omitempty"`
    WebApp          *WebApp                 `json:"web_app,omitempty"`
}

// KeyboardButtonPollType represents type of poll which is allowed to be created and sent when the corresponding button is pressed
type KeyboardButtonPollType struct {
    Type string `json:"type,omitempty"`
}

// WebApp represents a parameter of the inline keyboard button used to launch a Web App
type WebApp struct {
    URL string `json:"url"`
}

// ReplyKeyboardRemove represents a parameter to hide the current custom keyboard and display the default letter-keyboard
type ReplyKeyboardRemove struct {
    RemoveKeyboard bool `json:"remove_keyboard"`
    Selective      bool `json:"selective,omitempty"`
}

// ForceReply represents a parameter to display a reply interface to the user
type ForceReply struct {
    ForceReply            bool   `json:"force_reply"`
    InputFieldPlaceholder string `json:"input_field_placeholder,omitempty"`
    Selective             bool   `json:"selective,omitempty"`
}

// CallbackQuery represents an incoming callback query from a callback button in an inline keyboard
type CallbackQuery struct {
    ID              string   `json:"id"`
    From            User     `json:"from"`
    Message         *Message `json:"message,omitempty"`
    InlineMessageID string   `json:"inline_message_id,omitempty"`
    ChatInstance    string   `json:"chat_instance"`
    Data            string   `json:"data,omitempty"`
    GameShortName   string   `json:"game_short_name,omitempty"`
}

// Update represents an incoming update
type Update struct {
    UpdateID           int             `json:"update_id"`
    Message            *Message        `json:"message,omitempty"`
    EditedMessage      *Message        `json:"edited_message,omitempty"`
    ChannelPost        *Message        `json:"channel_post,omitempty"`
    EditedChannelPost  *Message        `json:"edited_channel_post,omitempty"`
    CallbackQuery      *CallbackQuery  `json:"callback_query,omitempty"`
    // Add other update types as needed
}

// Helper functions for creating keyboards
func NewInlineKeyboard(buttons [][]InlineKeyboardButton) *InlineKeyboardMarkup {
    return &InlineKeyboardMarkup{InlineKeyboard: buttons}
}

func NewInlineButton(text, callbackData string) InlineKeyboardButton {
    return InlineKeyboardButton{Text: text, CallbackData: callbackData}
}

func NewInlineURLButton(text, url string) InlineKeyboardButton {
    return InlineKeyboardButton{Text: text, URL: url}
}

func NewReplyKeyboard(buttons [][]KeyboardButton) *ReplyKeyboardMarkup {
    return &ReplyKeyboardMarkup{Keyboard: buttons}
}

func NewKeyboardButton(text string) KeyboardButton {
    return KeyboardButton{Text: text}
}

// AnswerCallbackQuery answers a callback query
func (c *Client) AnswerCallbackQuery(ctx context.Context, callbackQueryID, text string, showAlert bool) error {
    params := map[string]interface{}{
        "callback_query_id": callbackQueryID,
        "text":              text,
        "show_alert":        showAlert,
    }
    _, err := c.makeRequest(ctx, "answerCallbackQuery", params)
    return err
}

// makeRequest makes a request to the Telegram Bot API
func (c *Client) makeRequest(ctx context.Context, method string, params interface{}) (*APIResponse, error) {
    url := fmt.Sprintf("%s/%s", c.baseURL, method) 
	
	var reqBody io.Reader
	if params != nil {
		jsonData, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	if params != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	if !apiResp.OK {
		return &apiResp, fmt.Errorf("telegram API error %d: %s", apiResp.ErrorCode, apiResp.Description)
	}
	
	return &apiResp, nil
}

// GetMe returns information about the bot
func (c *Client) GetMe(ctx context.Context) (*User, error) {
	resp, err := c.makeRequest(ctx, "getMe", nil)
	if err != nil {
		return nil, err
	}
	
	var user User
	if err := json.Unmarshal(resp.Result, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}
	
	return &user, nil
}

// SendMessage sends a text message
func (c *Client) SendMessage(ctx context.Context, params SendMessageParams) (*Message, error) {
	resp, err := c.makeRequest(ctx, "sendMessage", params)
	if err != nil {
		return nil, err
	}
	
	var message Message
	if err := json.Unmarshal(resp.Result, &message); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}
	
	return &message, nil
}

// SetWebhook sets a webhook for the bot
func (c *Client) SetWebhook(ctx context.Context, params SetWebhookParams) error {
	_, err := c.makeRequest(ctx, "setWebhook", params)
	return err
}

// DeleteWebhook removes the webhook integration
func (c *Client) DeleteWebhook(ctx context.Context, dropPendingUpdates bool) error {
	params := map[string]bool{
		"drop_pending_updates": dropPendingUpdates,
	}
	_, err := c.makeRequest(ctx, "deleteWebhook", params)
	return err
}

// GetWebhookInfo gets current webhook status
func (c *Client) GetWebhookInfo(ctx context.Context) (*WebhookInfo, error) {
	resp, err := c.makeRequest(ctx, "getWebhookInfo", nil)
	if err != nil {
		return nil, err
	}
	
	var webhookInfo WebhookInfo
	if err := json.Unmarshal(resp.Result, &webhookInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook info: %w", err)
	}
	
	return &webhookInfo, nil
}
