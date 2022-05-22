package slackeventsrcsvc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/slack-go/slack"

	"github.com/autokitteh/autokitteh/sdk/api/apiproject"
)

func (s *Svc) getOAuthToken(ctx context.Context, pid apiproject.ProjectID, teamID string) ([]byte, error) {
	return s.CredsStore.Get(ctx, pid, "slack", teamID)
}

func (s *Svc) httpOAuthInstall(w http.ResponseWriter, r *http.Request) {
	if s.CredsStore == nil {
		http.Error(w, "not configured for oauth installs", http.StatusNotImplemented)
		return
	}

	ctx := r.Context()
	code := r.FormValue("code")

	resp, err := slack.GetOAuthV2ResponseContext(ctx, &http.Client{}, s.Config.ClientID, s.Config.ClientSecret, code, "")
	if err != nil {
		http.Error(w, fmt.Sprintf("get oauth token error: %v", err), http.StatusInternalServerError)
		return
	}

	s.L.Debug("slack auth", "resp", resp)

	respBytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, fmt.Sprintf("response marshal error: %v", err), http.StatusInternalServerError)
		return
	}

	if err := s.CredsStore.Set(ctx, apiproject.EmptyProjectID, "slack", resp.Team.ID, []byte(resp.AccessToken), string(respBytes)); err != nil {
		http.Error(w, fmt.Sprintf("store error: %v", err), http.StatusInternalServerError)
		return
	}

	// TODO: redirect to configure binding.
	fmt.Fprintf(w, "AutoKitteh Slack App is now installed for %v", resp.Team)
}
