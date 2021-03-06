import "autokitteh.io/program"

program.#Program & {
	modules: [
		{
			path: "fs:examples/default/test.kitteh"
			name: "spec"
			sources: {
				"slack": "slack_softkitteh"
			}
			context: {
				team_id: "TFPTT3QFN"
			}
		},
	]

	consts: {
		"sounds": {
			"cat": "meow"
			"dog": "woof"
			"pig": "oink"
			"cow": "moo"
		}
	}
}
