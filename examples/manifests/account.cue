package default

import "autokitteh.io/manifest"

manifest.#Manifest & {
	accounts: {
    internal: {}
    autokitteh: {}
  }

	eventsrcs: {
		"autokitteh.github": {}
		"autokitteh.slack": {}
		"autokitteh.twilio": {}
		"internal.cron": {}
		"internal.fs": {}
		"internal.http": {}
  }

	plugins: {
	  "autokitteh.slack": exec: name: "slack"
		"autokitteh.github": exec: name: "github"
	  "autokitteh.googlesheets": exec: name: "googlesheets"
		"autokitteh.twilio": exec: name: "twilio"
		"autokitteh.test": exec: name: "test"
	}
}
