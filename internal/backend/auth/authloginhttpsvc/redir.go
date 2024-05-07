package authloginhttpsvc

import (
	"net/http"
	"net/url"
)

const redirCookieName = "redir"

func RedirectToLogin(w http.ResponseWriter, r *http.Request, src *url.URL) {
	if src != nil {
		http.SetCookie(w, &http.Cookie{
			Name:  redirCookieName,
			Value: src.String(),
			Path:  "/",
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
