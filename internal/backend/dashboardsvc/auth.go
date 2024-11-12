package dashboardsvc

import (
	"fmt"
	"html/template"
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/web/webdashboard"
)

func (s *svc) initAuth() {
	s.HandleFunc(rootPath+"auth/{$}", s.auth)
	s.HandleFunc("POST "+rootPath+"auth/tokens", s.createToken)
}

func (s *svc) auth(w http.ResponseWriter, r *http.Request) {
	if err := webdashboard.Tmpl(r).ExecuteTemplate(w, "auth.html", struct {
		Title    string
		UserJSON template.HTML
		Token    string
	}{
		Title:    "Auth",
		UserJSON: marshalObject(authcontext.GetAuthnUser(r.Context()).ToProto()),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *svc) createToken(w http.ResponseWriter, r *http.Request) {
	tok, err := s.Auth().CreateToken(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%q", tok)
}
