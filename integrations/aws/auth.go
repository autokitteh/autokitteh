package aws

type authData struct {
	Region      string
	AccessKeyID string `var:"secret"`
	SecretKey   string `var:"secret"`
	Token       string `var:"secret"`
}
