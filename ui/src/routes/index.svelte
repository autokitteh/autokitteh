<script context="module">
	export const prerender = true;
</script>

<script>
	import { onMount } from 'svelte';
	let foo = null;
	const fetchdata = () =>
		fetch('/api')
			.then((res) => res.json())
			.then((x) => (foo = x));
	onMount(fetchdata);
</script>

<section>
	<h1>Welcome</h1>

	<pre>
		{JSON.stringify(foo, null, 2)}
	</pre>
</section>

<style>
	section {
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		flex: 1;
	}

	h1 {
		width: 100%;
	}

	.welcome {
		display: block;
		position: relative;
		width: 100%;
		height: 0;
		padding: 0 0 calc(100% * 495 / 2048) 0;
	}

	.welcome img {
		position: absolute;
		width: 100%;
		height: 100%;
		top: 0;
		display: block;
	}
</style>
