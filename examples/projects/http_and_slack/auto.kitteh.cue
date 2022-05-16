package http_and_slack

import "autokitteh.io/program"

let #slack_team_id = "TFPTT3QFN"

program.#Program & {
	modules: [
		{
			path: "fs:examples/projects/http_and_slack/slack.kitteh"
			name: "spec"
			sources: {
				"slack": "slack_softkitteh"
			}
			context: {
				team_id: #slack_team_id
			}
		},
		{
			path: "fs:examples/projects/http_and_slack/http.kitteh"
			name: "spec"
			sources: {
				"http": "http"
			}
			context: {
				slack_team_id: #slack_team_id
				slack_channel: "CFPAVMMU0" // #general
			}
		},
	]
}
