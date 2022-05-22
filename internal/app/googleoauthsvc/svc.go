package googleoauthsvc

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/oauth2"

	"github.com/autokitteh/autokitteh/sdk/api/apiproject"
	"github.com/autokitteh/autokitteh/internal/pkg/credsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/googleoauth"
	"github.com/autokitteh/autokitteh/pkg/kvstore"
	L "github.com/autokitteh/autokitteh/pkg/l"
)

type Config = googleoauth.Config

type Svc struct {
	Config          Config
	CredsStore      *credsstore.Store
	OAuthStateStore kvstore.Store

	L L.Nullable

	oauth2Config *oauth2.Config
}

type oauthState struct {
	ProjectID apiproject.ProjectID `json:"project_id"`
	Name      string               `json:"nam"`
}

func (s *Svc) Register(r *mux.Router) {
	s.oauth2Config = googleoauth.MakeConfig(s.Config)

	r.HandleFunc("/google-oauth/projects/{pid}/{name}/oauth/install", s.httpOAuthInstall).Methods("GET")
	r.HandleFunc("/google-oauth/oauth/installed", s.httpOAuthInstalled).Methods("GET")
}

func (s *Svc) httpOAuthInstall(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, name := apiproject.ProjectID(vars["pid"]), vars["name"]

	bs := make([]byte, 16)
	if _, err := rand.Read(bs); err != nil {
		http.Error(w, fmt.Sprintf("rand read: %v", err), http.StatusInternalServerError)
		return
	}

	code := hex.EncodeToString(bs)

	state, err := json.Marshal(oauthState{ProjectID: pid, Name: name})
	if err != nil {
		panic(err)
	}

	if err := s.OAuthStateStore.Put(r.Context(), code, state); err != nil {
		http.Error(w, fmt.Sprintf("state put: %v", err), http.StatusInternalServerError)
		return
	}

	// Both oauth2.AccessTypeOffline and approval_prompt=force are required
	// for refresh token to be generated.
	// See https://stackoverflow.com/questions/42707791/golang-google-drive-oauth2-not-returning-refresh-token.
	url := s.oauth2Config.AuthCodeURL(code, oauth2.AccessTypeOffline)
	url += "&approval_prompt=force"

	http.Redirect(w, r, url, http.StatusFound)
}

func (s *Svc) httpOAuthInstalled(w http.ResponseWriter, r *http.Request) {
	encodedState, err := s.OAuthStateStore.Get(r.Context(), r.URL.Query().Get("state"))
	if err != nil {
		http.Error(w, fmt.Sprintf("get state: %v", err), http.StatusForbidden)
		return
	}

	code := r.URL.Query().Get("code")

	if err := s.OAuthStateStore.Put(r.Context(), code, nil); err != nil {
		s.L.Error("state store delete error", "code", code, "err", err)

		// no abort
	}

	var state oauthState
	if err := json.Unmarshal(encodedState, &state); err != nil {
		http.Error(w, fmt.Sprintf("unmarshal state: %v", err), http.StatusInternalServerError)
		return
	}

	tok, err := s.oauth2Config.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, fmt.Sprintf("oauth installed error: %v", err), http.StatusInternalServerError)
		return
	}

	bs, err := json.Marshal(tok)
	if err != nil {
		http.Error(w, fmt.Sprintf("token marshal error: %v", err), http.StatusInternalServerError)
		return
	}

	if err := s.CredsStore.Set(r.Context(), state.ProjectID, "googleoauth", state.Name, bs, nil); err != nil {
		http.Error(w, fmt.Sprintf("store error: %v", err), http.StatusInternalServerError)
		return
	}

	// TODO: redirect to configure binding.
	fmt.Fprintf(w, "AutoKitteh Google Sheets App is now installed for project %v", state.ProjectID)
}
