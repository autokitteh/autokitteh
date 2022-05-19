package http_wait_event

import "encoding/json"

import "autokitteh.io/manifest"

manifest.#Manifest & {
	projects: {
		"autokitteh.http_wait": {
			main_path: "fs:examples/projects/http_wait_event/auto.kitteh"
			src_bindings: {
				"http": {
					src_id:     "internal.http"
					assoc:      "autokitteh.http_wait.http"
					src_config: json.Marshal({
						name: "http_wait_event"
						routes: [
							{
								name: "catchall"
								path: "*"
							},
						]
					})
				}
			}
		}
	}
}
