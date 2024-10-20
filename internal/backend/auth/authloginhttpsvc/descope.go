package authloginhttpsvc

import (
	_ "embed"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/descope/go-sdk/descope/client"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authloginhttpsvc/web"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const descopeLoginPath = "/auth/descope/login"

func registerDescopeRoutes(mux *http.ServeMux, cfg descopeConfig, onSuccess func(sdktypes.UserAuthProvider) http.Handler) error {
	if cfg.ProjectID == "" {
		return errors.New("descope login is enabled, but missing DESCOPE_PROJECT_ID")
	}

	client, err := client.NewWithConfig(&client.Config{ProjectID: cfg.ProjectID, ManagementKey: cfg.ManagementKey})
	if err != nil {
		return err
	}

	mux.Handle(descopeLoginPath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwt := r.URL.Query().Get("jwt")

		if jwt == "" {
			if err := web.DescopeLoginTemplate.Execute(w, struct{ ProjectID string }{ProjectID: cfg.ProjectID}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			return
		}

		// post-login

		authorized, tok, err := client.Auth.ValidateSessionWithToken(r.Context(), jwt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !authorized {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		name, _ := tok.CustomClaim("name").(string)
		email, _ := tok.CustomClaim("email").(string)

		if email == "" {
			http.Error(w, "email is required in jwt claims", http.StatusBadRequest)
			return
		}

		details, err := json.Marshal(map[string]string{
			"name":  name,
			"email": email,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		p, err := sdktypes.NewUserAuthProvider(
			"descope",
			email,
			email,
			details,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		onSuccess(p).ServeHTTP(w, r)
	}))

	return nil
}
