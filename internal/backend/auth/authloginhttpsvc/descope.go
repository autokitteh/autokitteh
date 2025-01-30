package authloginhttpsvc

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/http"

	"github.com/descope/go-sdk/descope/client"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authloginhttpsvc/web"
)

const (
	descopeRootPath  = "/auth/descope"
	descopeLoginPath = descopeRootPath + "/login"
)

func registerDescopeRoutes(mux *http.ServeMux, cfg descopeConfig, onSuccess func(context.Context, *loginData) http.Handler) error {
	if cfg.ProjectID == "" {
		return errors.New("descope login is enabled, but missing DESCOPE_PROJECT_ID")
	}

	client, err := client.NewWithConfig(&client.Config{ProjectID: cfg.ProjectID})
	if err != nil {
		return err
	}

	mux.HandleFunc(descopeRootPath+"/project_id", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "%q", cfg.ProjectID)
	})

	mux.HandleFunc(descopeLoginPath, func(w http.ResponseWriter, r *http.Request) {
		jwt := r.URL.Query().Get("jwt")

		if jwt == "" {
			if err := web.DescopeLoginTemplate.Execute(w, struct{ ProjectID string }{ProjectID: cfg.ProjectID}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			return
		}

		ctx := r.Context()

		// post-login

		authorized, tok, err := client.Auth.ValidateSessionWithToken(ctx, jwt)
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

		ld := loginData{
			ProviderName: "descope",
			Email:        email,
			DisplayName:  name,
		}

		onSuccess(ctx, &ld).ServeHTTP(w, r)
	})

	return nil
}
