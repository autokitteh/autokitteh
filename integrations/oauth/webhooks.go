package oauth

import (
	"net/http"
)

func HTTPError(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}

// startOAuthFlow starts a 3-legged OAuth 2.0 flow by redirecting the user (via a web
// page) to the authorization endpoint of a third-party service named in the request.
func (o *OAuth) startOAuthFlow(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement this function.
}

// exchangeCodeToToken receives a redirect back from a third-party service's
// authorization endpoint (the OAuth 2.0 2nd leg), and exchanges the received
// authorization code for an new access token (the 3rd leg). If all goes well,
// it redirects the token back to the named integration's own OAuth webhook in
// order to complete the initialization procedure of an AutoKitteh connection.
func (o *OAuth) exchangeCodeToToken(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement this function.
}
