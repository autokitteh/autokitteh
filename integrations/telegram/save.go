package telegram

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"

	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// ServeHTTP saves a new autokitteh connection with user-submitted data.
func (h handler) handleSave(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	// Check the "Content-Type" header.
	if common.PostWithoutFormContentType(r) {
		ct := r.Header.Get(common.HeaderContentType)
		l.Warn("save connection: unexpected POST content type", zap.String("content_type", ct))
		c.AbortBadRequest("unexpected content type")
		return
	}

	/// Read and parse POST request body.
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

	secretToken := r.FormValue("secret_token")

	vsid := sdktypes.NewVarScopeID(cid)
	authType := common.SaveAuthType(r, h.vars, vsid)
	l = l.With(zap.String("auth_type", authType))

	// Validate the bot token usability.
	bot, err := getBotInfoWithToken(token, r)
	if err != nil {
		l.Warn("bot token test failed", zap.Error(err))
		c.AbortBadRequest("failed to use the provided token")
		return
	}

	// Save bot token.
	v := sdktypes.NewVar(BotToken).SetValue(token).SetSecret(true)
	if err := h.vars.Set(r.Context(), v.WithScopeID(vsid)); err != nil {
		l.Error("save connection: failed to save bot token", zap.Error(err))
		c.AbortServerError("failed to save bot token")
		return
	}

	// Save secret token if provided.
	if secretToken != "" {
		sv := sdktypes.NewVar(SecretToken).SetValue(secretToken)
		if err := h.vars.Set(r.Context(), sv.WithScopeID(vsid)); err != nil {
			l.Error("save connection: failed to save secret token", zap.Error(err))
			c.AbortServerError("failed to save secret token")
			return
		}
	}

	l.Info("Telegram bot connection saved successfully",
		zap.Int64("bot_id", bot.ID),
		zap.String("bot_username", bot.Username))
}

// getBotInfoWithToken validates the bot token by calling Telegram's getMe API
func getBotInfoWithToken(botToken string, r *http.Request) (*TelegramUser, error) {
	if botToken == "" {
		return nil, errors.New("bot token is required")
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", botToken)

	respBody, err := common.HTTPGet(r.Context(), url, "")
	if err != nil {
		return nil, fmt.Errorf("failed to call Telegram API: %w", err)
	}

	var telegramResp TelegramResponse
	if err := json.Unmarshal(respBody, &telegramResp); err != nil {
		return nil, fmt.Errorf("failed to decode Telegram API response: %w", err)
	}

	if !telegramResp.OK {
		return nil, errors.New("Telegram API returned error")
	}

	if !telegramResp.Result.IsBot {
		return nil, errors.New("provided token is not for a bot")
	}

	return &telegramResp.Result, nil
}
