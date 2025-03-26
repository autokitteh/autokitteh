package authsessions

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	sessionName    = "ak_user_session"
	loggedInCookie = "ak_logged_in"
)

type store struct {
	// store    sessions.Store[[]byte]
	domain     string
	secure     bool
	sameSite   http.SameSite
	tokens     authtokens.Tokens
	expiration time.Duration
}

type Store interface {
	Set(http.ResponseWriter, sdktypes.User) error
	Get(*http.Request) (sdktypes.User, error)
	Delete(http.ResponseWriter)
}

func New(cfg *Config, tokens authtokens.Tokens) (Store, error) {
	domain := cfg.Domain
	if len(domain) > 0 && !strings.HasPrefix(domain, ".") {
		domain = "." + cfg.Domain
	}

	return &store{
		domain:     domain,
		secure:     cfg.Secure,
		sameSite:   cfg.SameSite,
		tokens:     tokens,
		expiration: time.Duration(cfg.ExpirationMinutes) * time.Minute,
	}, nil
}

func (s store) Set(w http.ResponseWriter, user sdktypes.User) error {
	jwt, err := s.tokens.Create(user)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionName,
		Value:    jwt,
		Path:     "/",
		Domain:   s.domain,
		SameSite: s.sameSite,
		Secure:   s.secure,
		HttpOnly: true,
		Expires:  time.Now().Add(s.expiration),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     loggedInCookie,
		Value:    "true",
		Path:     "/",
		Domain:   s.domain,
		SameSite: s.sameSite,
		Secure:   s.secure,
		Expires:  time.Now().Add(s.expiration),
	})

	return nil
}

func (s store) Get(req *http.Request) (sdktypes.User, error) {
	cookie, err := req.Cookie(loggedInCookie)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return sdktypes.InvalidUser, nil
		}
		return sdktypes.InvalidUser, err
	}

	if cookie.Value != "true" {
		return sdktypes.InvalidUser, errors.New("invalid logged in cookie")
	}

	sessionCookie, err := req.Cookie(sessionName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return sdktypes.InvalidUser, nil
		}
		return sdktypes.InvalidUser, err
	}

	user, err := s.tokens.Parse(sessionCookie.Value)
	if err != nil {
		return sdktypes.InvalidUser, err
	}

	return user, nil
}

func (s *store) Delete(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:    sessionName,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	})
	http.SetCookie(w, &http.Cookie{
		Name:    loggedInCookie,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	})
}
