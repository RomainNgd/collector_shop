import { claimsToUser, decodeJwtPayload } from '$lib/server/jwt';
import type { Handle } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
	const token = event.cookies.get('auth_token');

	if (!token) {
		event.locals.user = null;
		return resolve(event);
	}

	const claims = decodeJwtPayload(token);
	const user = claims ? claimsToUser(claims) : null;

	if (!user) {
		event.cookies.delete('auth_token', { path: '/' });
		event.locals.user = null;
		return resolve(event);
	}

	event.locals.user = user;
	return resolve(event);
};
