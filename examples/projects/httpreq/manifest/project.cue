package http_wait_value

import "encoding/json"

import "autokitteh.io/manifest"

manifest.#Manifest & {
	projects: {
		"autokitteh.httpreq": {
			main_path: "fs:examples/projects/httpreq/auto.kitteh"
      plugins: {
        "internal.http": {}
      }
			src_bindings: {
				"http": {
					src_id:     "internal.http"
					assoc:      "autokitteh.httpreq.http",
					src_config: json.Marshal({
						name: "httpreq"
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

