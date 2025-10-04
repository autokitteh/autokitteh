package fixtures

import "os"

var webhookAddress = os.Getenv("SERVICE_ADDRESS")

func init() {
	if webhookAddress == "" {
		// fallback to legacy var name.
		webhookAddress = os.Getenv("WEBHOOK_ADDRESS")
	}
}

func ServiceAddress() string { return webhookAddress }
func ServiceBaseURL() string { return "https://" + webhookAddress }
