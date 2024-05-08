package authloginhttpsvc

import (
	_ "embed"
	"errors"
	"net/http"

	"github.com/descope/go-sdk/descope/client"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authloginhttpsvc/web"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func registerDescopeRoutes(mux *http.ServeMux, cfg descopeConfig, onSuccess func(sdktypes.User) http.Handler) error {
	if cfg.ProjectID == "" {
		return errors.New("descope login is enabled, but missing DESCOPE_PROJECT_ID")
	}

	client, err := client.NewWithConfig(&client.Config{ProjectID: cfg.ProjectID})
	if err != nil {
		return err
	}

	mux.Handle("/auth/descope/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := web.DescopeLoginTemplate.Execute(w, struct{ ProjectID string }{ProjectID: cfg.ProjectID}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))

	mux.Handle("/auth/descope/loggedin", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorized, _, err := client.Auth.ValidateAndRefreshSessionWithRequest(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !authorized {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		user, err := client.Auth.Me(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		details := map[string]string{
			"id":    user.UserID,
			"email": user.Email,
			"name":  user.Name,
		}

		onSuccess(sdktypes.NewUser("descope", details)).ServeHTTP(w, r)
	}))

	return nil
}
