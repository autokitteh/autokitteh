package authloginhttpsvc

import (
	"context"
	_ "embed"
	"errors"
	"net/http"

	"github.com/descope/go-sdk/descope/client"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const descopeLoginPath = "/auth/descope/login"

func registerDescopeRoutes(mux *http.ServeMux, cfg descopeConfig, onSuccess func(sdktypes.User) http.Handler) error {
	if cfg.ProjectID == "" {
		return errors.New("descope login is enabled, but missing DESCOPE_PROJECT_ID")
	}

	client, err := client.NewWithConfig(&client.Config{ProjectID: cfg.ProjectID, ManagementKey: cfg.ManagementKey})
	if err != nil {
		return err
	}

	mux.Handle(descopeLoginPath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url, err := client.Auth.OAuth().SignUpOrIn(context.Background(), "google", "", nil, nil, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, url, http.StatusFound)
	}))

	mux.Handle("/auth/descope/loggedin", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authInfo, err := client.Auth.OAuth().ExchangeToken(context.Background(), r.URL.Query().Get("code"), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		details := map[string]string{
			"email": authInfo.User.Email,
			"name":  authInfo.User.Name,
		}

		onSuccess(sdktypes.NewUser("descope", details)).ServeHTTP(w, r)
	}))

	return nil
}
