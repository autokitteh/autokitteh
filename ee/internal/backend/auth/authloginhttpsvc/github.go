package authloginhttpsvc

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/dghubble/gologin/v2/github"
	"go.uber.org/zap"

	"golang.org/x/oauth2"
	githubOAuth2 "golang.org/x/oauth2/github"
)

func registerGithubOAuthRoutes(mux *http.ServeMux, z *zap.Logger, cfg oauth2Config, successHandler http.Handler) error {
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

	mux.Handle("/auth/github/login", setRedirectCookie(github.StateHandler(cfg.cookieConfig(), github.LoginHandler(&oauth2Config, nil))))
	mux.Handle("/auth/github/callback", extractRedirectFromCookie(github.StateHandler(cfg.cookieConfig(), github.CallbackHandler(&oauth2Config, successHandler, nil))))
	return nil
}

func getGithubUserDataFromRequest(req *http.Request) (*externalIdentity, error) {
	ctx := req.Context()
	githubUser, err := github.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return &externalIdentity{
		UserID: strconv.FormatInt(*githubUser.ID, 10),
		Name:   *githubUser.Name,
		Email:  *githubUser.Email,
		Type:   "github",
	}, nil
}
