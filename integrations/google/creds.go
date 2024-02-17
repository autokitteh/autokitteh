package google

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"go.uber.org/zap"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

const (
	// credsPath is the URL path for our handler to save a new autokitteh
	// credentials-based connection, after the user submits it via a web form.
	credsPath = "/google/save_json"

	HeaderContentType = "Content-Type"
	ContentTypeForm   = "application/x-www-form-urlencoded"
)

// HandleCreds saves a new autokitteh connection with a user-submitted token.
func (h handler) HandleCreds(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", r.URL.Path))

	// Check "Content-Type" header.
	ct := r.Header.Get(HeaderContentType)
	if ct != ContentTypeForm {
		l.Warn("Unexpected header value",
			zap.String("header", HeaderContentType),
			zap.String("got", ct),
			zap.String("want", ContentTypeForm),
		)
		e := fmt.Sprintf("Unexpected Content-Type header: %q", ct)
		u := fmt.Sprintf("%serror.html?error=%s", uiPath, url.QueryEscape(e))
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	// Read and parse POST request body.
	err := r.ParseForm()
	if err != nil {
		l.Warn("Failed to parse inbound HTTP request",
			zap.Error(err),
		)
		e := "Form parsing error: " + err.Error()
		u := fmt.Sprintf("%serror.html?error=%s", uiPath, url.QueryEscape(e))
		http.Redirect(w, r, u, http.StatusFound)
		return
	}
	creds := r.Form.Get("json")

	// Test the PAT's usability and get authoritative metadata details.
	opt := option.WithCredentialsJSON([]byte(creds))
	svc, err := googleoauth2.NewService(r.Context(), opt)
	if err != nil {
		l.Warn("Service account credentials error",
			zap.Error(err),
		)
		e := "Service account credentials error: " + err.Error()
		u := fmt.Sprintf("%serror.html?error=%s", uiPath, url.QueryEscape(e))
		http.Redirect(w, r, u, http.StatusFound)
		return
	}
	ui, ti, err := h.getUserDetails(r.Context(), w, svc)
	if err != nil {
		e := "Unusable service account credentials error: " + err.Error()
		u := fmt.Sprintf("%serror.html?error=%s", uiPath, url.QueryEscape(e))
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	// Save a new connection, and return to the user an autokitteh connection token.
	connToken, err := h.createCredsConnection(ui, ti, creds)
	if err != nil {
		l.Warn("Failed to save new connection secrets",
			zap.Error(err),
		)
		e := "Connection saving error: " + err.Error()
		u := fmt.Sprintf("%serror.html?error=%s", uiPath, url.QueryEscape(e))
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	// Redirect the user to a success page: give them the connection token.
	l.Debug("Saved new autokitteh connection")
	u := fmt.Sprintf("%ssuccess.html?token=%s", uiPath, connToken)
	http.Redirect(w, r, u, http.StatusFound)
}

func (h handler) getUserDetails(ctx context.Context, w http.ResponseWriter, svc *googleoauth2.Service) (*googleoauth2.Userinfo, *googleoauth2.Tokeninfo, error) {
	ui, err := svc.Userinfo.V2.Me.Get().Do()
	if err != nil {
		h.logger.Warn("OAuth user info retrieval error",
			zap.Error(err),
		)
		return nil, nil, err
	}
	if ui.Email == "" || !*ui.VerifiedEmail {
		h.logger.Warn("OAuth user info is bad",
			zap.Any("userInfo", ui),
			zap.Error(err),
		)
		return nil, nil, err
	}

	ti, err := svc.Tokeninfo().Do()
	if err != nil {
		h.logger.Warn("OAuth token info retrieval error",
			zap.Any("userInfo", ui),
			zap.Error(err),
		)
		return nil, nil, err
	}
	if ti.Email != ui.Email {
		h.logger.Warn("OAuth token info is bad",
			zap.Any("userInfo", ui),
			zap.Any("tokenInfo", ti),
			zap.Error(err),
		)
		return nil, nil, err
	}
	return ui, ti, nil
}

func (h handler) createCredsConnection(u *googleoauth2.Userinfo, t *googleoauth2.Tokeninfo, creds string) (string, error) {
	token, err := h.secrets.Create(context.Background(), h.scope,
		// Connection token --> JSON credentials (to call API methods).
		map[string]string{
			// Google.
			"userID": u.Id,
			"email":  u.Email,
			"scopes": t.Scope,
			// Service account credentials.
			"JSON": creds,
		},
		// Google user ID --> connection token(s) (to dispatch API events).
		fmt.Sprintf("users/%s", u.Id),
	)
	if err != nil {
		return "", err
	}
	return token, nil
}
