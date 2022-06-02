package default

import "encoding/json"

import "autokitteh.io/manifest"

_Project: manifest.#Project & {
	name:         string
	id:           string
	account_name: "autokitteh"
	plugins: {
		"autokitteh.github": {}
		"autokitteh.googlesheets": {}
		"autokitteh.slack": {}
		"autokitteh.test": {}
		"autokitteh.twilio": {}
		"internal.http": {}
		"internal.os": {}
		"internal.time": {}
	}
	src_bindings: {
		"twilio": {
			src_id: "autokitteh.twilio"
			assoc:  "ACcffc0e031a5d17f6ad03ce06b0be25c4"
		}
		"fs_autokitteh": {
			src_id:     "autokitteh.fs"
			assoc:      id
			src_config: json.Marshal({
				path:     "examples/default"
				ops_mask: 31
			})
		}
		"http_autokitteh": {
			src_id:     "autokitteh.http"
			assoc:      "\(id).autokitteh"
			src_config: json.Marshal({
				name: "http_autokitteh"
				routes: [
					{
						name: "autokitteh"
						path: "*"
					},
				]
			})
		}
		"github_softkitteh": {
			src_id: "autokitteh.github"
			assoc:  "softkitteh"
		}
		"slack_softkitteh": {
			src_id: "autokitteh.slack"
			assoc:  "TFPTT3QFN"
		}
		"slack_manifest": {
			src_id: "autokitteh.slack"
			assoc:  "T02T1NWQK62"
		}
		"cron_every_minute": {
			src_id:     "autokitteh.cron"
			src_config: "@every 1m"
		}
	}
}
