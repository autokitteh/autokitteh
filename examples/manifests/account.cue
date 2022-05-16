package default

import "autokitteh.io/manifest"

manifest.#Manifest & {
	accounts: [
		{name: "internal"},
		{name: "autokitteh"},
	]

	eventsrcs: [ for _id in [
		"autokitteh.github",
		"autokitteh.slack",
		"autokitteh.twilio",
		"internal.cron",
		"internal.fs",
		"internal.http",
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
