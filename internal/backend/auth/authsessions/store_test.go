package authsessions

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dghubble/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// A type used as a mocked underlying sessions store
// With this we can control in each test what each dependent function would return
// So we can test for edge cases easily
type baseMockSessionStore struct {
	sessions.Store[[]byte]
	nnew    func(name string) *sessions.Session[[]byte]
	get     func(*http.Request, string) (*sessions.Session[[]byte], error)
	save    func(http.ResponseWriter, *sessions.Session[[]byte]) error
	destroy func(http.ResponseWriter, string)
}

func (b baseMockSessionStore) New(name string) *sessions.Session[[]byte] {
	return b.nnew(name)
}

func (b baseMockSessionStore) Get(req *http.Request, name string) (*sessions.Session[[]byte], error) {
	return b.get(req, name)
}

// Save writes a Session to the ResponseWriter
func (b baseMockSessionStore) Save(w http.ResponseWriter, session *sessions.Session[[]byte]) error {
	return b.save(w, session)
}

// Destroy removes (expires) a named Session
func (b baseMockSessionStore) Destroy(w http.ResponseWriter, name string) {
	b.destroy(w, name)
}

func TestStoreGetNoLoggedinCookieReturnError(t *testing.T) {
	s := store{store: baseMockSessionStore{}}
	r := http.Request{}
	sd, err := s.Get(&r)
	assert.Nil(t, sd)
	assert.ErrorIs(t, err, http.ErrNoCookie)
}

func TestStoreGetInvalidLoggedInCookie(t *testing.T) {
	s := store{store: baseMockSessionStore{}}
	r := http.Request{
		Header: http.Header{
			"Cookie": []string{"ak_logged_in=kjs"},
		},
	}
	sd, err := s.Get(&r)
	assert.Nil(t, sd)
	assert.Equal(t, err.Error(), "invalid logged in cookie")
}

func TestStoreGetNoSessionCookie(t *testing.T) {
	baseStore := baseMockSessionStore{
		get: func(r *http.Request, s string) (*sessions.Session[[]byte], error) {
			return nil, http.ErrNoCookie
		},
	}

	s := store{store: baseStore}
	r := http.Request{
		Header: http.Header{
			"Cookie": []string{"ak_logged_in=8383"},
		},
	}
	sd, err := s.Get(&r)
	assert.Nil(t, sd)
	assert.Nil(t, err)
}

func TestStoreGetInvalidSessionCookie(t *testing.T) {
	baseStore := baseMockSessionStore{
		get: func(r *http.Request, s string) (*sessions.Session[[]byte], error) {
			return nil, errors.New("something")
		},
	}

	s := store{store: baseStore}
	r := http.Request{
		Header: http.Header{
			"Cookie": []string{"ak_logged_in=8383"},
		},
	}
	sd, err := s.Get(&r)
	assert.Nil(t, sd)
	assert.Error(t, err)
}

func TestStoreGetInvalidSessionCookieStructure(t *testing.T) {
	baseStore := baseMockSessionStore{
		get: func(r *http.Request, s string) (*sessions.Session[[]byte], error) {
			return &sessions.Session[[]byte]{}, nil
		},
	}

	s := store{store: baseStore}
	r := http.Request{
		Header: http.Header{
			"Cookie": []string{"ak_logged_in=8383"},
		},
	}
	sd, err := s.Get(&r)
	assert.Nil(t, sd)
	assert.Error(t, err)
}

func TestStoreGetInvalidValidator(t *testing.T) {
	baseStore := baseMockSessionStore{}

	var mockSession *sessions.Session[[]byte]

	baseStore.nnew = func(name string) *sessions.Session[[]byte] {
		mockSession = sessions.NewSession(baseStore, sessionName)
		return mockSession
	}
	s := store{store: &baseStore}
	validator := int64(111)
	invalidValidator := validator + 1
	session, err := s.newSessionWithData(&sessionData{Validator: validator, User: sdktypes.DefaultUser})
	require.Nil(t, err)

	baseStore.get = func(r *http.Request, s string) (*sessions.Session[[]byte], error) {
		return session, nil
	}

	r := http.Request{
		Header: http.Header{
			"Cookie": []string{fmt.Sprintf("ak_logged_in=%d", invalidValidator)},
		},
	}
	sd, err := s.Get(&r)
	assert.Nil(t, sd)
	assert.Error(t, err)
}

func TestStoreGetSuccess(t *testing.T) {
	baseStore := baseMockSessionStore{}

	var mockSession *sessions.Session[[]byte]

	baseStore.nnew = func(name string) *sessions.Session[[]byte] {
		mockSession = sessions.NewSession(baseStore, sessionName)
		return mockSession
	}
	s := store{store: &baseStore}
	validator := int64(111)
	mockedSessionData := &sessionData{Validator: validator, User: sdktypes.DefaultUser}
	session, err := s.newSessionWithData(mockedSessionData)
	require.Nil(t, err)

	baseStore.get = func(r *http.Request, s string) (*sessions.Session[[]byte], error) {
		return session, nil
	}

	r := http.Request{
		Header: http.Header{
			"Cookie": []string{fmt.Sprintf("ak_logged_in=%d", validator)},
		},
	}
	sd, err := s.Get(&r)
	assert.Nil(t, err)
	assert.Equal(t, sd, mockedSessionData)
}

// Set Tests
func TestStoreSetSessionSuccess(t *testing.T) {
	// Prepare
	w := httptest.NewRecorder()
	didSaveHTTPOnlyCookie := false
	baseStore := baseMockSessionStore{
		save: func(w http.ResponseWriter, s *sessions.Session[[]byte]) error {
			// If this is called, it means the session save method was called
			// meaning the cookie is saved to the http request
			didSaveHTTPOnlyCookie = true
			return nil
		},
	}

	var mockSession *sessions.Session[[]byte]

	baseStore.nnew = func(name string) *sessions.Session[[]byte] {
		mockSession = sessions.NewSession(baseStore, sessionName)
		return mockSession
	}

	s := store{store: baseStore}

	validator := time.Now().Unix()
	// Test
	err := s.Set(w, &sessionData{Validator: validator, User: sdktypes.DefaultUser})

	// Assert
	assert.Nil(t, err)
	responseCookies := w.Result().Cookies()

	// Verify is logged in cookie
	assert.Equal(t, len(responseCookies), 1) // http only cookie ie not fully set since we mock the session store
	cookie := responseCookies[0]
	assert.Equal(t, cookie.Name, loggedInCookie)
	assert.Equal(t, cookie.Value, fmt.Sprintf("%d", validator))
	assert.Equal(t, cookie.Path, "/")

	// verify session cookie
	assert.True(t, didSaveHTTPOnlyCookie)
	assert.Equal(t, mockSession.Name(), sessionName)
}

func TestStoreSetSessionSaveFailed(t *testing.T) {
	// Prepare
	w := httptest.NewRecorder()
	baseStore := baseMockSessionStore{
		save: func(w http.ResponseWriter, s *sessions.Session[[]byte]) error {
			return errors.New("something")
		},
	}

	var mockSession *sessions.Session[[]byte]

	baseStore.nnew = func(name string) *sessions.Session[[]byte] {
		mockSession = sessions.NewSession(baseStore, sessionName)
		return mockSession
	}

	s := store{store: baseStore}

	validator := time.Now().Unix()

	// Test
	err := s.Set(w, &sessionData{Validator: validator, User: sdktypes.DefaultUser})

	// Assert
	assert.Error(t, err)
	responseCookies := w.Result().Cookies()

	// Verify is logged in cookie
	assert.Equal(t, len(responseCookies), 0) // http only cookie ie not fully set since we mock the session store
}
