package googleoauth

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Config struct {
	ClientID     string   `envconfig:"CLIENT_ID" json:"client_id"`
	ClientSecret string   `envconfig:"CLIENT_SECRET" json:"client_secret"`
	RedirectURL  string   `envconfig:"REDIRECT_URL" json:"redirect_url"`
	Scopes       []string `envconfig:"SCOPES" json:"scopes"`
}

func MakeConfig(cfg Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       cfg.Scopes,
		Endpoint:     google.Endpoint,
	}
}
