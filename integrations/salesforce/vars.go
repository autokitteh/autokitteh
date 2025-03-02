package salesforce

// privateOAuth contains the user-provided details of a private Salesforce OAuth 2.0 app.
type privateOAuth struct {
	ClientID     string `var:"private_client_id"`
	ClientSecret string `var:"private_client_secret,secret"`
}
