package discord

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type DiscordAPI struct {
	Token string
}

func (api DiscordAPI) Test(ctx context.Context) error {
	dg, err := discordgo.New("Bot " + api.Token)
	if err != nil {
		return fmt.Errorf("failed to create Discord session: %w", err)
	}

	// Check if the bot can successfully retrieve its own user data.
	_, err = dg.User("@me")
	if err != nil {
		return fmt.Errorf("failed to authenticate with Discord: %w", err)
	}
	return nil
}
