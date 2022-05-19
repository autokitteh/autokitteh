package http_wait_value

import "encoding/json"

import "autokitteh.io/manifest"

manifest.#Manifest & {
	projects: {
		"autokitteh.http_wait_value": {
			main_path: "fs:examples/projects/http_wait_value/auto.kitteh"
			src_bindings: {
				"http": {
					src_id:     "internal.http"
					assoc:      "autokitteh.http_wait_value.http"
					src_config: json.Marshal({
						name: "http_wait_value"
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
