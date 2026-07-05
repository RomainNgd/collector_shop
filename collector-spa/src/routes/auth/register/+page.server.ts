import { buildApiHeaders, buildInternalApiPath, readApiResponse } from '$lib/server/api';
import { getFormString } from '$lib/server/forms';
import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';

type RegisterApiResponse = {
	success?: boolean;
	data?: {
		id?: number;
		email?: string;
		role?: string;
	};
	error?: {
		message?: string;
	};
};

const getRegistrationError = (status: number): string => {
	if (status === 409) {
		return 'Cette adresse email est deja utilisee';
	}

	if (status === 400) {
		return 'Verifie ton email et utilise un mot de passe de 8 caracteres minimum';
	}

	return 'Impossible de creer le compte pour le moment';
};

export const load: PageServerLoad = ({ locals }) => {
	if (locals.user) {
		redirect(303, '/');
	}
};

export const actions: Actions = {
	default: async ({ request, fetch, cookies, url }) => {
		const formData = await request.formData();
		const email = getFormString(formData, 'email').trim();
		const password = getFormString(formData, 'password');
		const confirmPassword = getFormString(formData, 'confirmPassword');

		if (!email || !password || !confirmPassword) {
			return fail(400, {
				error: 'Email, mot de passe et confirmation requis',
				email
			});
		}

		if (password.length < 8) {
			return fail(400, {
				error: 'Le mot de passe doit contenir au moins 8 caracteres',
				email
			});
		}

		if (password !== confirmPassword) {
			return fail(400, {
				error: 'Les mots de passe ne correspondent pas',
				email
			});
		}

		const response = await fetch(buildInternalApiPath('/auth/register'), {
			method: 'POST',
			headers: buildApiHeaders({ contentType: 'application/json' }),
			body: JSON.stringify({ email, password })
		});

		const result = await readApiResponse<RegisterApiResponse['data']>(response);

		if (!response.ok || result.payload?.success !== true) {
			return fail(response.status, {
				error: getRegistrationError(response.status),
				email
			});
		}

		cookies.set('register_success_email', email, {
			path: '/',
			httpOnly: true,
			sameSite: 'lax',
			secure: url.protocol === 'https:',
			maxAge: 60
		});

		return {
			success: true,
			email,
			message: 'Compte cree avec succes. Redirection vers la connexion...'
		};
	}
};
