package authloginhttpsvc

import (
	_ "embed"
	"errors"
	"maps"
	"net/http"

	"github.com/descope/go-sdk/descope/client"
	j "github.com/golang-jwt/jwt/v5"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authloginhttpsvc/web"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func registerDescopeRoutes(mux *http.ServeMux, cfg descopeConfig, onSuccess func(sdktypes.User) http.Handler) error {
	if cfg.ProjectID == "" {
		return errors.New("descope login is enabled, but missing DESCOPE_PROJECT_ID")
	}

	client, err := client.NewWithConfig(&client.Config{ProjectID: cfg.ProjectID, ManagementKey: cfg.ManagementKey})
	if err != nil {
		return err
	}

	mux.Handle("/auth/descope/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := web.DescopeLoginTemplate.Execute(w, struct{ ProjectID string }{ProjectID: cfg.ProjectID}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))

	mux.Handle("/auth/descope/loggedin", extractRedirectFromCookie(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorized, tok, err := client.Auth.ValidateAndRefreshSessionWithRequest(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !authorized {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var claims j.RegisteredClaims
		if _, _, err = j.NewParser().ParseUnverified(tok.JWT, &claims); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		details := map[string]string{
			"id": claims.Subject,
		}

		if cfg.ManagementKey != "" {
			u, err := client.Management.User().LoadByUserID(r.Context(), claims.Subject)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			maps.Copy(details, map[string]string{
				"email": u.Email,
				"name":  u.Name,
			})
		}

		onSuccess(sdktypes.NewUser("descope", details)).ServeHTTP(w, r)
	})))

	return nil
}
