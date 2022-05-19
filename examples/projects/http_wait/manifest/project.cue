package http_wait

import "encoding/json"

import "autokitteh.io/manifest"

manifest.#Manifest & {
	projects: {
		"autokitteh.http_wait": {
			account_name: "autokitteh"
			main_path:    "fs:examples/projects/http_wait/auto.kitteh"
			src_bindings: {
				"http": {
					src_id:     "internal.http"
					assoc:      "http_wait.http"
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
		}
  }
}
