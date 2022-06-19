import { api } from './_api';

/** @type {import('./__types').RequestHandler} */
export const get = async (
	// { locals }
	) => {
	// locals.userid comes from src/hooks.js
	// const response = await api('get', `accounts/${locals.userid}`);
	try {
		const response = await (await api('get', `accounts/autokitteh`)).json();
		console.log({response})
		return {
			status: 200,
			body: response
		}
	} catch (err) {
		return {
			status: 500,
			body: {errors: err}
		}
	}
};

// /** @type {import('./index').RequestHandler} */
// export const post = async ({ request, locals }) => {
// 	const form = await request.formData();

// 	await api('post', `todos/${locals.userid}`, {
// 		text: form.get('text')
// 	});

// 	return {};
// };

// // If the user has JavaScript disabled, the URL will change to
// // include the method override unless we redirect back to /todos
// const redirect = {
// 	status: 303,
// 	headers: {
// 		location: '/todos'
// 	}
// };

// /** @type {import('./index').RequestHandler} */
// export const patch = async ({ request, locals }) => {
// 	const form = await request.formData();

// 	await api('patch', `todos/${locals.userid}/${form.get('uid')}`, {
// 		text: form.has('text') ? form.get('text') : undefined,
// 		done: form.has('done') ? !!form.get('done') : undefined
// 	});

// 	return redirect;
// };

// /** @type {import('./index').RequestHandler} */
// export const del = async ({ request, locals }) => {
// 	const form = await request.formData();

// 	await api('delete', `todos/${locals.userid}/${form.get('uid')}`);

// 	return redirect;
// };
