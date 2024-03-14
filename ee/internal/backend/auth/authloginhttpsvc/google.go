package authloginhttpsvc

import (
	"errors"
	"net/http"

	"github.com/dghubble/gologin/v2/google"
	"go.uber.org/zap"

	"golang.org/x/oauth2"
	googleOAuth2 "golang.org/x/oauth2/google"
)

func registerGoogleOAuthRoutes(mux *http.ServeMux, z *zap.Logger, cfg oauth2Config, successHandler http.Handler) error {
	if cfg.ClientID == "" {
		return errors.New("google login is enabled, but missing GOOGLE_CLIENT_ID")
	}

	if cfg.ClientSecret == "" {
		return errors.New("google login is enabled, but missing GOOGLE_CLIENT_SECRET")
	}

	if cfg.RedirectURL == "" {
		return errors.New("GOOGLE_REDIRECT_URL not defined, using localhost callback")
	}

	oauth2Config := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     googleOAuth2.Endpoint,
		Scopes:       []string{"profile", "email"},
	}

	mux.Handle("/auth/google/login", setRedirectCookie(google.StateHandler(cfg.cookieConfig(), google.LoginHandler(&oauth2Config, nil))))
	mux.Handle("/auth/google/callback", extractRedirectFromCookie(google.StateHandler(cfg.cookieConfig(), google.CallbackHandler(&oauth2Config, successHandler, nil))))
	return nil
}

func getGoogleUserDataFromRequest(req *http.Request) (*externalIdentity, error) {
	ctx := req.Context()
	googleUser, err := google.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return &externalIdentity{
		UserID: googleUser.Id,
		Name:   googleUser.Name,
		Email:  googleUser.Email,
		Type:   "google",
	}, nil
}
