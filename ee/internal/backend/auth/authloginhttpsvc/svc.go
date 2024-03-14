package authloginhttpsvc

import (
	_ "embed"
	"errors"
	"net/http"

	"go.autokitteh.dev/autokitteh/ee/internal/backend/auth/authsessions"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/db"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/muxes"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

//go:embed login.html
var loginHTML []byte

type externalIdentity struct {
	Name   string
	UserID string
	Email  string
	Type   string
}

type externalIDDataExtractorFunc func(req *http.Request) (*externalIdentity, error)

type Deps struct {
	fx.In

	Muxes    muxes.Muxes
	Z        *zap.Logger
	Cfg      *Config
	DB       db.DB
	Sessions authsessions.Store
	Users    sdkservices.Users
	Tokens   authtokens.Tokens
}

type svc struct {
	Deps

	deviceCodeHandler deviceCodeHandler
}

func Init(deps Deps) error {
	svc := &svc{
		Deps: deps,
		deviceCodeHandler: deviceCodeHandler{
			logger:   deps.Z,
			sessions: deps.Sessions,
			tokens:   deps.Tokens,
			codes:    make(map[string]string, 16),
		},
	}

	return svc.registerRoutes(deps.Muxes)
}

func (a *svc) registerRoutes(muxes muxes.Muxes) error {
	if a.Cfg.GoogleOAuth.Enabled {
		if err := registerGoogleOAuthRoutes(muxes.NoAuth, a.Z, a.Deps.Cfg.GoogleOAuth, a.newSuccessLoginHandler(getGoogleUserDataFromRequest)); err != nil {
			return err
		}
	}

	if a.Cfg.GithubOAuth.Enabled {
		if err := registerGithubOAuthRoutes(muxes.NoAuth, a.Z, a.Cfg.GithubOAuth, a.newSuccessLoginHandler(getGithubUserDataFromRequest)); err != nil {
			return err
		}
	}

	a.deviceCodeHandler.registerRoutes(muxes)

	muxes.NoAuth.HandleFunc("/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			a.Deps.Sessions.Delete(w)
		}
		http.Redirect(w, r, "/", http.StatusFound)
	})

	muxes.NoAuth.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(loginHTML)
	})

	return nil
}

func (a *svc) newSuccessLoginHandler(externalIDDataExtractor externalIDDataExtractorFunc) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		externalID, err := externalIDDataExtractor(req)
		if err != nil {
			a.Z.Error("failed extracting extenal id", zap.Error(err))
			http.Error(w, "could not indentify user", http.StatusInternalServerError)
			return
		}

		z := a.Z.With(zap.String("external_user_id", externalID.UserID))

		u, err := a.DB.GetUserByExternalID(req.Context(), externalID.UserID)
		if err != nil && !errors.Is(err, sdkerrors.ErrNotFound) {
			z.Error("failed getting user from db", zap.Error(err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if !u.IsValid() {
			z.Warn("user not found")
			http.Error(w, "user not found", http.StatusForbidden)
			return

			/* TODO(ENG-300):
			   should we add a user when not found? Maybe not,
			   because we don't know if the user is allowed to login,
			   and we don't know if the user is allowed to create an account.

			if u, err = sdktypes.UserFromProto(&sdktypes.UserPB{
				Name: strcase.ToCamel(externalID.Name),
			}); err != nil {
				a.Z.Error("failed user from proto", zap.Error(err))
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			err = a.DB.Transaction(req.Context(), func(tx db.DB) error {
				userID, err := a.Users.Create(req.Context(), u)
				if err != nil {
					return err
				}

				u = kittehs.Must1(u.Update(func(pb *sdktypes.UserPB) { pb.UserId = userID.String() }))

				return a.DB.AddExternalIDToUser(req.Context(), userID, externalID.UserID, externalID.Type, externalID.Email)
			})
			if err != nil {
				a.Z.Error("failed creating user", zap.Error(err))
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			*/
		}

		sd := authsessions.SessionData{UserID: u.ID()}

		if err := a.Deps.Sessions.Set(w, &sd); err != nil {
			z.Error("failed storing session", zap.Error(err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		redirect := "/"
		if cookie, _ := req.Cookie(redirectCookieName); cookie != nil {
			redirect = cookie.Value
		}

		http.Redirect(w, req, redirect, http.StatusFound)
	}

	return http.HandlerFunc(fn)
}
