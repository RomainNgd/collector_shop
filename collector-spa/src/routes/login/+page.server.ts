import {
	buildApiHeaders,
	buildInternalApiPath,
	getApiErrorMessage,
	readApiResponse
} from '$lib/server/api';
import { setAuthCookies } from '$lib/server/auth';
import { claimsToUser, decodeJwtPayload } from '$lib/server/jwt';
import { getFormString } from '$lib/server/forms';
import { ADMIN_ROLE } from '$lib/types';
import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';

type LoginApiResponse = {
	success?: boolean;
	data?: {
		token?: string;
		refresh_token?: string;
	};
	error?: {
		message?: string;
	};
};

export const load: PageServerLoad = ({ locals, cookies, url }) => {
	if (locals.user) {
		redirect(303, '/');
	}

	const registeredEmail = String(cookies.get('register_success_email') ?? '').trim();
	if (registeredEmail) {
		cookies.delete('register_success_email', { path: '/' });
	}

	return {
		registered: registeredEmail.length > 0,
		registeredEmail,
		sessionExpired: url.searchParams.get('reason') === 'session_expired'
	};
};

export const actions: Actions = {
	default: async ({ request, fetch, cookies, url }) => {
		const formData = await request.formData();
		const email = getFormString(formData, 'email').trim();
		const password = getFormString(formData, 'password');

		if (!email || !password) {
			return fail(400, {
				error: 'Email et mot de passe requis',
				email
			});
		}

		const response = await fetch(buildInternalApiPath('/auth/login'), {
			method: 'POST',
			headers: buildApiHeaders({ contentType: 'application/json' }),
			body: JSON.stringify({ email, password })
		});

		const result = await readApiResponse<LoginApiResponse['data']>(response);
		const token = result.payload?.data?.token;
		const refreshToken = result.payload?.data?.refresh_token;

		if (!response.ok || result.payload?.success !== true || !token || !refreshToken) {
			return fail(401, {
				error:
					getApiErrorMessage(response, result, 'Identifiants invalides') ??
					'Identifiants invalides',
				email
			});
		}

		const claims = decodeJwtPayload(token);
		const user = claims ? claimsToUser(claims) : null;
		if (!user) {
			return fail(401, {
				error: 'Token invalide ou expire',
				email
			});
		}

		setAuthCookies(cookies, url, token, refreshToken);

		if (user.role === ADMIN_ROLE) {
			redirect(303, '/administration');
		}

		redirect(303, '/');
	}
};
