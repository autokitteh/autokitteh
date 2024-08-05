package authsessions

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/dghubble/sessions"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	sessionName        = "ak_user_session"
	sessionDataKeyName = "ak_data"
	loggedInCookie     = "ak_logged_in"
)

type SessionData struct {
	User sdktypes.User
}

type store struct {
	store sessions.Store[[]byte]
}

type Store interface {
	Set(http.ResponseWriter, *SessionData) error
	Get(*http.Request) (*SessionData, error)
	Delete(http.ResponseWriter)
}

func New(cfg *Config) (Store, error) {
	rawKeyPairs := kittehs.Transform(strings.Split(cfg.CookieKeys, ","), strings.TrimSpace)
	if len(rawKeyPairs)&1 != 0 {
		return nil, errors.New("key pairs must be an even length list of hex encoded keys")
	}

	keyPairs, err := kittehs.TransformError(rawKeyPairs, hex.DecodeString)
	if err != nil {
		return nil, fmt.Errorf("invalid key pairs: %w", err)
	}

	return &store{
		store: sessions.NewCookieStore[[]byte](cfg.Cookie, keyPairs...),
	}, nil
}

func (s store) Set(w http.ResponseWriter, data *SessionData) error {
	session := s.store.New(sessionName)

	bs, err := json.Marshal(data.User)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	session.Set(sessionDataKeyName, bs)

	if err := session.Save(w); err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:  loggedInCookie,
		Value: "true",
		Path:  "/",
	})

	return nil
}

func (s store) Get(req *http.Request) (*SessionData, error) {
	loggedIn, err := req.Cookie(loggedInCookie)
	if err != nil {
		return nil, err
	}

	if loggedIn == nil {
		return nil, errors.New("logged in cookie missing")
	}

	session, err := s.store.Get(req, sessionName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, nil
		}
		return nil, err
	}

	bs := session.Get(sessionDataKeyName)

	var sd SessionData
	if err := json.Unmarshal(bs, &sd.User); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	return &sd, nil
}

func (s *store) Delete(w http.ResponseWriter) {
	s.store.Destroy(w, sessionName)
}
