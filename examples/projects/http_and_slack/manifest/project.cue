package http_and_slack

import "encoding/json"

import "autokitteh.io/manifest"

manifest.#Manifest & {
	projects: [
		{
			id:           "autokitteh.http_and_slack"
			name:         "http_and_slack"
			account_name: "autokitteh"
			main_path:    "fs:examples/projects/http_and_slack/auto.kitteh.cue"
			plugins: {
				"autokitteh.slack": {}
			}
			src_bindings: {
				"http": {
					src_id:     "internal.http"
					assoc:      "\(id).http"
					src_config: json.Marshal({
						name: "http_and_slack"
						routes: [
							{
								name: "catchall"
								path: "*"
							},
						]
					})
				}
				"slack_softkitteh": {
					src_id: "autokitteh.slack"
					assoc:  "TFPTT3QFN"
				}
			}
		},
	]
}
