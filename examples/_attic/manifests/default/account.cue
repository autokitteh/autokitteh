package default

import "autokitteh.io/manifest"

manifest.#Manifest & {
	accounts: [{name: "autokitteh"}]

	eventsrcs: [ for _id in [
		"autokitteh.twilio",
		"autokitteh.slack",
		"autokitteh.fs",
		"internal.http",
		"autokitteh.cron",
		"autokitteh.github",
	] {id: _id}]

	plugins: [
		{
			id: "autokitteh.slack"
			exec: name: "slack"
		},
		{
			id: "autokitteh.github"
			exec: name: "github"
		},
		{
			id: "autokitteh.googlesheets"
			exec: name: "googlesheets"
		},
		{
			id: "autokitteh.twilio"
			exec: name: "twilio"
		},
		{
			id: "autokitteh.test"
			exec: name: "test"
		},
	]
}
