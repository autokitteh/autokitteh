package authloginhttpsvc

import (
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/ee/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/auth/authsessions"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/ee/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

//go:embed code.html
var codeHTML []byte

var codeTemplate = template.Must(template.New("code").Parse(string(codeHTML)))

type deviceCodeHandler struct {
	logger   *zap.Logger
	sessions authsessions.Store
	tokens   authtokens.Tokens

	// TODO: put into redis?
	codes map[string]string
	mu    sync.Mutex
}

func (d *deviceCodeHandler) registerRoutes(muxes muxes.Muxes) {
	muxes.Auth.HandleFunc("/auth/device/code", d.deviceCodeHandler)
	muxes.NoAuth.HandleFunc("/auth/device/exchange", d.exchangeCodeHandler)
}

func (d *deviceCodeHandler) newCode() string { return uuid.NewString() }

func (d *deviceCodeHandler) deviceCodeHandler(w http.ResponseWriter, r *http.Request) {
	userID := authcontext.GetAuthnUserID(r.Context())
	if !userID.IsValid() {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}

	if err := d.sessions.Set(w, &authsessions.SessionData{UserID: userID}); err != nil {
		http.Redirect(w, r, "/error.html?err="+err.Error(), http.StatusTemporaryRedirect)
		return
	}

	code := d.newCode() // TODO: replace with short id easy to copy
	d.setCode(code, userID)

	data := struct{ Code string }{Code: code}

	if err := codeTemplate.Execute(w, data); err != nil {
		http.Redirect(w, r, "/error.html?err="+err.Error(), http.StatusTemporaryRedirect)
	}
}

func (d *deviceCodeHandler) exchangeCodeHandler(w http.ResponseWriter, r *http.Request) {
	l := d.logger

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	if !q.Has("code") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	code := q.Get("code")

	userID, err := d.exchangeDeviceCodeForUserID(r.Context(), code)
	if err != nil {
		l.Error("failed to exchange code for user id", zap.Error(err))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := d.tokens.Create(userID)
	if err != nil {
		l.Error("failed to create token", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO: encrypt token for CLI use. Does it need to be encrypted?

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"token": "%s"}`, token)

	// TODO: remove code for codeToSession to prevent reuse
	// still here for testing purposes
}

func (d *deviceCodeHandler) setCode(code string, userID sdktypes.UserID) {
	// TODO: this wont work with multiple instances.
	d.mu.Lock()
	defer d.mu.Unlock()
	d.codes[code] = userID.String()
}

func (d *deviceCodeHandler) exchangeDeviceCodeForUserID(ctx context.Context, code string) (sdktypes.UserID, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	token, ok := d.codes[code]
	if !ok {
		return sdktypes.InvalidUserID, sdkerrors.ErrUnauthorized
	}

	delete(d.codes, code)

	return sdktypes.ParseUserID(token)
}
