package discord

import (
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/discord/internal/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	headerContentType = "Content-Type"
	contentTypeForm   = "application/x-www-form-urlencoded"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to save data from web form submissions as connections.
type handler struct {
	logger *zap.Logger
}

func NewHTTPHandler(l *zap.Logger) http.Handler {
	return handler{logger: l}
}

// ServeHTTP saves a new autokitteh connection with user-submitted data.
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	// Check "Content-Type" header.
	contentType := r.Header.Get(headerContentType)
	if !strings.HasPrefix(contentType, contentTypeForm) {
		c.AbortBadRequest("unexpected content type")
		return
	}

	// Read and parse POST request body.
	if err := r.ParseForm(); err != nil {
		l.Warn("Failed to parse incoming HTTP request", zap.Error(err))
		c.AbortBadRequest("form parsing error")
		return
	}

	bt := r.Form.Get("botToken")

	// get the bot's ID
	bot, err := infoWithToken(bt)
	if err != nil {
		l.Warn("Failed to create new bot", zap.Error(err))
		c.AbortBadRequest("failed to create new bot")
		return
	}

	c.Finalize(sdktypes.NewVars().
		Set(vars.BotID, bot.ID, false).
		Set(vars.BotTokenName, bt, true).
		Set(vars.AuthType, integrations.Init, false))
}

func infoWithToken(botToken string) (*discordgo.User, error) {
	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		return nil, err
	}

	bot, err := dg.User("@me")
	if err != nil {
		return nil, err
	}
	return bot, nil
}
