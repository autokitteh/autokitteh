package authsessions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authtokens/authjwttokens"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var testUID = sdktypes.NewUserID()

func newStore(t *testing.T) Store {
	cfg := Configs.Dev
	tokens, _ := authjwttokens.New(authjwttokens.Configs.Dev)

	s, err := New(cfg, tokens)
	require.NoError(t, err)
	return s
}

func newRequestWithLoggedInCookie(value string) http.Request {
	return http.Request{
		Header: http.Header{
			"Cookie": []string{"ak_logged_in=" + value},
		},
	}
}

func TestStoreGetNoLoggedinCookie(t *testing.T) {
	s := newStore(t)
	r := http.Request{}
	sd, err := s.Get(&r)
	assert.Equal(t, sd, sdktypes.InvalidUser)
	assert.Nil(t, err)
}

func TestCookieIsHttpOnly(t *testing.T) {
	s := newStore(t)
	w := httptest.NewRecorder()

	err := s.Set(w, sdktypes.NewUser().WithID(testUID))
	require.Nil(t, err)
	cookies := w.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == sessionName {
			require.True(t, cookie.HttpOnly)
			break
		}
	}
}

func TestStoreGetInvalidLoggedInCookie(t *testing.T) {
	s := newStore(t)
	r := http.Request{
		Header: http.Header{
			"Cookie": []string{"ak_logged_in=k1k2"},
		},
	}
	sd, err := s.Get(&r)
	assert.Equal(t, sd, sdktypes.InvalidUser)
	assert.Error(t, err, "invalid logged in cookie")
}

func TestStoreGetNoSessionCookie(t *testing.T) {
	s := newStore(t)
	r := newRequestWithLoggedInCookie("true")
	sd, err := s.Get(&r)
	assert.Equal(t, sd, sdktypes.InvalidUser)
	assert.Nil(t, err)
}

func TestStoreGetInvalidSessionCookie(t *testing.T) {
	s := newStore(t)
	r := newRequestWithLoggedInCookie(uuid.NewString())

	r.AddCookie(&http.Cookie{
		Name:  sessionName,
		Value: "invalid_value",
	})

	sd, err := s.Get(&r)
	assert.Equal(t, sd, sdktypes.InvalidUser)
	assert.Error(t, err)
}

func TestStoreGetValidatorMismatch(t *testing.T) {
	s := newStore(t)
	w := httptest.NewRecorder()
	u := sdktypes.NewUser().WithID(testUID)
	err := s.Set(w, u)
	assert.Nil(t, err)

	otherValidator := uuid.NewString()
	r := newRequestWithLoggedInCookie(otherValidator)
	r.AddCookie(w.Result().Cookies()[0])

	sd, err := s.Get(&r)
	assert.Equal(t, sd, sdktypes.InvalidUser)
	assert.Error(t, err)
}

func TestStoreSetGetSuccess(t *testing.T) {
	s := newStore(t)
	w := httptest.NewRecorder()
	validator := "true"
	u := sdktypes.NewUser().WithID(testUID)
	expectedSessionData := u
	err := s.Set(w, expectedSessionData)
	assert.Nil(t, err)

	securedCookie := w.Result().Cookies()[0]

	r := newRequestWithLoggedInCookie(validator)
	r.AddCookie(securedCookie)

	sd, err := s.Get(&r)
	assert.Nil(t, err)
	assert.Equal(t, sd.ID(), expectedSessionData.ID())
}
