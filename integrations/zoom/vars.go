package zoom

// privateApp contains the user-provided details of a private
// Zoom OAuth 2.0 app or Service-to-Service internal app.
type privateApp struct {
	AccountID    string `var:"private_account_id"`
	ClientID     string `var:"private_client_id"`
	ClientSecret string `var:"private_client_secret,secret"`
	SecretToken  string `var:"private_secret_token,secret"`
}
