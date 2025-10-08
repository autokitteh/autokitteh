package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handleSave saves a new autokitteh connection with user-submitted data.
// Embeds the bot ID in the webhook URL path since Telegram doesn't provide
// bot identification in webhook payloads, allowing proper request routing.
func (h handler) handleSave(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	// Check the "Content-Type" header.
	if common.PostWithoutFormContentType(r) {
		ct := r.Header.Get(common.HeaderContentType)
		l.Warn("save connection: unexpected POST content type", zap.String("content_type", ct))
		c.AbortBadRequest("unexpected content type")
		return
	}

	// Read and parse POST request body.
	if err := r.ParseForm(); err != nil {
		l.Warn("Failed to parse incoming HTTP request", zap.Error(err))
		c.AbortBadRequest("form parsing error")
		return
	}

	// Sanity check: the connection ID is valid.
	cid, err := sdktypes.StrictParseConnectionID(c.ConnectionID)
	if err != nil {
		l.Warn("save connection: invalid connection ID", zap.Error(err))
		c.AbortBadRequest("invalid connection ID")
		return
	}

	token := r.FormValue("bot_token")
	if token == "" {
		l.Warn("save connection: missing bot token")
		c.AbortBadRequest("missing bot token")
		return
	}

	vsid := sdktypes.NewVarScopeID(cid)

	// Validate the bot token usability.
	bot, err := getBotInfoWithToken(token, r.Context())
	if err != nil {
		l.Warn("bot token test failed", zap.Error(err))
		c.AbortBadRequest("Authentication failed. Check your Telegram bot token and retry")
		return
	}
	// Use bot ID as webhook identifier - much better than random!
	botID := strconv.FormatInt(bot.ID, 10)
	webhookSecret := strings.ReplaceAll(sdktypes.NewUUID().String(), "-", "")
	webhookURL, err := constructWebhookURL(botID)
	if err != nil {
		l.Error("failed to construct webhook URL", zap.Error(err))
		c.AbortServerError("failed to construct webhook URL")
		return
	}

	// Register webhook with Telegram.
	if err := setTelegramWebhook(r.Context(), token, webhookURL, webhookSecret); err != nil {
		l.Error("failed to register webhook with Telegram", zap.Error(err))
		c.AbortServerError("failed to connect to Telegram, please try again later")
		return
	}

	// Prepare all variables to save at once using the struct
	telegramVars := TelegramVars{
		BotToken:    token,
		SecretToken: webhookSecret,
		BotID:       botID,
		WebhookURL:  webhookURL,
	}

	common.SaveAuthType(r, h.vars, vsid)
	vars := sdktypes.EncodeVars(telegramVars)
	if err := h.vars.Set(r.Context(), vars.WithScopeID(vsid)...); err != nil {
		l.Error("save connection: failed to save connection variables", zap.Error(err))
		c.AbortServerError("failed to save connection variables")
		return
	}
}

// getBotInfoWithToken validates the bot token by calling Telegram's getMe API
func getBotInfoWithToken(botToken string, ctx context.Context) (*TelegramUser, error) {
	if botToken == "" {
		return nil, errors.New("bot token is required")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", botToken)

	respBody, err := common.HTTPGet(ctx, url, "")
	if err != nil {
		return nil, fmt.Errorf("failed to call Telegram API: %w", err)
	}

	var telegramResp TelegramResponse
	if err := json.Unmarshal(respBody, &telegramResp); err != nil {
		return nil, fmt.Errorf("failed to decode Telegram API response: %w", err)
	}

	if !telegramResp.OK {
		return nil, errors.New("telegram API returned error")
	}

	if !telegramResp.Result.IsBot {
		return nil, errors.New("provided token is not for a bot")
	}

	return &telegramResp.Result, nil
}

// constructWebhookURL builds the full webhook URL for this connection.
func constructWebhookURL(botID string) (string, error) {
	baseURL := os.Getenv("WEBHOOK_ADDRESS")
	if baseURL == "" {
		err := errors.New("WEBHOOK_ADDRESS environment variable is not set")
		return "", err
	}

	return path.Join(baseURL, "telegram/webhook", botID), nil
}

// setTelegramWebhook registers the webhook with Telegram using their API
func setTelegramWebhook(ctx context.Context, botToken, webhookURL, secretToken string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook", botToken)

	payload := map[string]string{
		"url":          webhookURL,
		"secret_token": secretToken,
	}

	_, err := common.HTTPPostJSON(ctx, url, "", payload)
	if err != nil {
		return fmt.Errorf("failed to set Telegram webhook: %w", err)
	}

	return nil
}
