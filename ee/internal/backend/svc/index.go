package svc

import (
	_ "embed"
	"net/http"
	"text/template"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/ee/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

//go:embed index.html
var indexContent string

var indexTemplate = kittehs.Must1(template.New("ee_index.html").Parse(indexContent))

func indexOption() fx.Option {
	return fx.Invoke(func(z *zap.Logger, mux *http.ServeMux, users sdkservices.Users) {
		mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userName string

			ctx := r.Context()

			userID := authcontext.GetAuthnUserID(ctx)
			if userID.IsValid() {
				user, err := users.GetByID(ctx, userID)
				if err != nil {
					z.Error("get user error", zap.String("user_id", userID.String()), zap.Error(err))
					http.Error(w, "get user error", http.StatusInternalServerError)
					return
				}

				userName = user.Name().String()
			}

			kittehs.Must0(indexTemplate.Execute(w, map[string]any{"userName": userName}))
		}))
	})
}
