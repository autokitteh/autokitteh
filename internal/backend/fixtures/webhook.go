package fixtures

import "os"

var serviceAddress string

func initServiceAddress() {
	// This needs to be done lazyly to let the main load dotenv.

	if serviceAddress == "" {
		serviceAddress = os.Getenv("SERVICE_ADDRESS")
	}

	if serviceAddress == "" {
		serviceAddress = os.Getenv("WEBHOOK_ADDRESS")
	}
}

func ServiceAddress() string {
	initServiceAddress()

	return serviceAddress
}

func ServiceBaseURL() string {
	initServiceAddress()

	if serviceAddress != "" {
		return "https://" + serviceAddress
	}

	return ""
}
