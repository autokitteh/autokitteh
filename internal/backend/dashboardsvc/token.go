package dashboardsvc

import (
	"net/http"

	"go.autokitteh.dev/autokitteh/web/webdashboard"
)

func (s Svc) initToken() {
	s.Muxes.Auth.HandleFunc("GET /dashboard/token", s.token)
}

func (s Svc) token(w http.ResponseWriter, r *http.Request) {
	tok, err := s.Svcs.Auth().CreateToken(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := webdashboard.Tmpl(r).ExecuteTemplate(w, "token.html", struct {
		Message string
		Title   string
		Token   string
	}{
		Title: "Generate Token",
		Token: tok,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
