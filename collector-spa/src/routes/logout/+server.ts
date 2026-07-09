import { buildApiHeaders, buildInternalApiPath } from '$lib/server/api';
import { clearAuthCookies, REFRESH_TOKEN_COOKIE } from '$lib/server/auth';
import { redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ cookies, fetch }) => {
	const refreshToken = cookies.get(REFRESH_TOKEN_COOKIE);

	if (refreshToken) {
		try {
			await fetch(buildInternalApiPath('/auth/logout'), {
				method: 'POST',
				headers: buildApiHeaders({ contentType: 'application/json' }),
				body: JSON.stringify({ refresh_token: refreshToken })
			});
		} catch {
			// Best-effort: logging the user out locally must succeed even if
			// the backend is unreachable.
		}
	}

	clearAuthCookies(cookies);
	redirect(303, '/');
};
