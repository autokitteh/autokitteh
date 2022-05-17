package http_sounds

import "encoding/json"

import "autokitteh.io/manifest"

manifest.#Manifest & {
	projects: [
		{
			id:           "autokitteh.http_sounds"
			name:         "http_sounds"
			account_name: "autokitteh"
			main_path:    "fs:examples/projects/http_sounds/auto.kitteh"
			src_bindings: {
				"http": {
					src_id:     "internal.http"
					assoc:      "\(id).http"
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
		},
	]
}
