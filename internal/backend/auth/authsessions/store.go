package authsessions

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dghubble/sessions"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	sessionName        = "ak_user_session"
	sessionDataKeyName = "ak_data"
	loggedInCookie     = "ak_logged_in"
)

type sessionData struct {
	User      sdktypes.User
	Validator int64
}

func NewSessionData(user sdktypes.User) sessionData {
	return sessionData{
		User:      user,
		Validator: time.Now().Unix(),
	}
}

type store struct {
	store sessions.Store[[]byte]
}

type Store interface {
	Set(http.ResponseWriter, *sessionData) error
	Get(*http.Request) (*sessionData, error)
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

func (s store) newSessionWithData(data *sessionData) (*sessions.Session[[]byte], error) {
	session := s.store.New(sessionName)

	bs, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user data: %w", err)
	}

	session.Set(sessionDataKeyName, bs)
	return session, nil
}

func (s store) Set(w http.ResponseWriter, data *sessionData) error {
	session, err := s.newSessionWithData(data)
	if err != nil {
		return err
	}

	if err := session.Save(w); err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:  loggedInCookie,
		Value: fmt.Sprintf("%d", data.Validator),
		Path:  "/",
	})

	return nil
}

func (s store) Get(req *http.Request) (*sessionData, error) {
	loggedIn, err := req.Cookie(loggedInCookie)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, nil
		}
		return nil, err
	}

	validator, err := strconv.ParseInt(loggedIn.Value, 10, 64)
	if err != nil {
		return nil, errors.New("invalid logged in cookie")
	}

	session, err := s.store.Get(req, sessionName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return nil, nil
		}
		return nil, err
	}

	bs := session.Get(sessionDataKeyName)

	var sd sessionData
	if err := json.Unmarshal(bs, &sd); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	if validator != sd.Validator {
		return nil, errors.New("invalid logged in cookie")
	}

	return &sd, nil
}

func (s *store) Delete(w http.ResponseWriter) {
	s.store.Destroy(w, sessionName)
	http.SetCookie(w, &http.Cookie{
		Name:    loggedInCookie,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	})
}
