package dashboardsvc

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/web/webdashboard"
)

func (s Svc) initAuth() {
	s.Muxes.Auth.HandleFunc("/auth/{$}", s.auth)
	s.Muxes.Auth.HandleFunc("POST /auth/tokens", s.createToken)
}

func (s Svc) auth(w http.ResponseWriter, r *http.Request) {
	userJSON := kittehs.Must1(json.Marshal(authcontext.GetAuthnUser(r.Context())))

	if err := webdashboard.Tmpl(r).ExecuteTemplate(w, "auth.html", struct {
		Title    string
		UserJSON string
		Token    string
	}{
		Title:    "Auth",
		UserJSON: string(userJSON),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s Svc) createToken(w http.ResponseWriter, r *http.Request) {
	tok, err := s.Svcs.Auth().CreateToken(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%q", tok)
}
