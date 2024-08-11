package authsessions

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func newStore(t *testing.T) Store {
	cfg := Configs.Dev
	s, err := New(cfg)
	assert.NoError(t, err)
	return s
}

func newRequestWithLoggedInCookie(value int64) http.Request {
	return http.Request{
		Header: http.Header{
			"Cookie": []string{fmt.Sprintf("ak_logged_in=%d", value)},
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
	assert.Equal(t, err.Error(), "invalid logged in cookie")
}

func TestStoreGetNoSessionCookie(t *testing.T) {
	s := newStore(t)
	r := newRequestWithLoggedInCookie(8383)
	sd, err := s.Get(&r)
	assert.Nil(t, sd)
	assert.Nil(t, err)
}

func TestStoreGetInvalidSessionCookie(t *testing.T) {
	s := newStore(t)
	r := newRequestWithLoggedInCookie(8383)

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
	validator := int64(123)
	err := s.Set(w, &sessionData{
		User:      sdktypes.DefaultUser,
		Validator: validator,
	})
	assert.Nil(t, err)

	otherValidator := validator + 1
	r := newRequestWithLoggedInCookie(otherValidator)
	r.AddCookie(w.Result().Cookies()[0])

	sd, err := s.Get(&r)
	assert.Nil(t, sd)
	assert.Error(t, err)
}

func TestStoreSetGetSuccess(t *testing.T) {
	s := newStore(t)
	w := httptest.NewRecorder()
	validator := int64(123)
	expectedSessionData := &sessionData{
		User:      sdktypes.DefaultUser,
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
	assert.Equal(t, sd.User, expectedSessionData.User)
}
