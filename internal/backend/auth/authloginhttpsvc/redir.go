package authloginhttpsvc

import (
	"net/http"
	"net/url"
)

const (
	redirCookieName = "redir"
	cookieMaxAge    = 5 * 60 // Short-lived cookie for OAuth flow
)

func RedirectToLogin(w http.ResponseWriter, r *http.Request, src *url.URL, domain string, secure bool, sameSite http.SameSite) {
	if src != nil {
		http.SetCookie(w, &http.Cookie{
			Name:     redirCookieName,
			Value:    src.String(),
			Path:     "/",
			Domain:   domain,
			Secure:   secure,
			SameSite: sameSite,
			HttpOnly: false,
			MaxAge:   cookieMaxAge,
		})
	}

	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}

func getRedirect(r *http.Request) string {
	if c, err := r.Cookie(redirCookieName); err == nil {
		return c.Value
	}

	return "/"
}
