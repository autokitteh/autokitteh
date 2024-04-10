define(() => {
	const styles = () => getComputedStyle(document.documentElement)

	return {
		getVar: (v) => styles().getPropertyValue(v),

		setVar: (k, v) => document.documentElement.style.setProperty(k, v)
	}
})