package authloginhttpsvc

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/dghubble/gologin/v2/github"
	"golang.org/x/oauth2"
	githubOAuth2 "golang.org/x/oauth2/github"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const githubLoginPath = "/auth/github/login"

func registerGithubOAuthRoutes(mux *http.ServeMux, cfg oauth2Config, onSuccess func(sdktypes.User) http.Handler) error {
	if cfg.ClientID == "" {
		return errors.New("github login is enabled, but missing GITHUB_CLIENT_ID")
	}

	if cfg.ClientSecret == "" {
		return errors.New("github is enabled, but missing GITHUB_CLIENT_SECRET")
	}

	if cfg.RedirectURL == "" {
		return errors.New("github is enabled, but missing GITHUB_REDIRECT_URL")
	}

	oauth2Config := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     githubOAuth2.Endpoint,
	}

	mux.Handle(githubLoginPath, github.StateHandler(cfg.cookieConfig(), github.LoginHandler(&oauth2Config, nil)))

	githubOnSuccess := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gu, err := github.UserFromContext(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if gu.Login == nil || gu.ID == nil || gu.Name == nil || gu.Email == nil {
			http.Error(w, "github user missing data", http.StatusInternalServerError)
			return
		}

		onSuccess(sdktypes.NewUser(
			"github",
			map[string]string{
				"login": *gu.Login,
				"id":    strconv.FormatInt(*gu.ID, 10),
				"name":  *gu.Name,
				"email": *gu.Email,
			},
		))
	})

	mux.Handle("/auth/github/callback", github.StateHandler(cfg.cookieConfig(), github.CallbackHandler(&oauth2Config, githubOnSuccess, nil)))

	return nil
}
