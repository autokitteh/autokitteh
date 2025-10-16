package discord

import (
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/discord/internal/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// ServeHTTP saves a new autokitteh connection with user-submitted data.
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	token := r.Form.Get("botToken")

	bot, err := infoWithToken(token)
	if err != nil {
		l.Warn("Failed to initialize bot client with provided token", zap.Error(err))
		c.AbortBadRequest("failed to use the provided token")
		return
	}

	h.OpenWebSocketConnection(token)

	c.Finalize(sdktypes.NewVars().
		Set(vars.BotID, bot.ID, false).
		Set(vars.BotToken, token, true).
		Set(vars.AuthType, integrations.Init, false))
}

func infoWithToken(botToken string) (*discordgo.User, error) {
	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		return nil, err
	}

	bot, err := dg.User("@me")
	if err != nil {
		return nil, fmt.Errorf("user(me): %w", err)
	}
	return bot, nil
}
