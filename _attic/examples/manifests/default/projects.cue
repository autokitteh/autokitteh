package default

import "autokitteh.io/manifest"

P0: manifest.#Manifest & {
	projects: [_Project & {
		id:        "P0"
		name:      "auto.kitteh"
		main_path: "fs:examples/default/auto.kitteh"
	}]
}

P1: manifest.#Manifest & {
	projects: [_Project & {
		id:        "P1"
		name:      "auto.kitteh.cue"
		main_path: "fs:examples/default/auto.kitteh.cue"
	}]
}

P2: manifest.#Manifest & {
	projects: [_Project & {
		id:        "P2"
		name:      "test.kitteh"
		main_path: "fs:examples/default/test.kitteh.yaml"
	}]
}
