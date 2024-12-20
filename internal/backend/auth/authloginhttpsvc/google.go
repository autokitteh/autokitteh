package authloginhttpsvc

import (
	"context"
	"errors"
	"net/http"

	"github.com/dghubble/gologin/v2/google"

	"golang.org/x/oauth2"
	googleOAuth2 "golang.org/x/oauth2/google"
)

const googleLoginPath = "/auth/google/login"

func registerGoogleOAuthRoutes(mux *http.ServeMux, cfg oauth2Config, onSuccess func(context.Context, *loginData) http.Handler) error {
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

	mux.Handle(googleLoginPath, google.StateHandler(cfg.cookieConfig(), google.LoginHandler(&oauth2Config, nil)))

	googleOnSuccess := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		gu, err := google.UserFromContext(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if gu.Id == "" || gu.Name == "" || gu.Email == "" {
			http.Error(w, "google user missing data", http.StatusInternalServerError)
			return
		}

		ld := loginData{
			ProviderName: "google",
			Email:        gu.Email,
			DisplayName:  gu.Name,
		}

		onSuccess(ctx, &ld).ServeHTTP(w, r)
	})

	mux.Handle("/auth/google/callback", google.StateHandler(cfg.cookieConfig(), google.CallbackHandler(&oauth2Config, googleOnSuccess, nil)))
	return nil
}
