package http_wait

import "encoding/json"

import "autokitteh.io/manifest"

manifest.#Manifest & {
	projects: [
		{
			id:           "http_wait"
			name:         "http_wait"
			account_name: "autokitteh"
			main_path:    "fs:examples/projects/http_wait/auto.kitteh"
			src_bindings: {
				"http": {
					src_id:     "internal.http"
					assoc:      "\(id).http"
					src_config: json.Marshal({
						name: "http_wait"
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
