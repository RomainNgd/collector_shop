import { env } from '$env/dynamic/private';
import { API_BASE_URL, API_INTERNAL_PREFIX } from '$lib/server/api';
import { claimsToUser, verifyJwtPayload } from '$lib/server/jwt';
import type { Handle, HandleFetch } from '@sveltejs/kit';

const joinApiUrl = (pathname: string, search: string) => {
	const apiUrl = new URL(API_BASE_URL);
	const upstreamPath =
		pathname === API_INTERNAL_PREFIX ? '/' : pathname.slice(API_INTERNAL_PREFIX.length);
	const basePath = apiUrl.pathname.endsWith('/') ? apiUrl.pathname.slice(0, -1) : apiUrl.pathname;

	apiUrl.pathname = `${basePath}${upstreamPath}`;
	apiUrl.search = search;

	return apiUrl;
};

export const handle: Handle = async ({ event, resolve }) => {
	const token = event.cookies.get('auth_token');

	if (!token) {
		event.locals.user = null;
		return resolve(event);
	}

	const claims = verifyJwtPayload(token, env.JWT_SECRET ?? '');
	const user = claims ? claimsToUser(claims) : null;

	if (!user) {
		event.cookies.delete('auth_token', { path: '/' });
		event.locals.user = null;
		return resolve(event);
	}

	event.locals.user = user;
	return resolve(event);
};

export const handleFetch: HandleFetch = async ({ event, request, fetch }) => {
	const requestUrl = new URL(request.url);
	const isInternalApiRequest =
		requestUrl.origin === event.url.origin &&
		(requestUrl.pathname === API_INTERNAL_PREFIX ||
			requestUrl.pathname.startsWith(`${API_INTERNAL_PREFIX}/`));

	if (!isInternalApiRequest) {
		return fetch(request);
	}

	const proxiedRequest = new Request(joinApiUrl(requestUrl.pathname, requestUrl.search), request);
	const token = event.cookies.get('auth_token');

	if (token && !proxiedRequest.headers.has('authorization')) {
		proxiedRequest.headers.set('authorization', `Bearer ${token}`);
	}

	return fetch(proxiedRequest);
};
