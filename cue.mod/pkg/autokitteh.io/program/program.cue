package program

import "autokitteh.io/values"

#Module: {
	path: string // load path
	name: string // name of value containing the spec
	sources: [string]: string // global -> specific
	context: [string]: values.#Value
}

#Program: {
	modules?: [...#Module]
	consts?: values.#Values
}
