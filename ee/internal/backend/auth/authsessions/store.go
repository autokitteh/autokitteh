package authsessions

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"

	"github.com/dghubble/sessions"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	sessionName        = "ak_user_session"
	sessionDataKeyName = "ak_data"
)

type SessionData struct {
	UserID sdktypes.UserID
}

type store struct {
	store sessions.Store[SessionData]
}

type Store interface {
	Set(http.ResponseWriter, *SessionData) error
	Get(*http.Request) (*SessionData, error)
	Delete(http.ResponseWriter)
}

func New(cfg *Config) (Store, error) {
	keyPairs, err := kittehs.TransformError(cfg.CookieKeys, hex.DecodeString)
	if err != nil {
		return nil, fmt.Errorf("invalid key pairs: %w", err)
	}

	if len(keyPairs)&1 != 0 {
		return nil, errors.New("key pairs must be an even length list of hex encoded keys")
	}

	return &store{
		store: sessions.NewCookieStore[SessionData](cfg.cookieConfig(), keyPairs...),
	}, nil
}

func (s store) Set(w http.ResponseWriter, data *SessionData) error {
	session := s.store.New(sessionName)
	session.Set(sessionDataKeyName, *data)
	return session.Save(w)
}

func (s store) Get(req *http.Request) (*SessionData, error) {
	session, err := s.store.Get(req, sessionName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, nil
		}
		return nil, err
	}

	usd := session.Get(sessionDataKeyName)
	return &usd, nil
}

func (s *store) Delete(w http.ResponseWriter) {
	s.store.Destroy(w, sessionName)
}
