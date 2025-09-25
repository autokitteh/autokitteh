package dashboardsvc

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/web/webdashboard"
)

type membership struct{ sdktypes.OrgMember }

func (p membership) FieldsOrder() []string       { return []string{"org_id", "user_id", "status", "roles"} }
func (p membership) HideFields() []string        { return nil }
func (p membership) ExtraFields() map[string]any { return nil }

func toMembership(sdkM sdktypes.OrgMember) membership { return membership{sdkM} }

func (s *svc) initUsers() {
	s.HandleFunc("GET "+rootPath+"users", s.getUsers)
	s.HandleFunc("POST "+rootPath+"users", s.postUsers)
	s.HandleFunc(rootPath+"users/{id}", s.user)
}

func (s *svc) getUsers(w http.ResponseWriter, r *http.Request) {
	if err := webdashboard.Tmpl(r).ExecuteTemplate(w, "users.html", struct {
		Title string
	}{
		Title: "Users",
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *svc) postUsers(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := sdktypes.SmartParseID[sdktypes.UserID](r.FormValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))

	u, err := s.Users().Get(r.Context(), id, email)
	if err != nil {
		if errors.Is(err, sdkerrors.ErrNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, rootPath+"users/"+u.ID().String(), http.StatusSeeOther)
}

func (s *svc) user(w http.ResponseWriter, r *http.Request) {
	id, err := sdktypes.SmartParseID[sdktypes.UserID](r.PathValue("id"))
	if err != nil {
		http.Error(w, fmt.Sprintf("user id: %v", err), http.StatusBadRequest)
		return
	}

	u, err := s.Users().Get(r.Context(), id, "")
	if err != nil {
		if errors.Is(err, sdkerrors.ErrNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		http.Error(w, fmt.Sprintf("get: %v", err), http.StatusInternalServerError)
		return
	}

	ms, _, err := s.Orgs().GetOrgsForUser(r.Context(), u.ID())
	if err != nil {
		http.Error(w, fmt.Sprintf("get orgs for user: %v", err), http.StatusInternalServerError)
		return
	}

	memberships := genListData(nil, kittehs.Transform(ms, toMembership))

	userJSON := marshalObject(u.ToProto())

	if err := webdashboard.Tmpl(r).ExecuteTemplate(w, "user.html", struct {
		Title       string
		UserJSON    template.HTML
		Memberships any
	}{
		Title:       "User: " + u.ID().String(),
		UserJSON:    userJSON,
		Memberships: memberships,
	}); err != nil {
		http.Error(w, fmt.Sprintf("render user template: %v", err), http.StatusInternalServerError)
	}
}
