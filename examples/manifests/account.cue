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
		"autokitteh.slack": exec: path:        "examples/plugins/slack"
		"autokitteh.github": exec: path:       "examples/plugins/github"
		"autokitteh.googlesheets": exec: path: "examples/plugins/googlesheets"
		"autokitteh.twilio": exec: path:       "examples/plugins/twilio"
		"autokitteh.test": exec: path:         "examples/plugins/test"
		"autokitteh.aws": {
			address: "127.0.0.1"
			port:    30001
		}
	}
}
