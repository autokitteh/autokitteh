package http_sounds

import "encoding/json"

import "autokitteh.io/manifest"

manifest.#Manifest & {
	projects: {
		"autokitteh.http_sounds": {
			main_path:    "fs:examples/projects/http_sounds/auto.kitteh"
			src_bindings: {
				"http": {
					src_id:     "internal.http"
					assoc:      "http_sounds.http"
					src_config: json.Marshal({
						name: "http_sounds"
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
