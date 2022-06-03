package default

import "autokitteh.io/manifest"

manifest.#Manifest & {
	accounts: {
		"autokitteh": {}
	}

	eventsrcs: {
		"autokitteh.github": {}
		"autokitteh.slack": {}
		"autokitteh.twilio": {}
		"internal.fs": {}
	}

	plugins: {
		"autokitteh.slack": exec: name:        "slack"
		"autokitteh.github": exec: name:       "github"
		"autokitteh.googlesheets": exec: name: "googlesheets"
		"autokitteh.twilio": exec: name:       "twilio"
		"autokitteh.test": exec: name:         "test"
	}
}
