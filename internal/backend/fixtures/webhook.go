package fixtures

import (
	"os"
	"sync"
)

var (
	serviceAddress string
	once           sync.Once
)

func initServiceAddress() {
	// This needs to be done lazily to let the main load dotenv.

	once.Do(func() {
		if serviceAddress == "" {
			serviceAddress = os.Getenv("SERVICE_ADDRESS")
		}

		if serviceAddress == "" {
			serviceAddress = os.Getenv("WEBHOOK_ADDRESS")
		}
	})
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
