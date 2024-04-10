define(() => {
	return {
		getById: (id) => document.getElementById(id),

		create: (element) => document.createElement(element),

		text: (text) => document.createTextNode(text)
	}
})