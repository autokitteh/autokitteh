package authsessions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var testUID = sdktypes.NewUserID()

func newStore(t *testing.T) Store {
	cfg := Configs.Dev
	s, err := New(cfg)
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
	assert.Nil(t, sd)
	assert.Nil(t, err)
}

func TestStoreGetInvalidLoggedInCookie(t *testing.T) {
	s := newStore(t)
	r := http.Request{
		Header: http.Header{
			"Cookie": []string{"ak_logged_in=k1k2"},
		},
	}
	sd, err := s.Get(&r)
	assert.Nil(t, sd)
	assert.Error(t, err, "invalid logged in cookie")
}

func TestStoreGetNoSessionCookie(t *testing.T) {
	s := newStore(t)
	r := newRequestWithLoggedInCookie(uuid.NewString())
	sd, err := s.Get(&r)
	assert.Nil(t, sd)
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
	assert.Nil(t, sd)
	assert.Error(t, err)
}

func TestStoreGetValidatorMismatch(t *testing.T) {
	s := newStore(t)
	w := httptest.NewRecorder()
	validator := uuid.NewString()
	err := s.Set(w, &sessionData{
		UserID:    testUID,
		Validator: validator,
	})
	assert.Nil(t, err)

	otherValidator := uuid.NewString()
	r := newRequestWithLoggedInCookie(otherValidator)
	r.AddCookie(w.Result().Cookies()[0])

	sd, err := s.Get(&r)
	assert.Nil(t, sd)
	assert.Error(t, err)
}

func TestStoreSetGetSuccess(t *testing.T) {
	s := newStore(t)
	w := httptest.NewRecorder()
	validator := uuid.NewString()
	expectedSessionData := &sessionData{
		UserID:    testUID,
		Validator: validator,
	}
	err := s.Set(w, expectedSessionData)
	assert.Nil(t, err)

	securedCookie := w.Result().Cookies()[0]

	r := newRequestWithLoggedInCookie(validator)
	r.AddCookie(securedCookie)

	sd, err := s.Get(&r)
	assert.Nil(t, err)
	assert.Equal(t, sd.Validator, expectedSessionData.Validator)
	assert.Equal(t, sd.UserID, expectedSessionData.UserID)
}
