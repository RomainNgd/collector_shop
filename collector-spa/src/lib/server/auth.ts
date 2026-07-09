import { redirect, type Cookies } from '@sveltejs/kit';
import { claimsToUser, verifyJwtPayload } from '$lib/server/jwt';
import { buildApiHeaders, buildInternalApiPath, readApiResponse } from '$lib/server/api';
import type { AuthUser } from '$lib/types';

export const AUTH_TOKEN_COOKIE = 'auth_token';
export const REFRESH_TOKEN_COOKIE = 'refresh_token';
export const REFRESH_TOKEN_PATH = '/api/auth';

type RefreshApiResponse = {
	success?: boolean;
	data?: {
		token?: string;
		refresh_token?: string;
	};
};

export const setAuthCookies = (cookies: Cookies, url: URL, token: string, refreshToken: string) => {
	const secure = url.protocol === 'https:';

	cookies.set(AUTH_TOKEN_COOKIE, token, {
		path: '/',
		httpOnly: true,
		sameSite: 'lax',
		secure,
		maxAge: 60 * 60 * 24
	});

	cookies.set(REFRESH_TOKEN_COOKIE, refreshToken, {
		path: REFRESH_TOKEN_PATH,
		httpOnly: true,
		sameSite: 'lax',
		secure,
		maxAge: 60 * 60 * 24 * 30
	});
};

export const clearAuthCookies = (cookies: Cookies) => {
	cookies.delete(AUTH_TOKEN_COOKIE, { path: '/' });
	cookies.delete(REFRESH_TOKEN_COOKIE, { path: REFRESH_TOKEN_PATH });
};

/**
 * Attempts a transparent refresh using the refresh_token cookie. On success,
 * re-sets both cookies and returns the derived user. On failure, clears both
 * cookies and returns null.
 */
export const attemptRefresh = async (
	cookies: Cookies,
	url: URL,
	fetchFn: typeof fetch,
	jwtSecret: string
): Promise<AuthUser | null> => {
	const refreshToken = cookies.get(REFRESH_TOKEN_COOKIE);
	if (!refreshToken) {
		return null;
	}

	try {
		const response = await fetchFn(buildInternalApiPath('/auth/refresh'), {
			method: 'POST',
			headers: buildApiHeaders({ contentType: 'application/json' }),
			body: JSON.stringify({ refresh_token: refreshToken })
		});

		if (!response.ok) {
			clearAuthCookies(cookies);
			return null;
		}

		const result = await readApiResponse<RefreshApiResponse['data']>(response);
		const newToken = result.payload?.data?.token;
		const newRefreshToken = result.payload?.data?.refresh_token;

		if (!newToken || !newRefreshToken) {
			clearAuthCookies(cookies);
			return null;
		}

		const claims = verifyJwtPayload(newToken, jwtSecret);
		const user = claims ? claimsToUser(claims) : null;
		if (!user) {
			clearAuthCookies(cookies);
			return null;
		}

		setAuthCookies(cookies, url, newToken, newRefreshToken);
		return user;
	} catch {
		clearAuthCookies(cookies);
		return null;
	}
};

/**
 * Guards a load function for a route that requires an active session. Use
 * this (rather than a bare `if (!locals.user) redirect(...)`) when the
 * caller reached this point after having lost a previously valid session
 * (e.g. a proxied API call rejected the access token), so the login page can
 * show a "session expired" message instead of a plain login prompt.
 */
export const requireSession = (user: AuthUser | null): AuthUser => {
	if (!user) {
		redirect(303, '/login?reason=session_expired');
	}
	return user;
};

/**
 * Redirects to the login page with a session-expired reason when a proxied
 * API call comes back unauthorized. Callers should invoke this for 401
 * responses specifically; other error statuses should keep surfacing as
 * regular SvelteKit errors.
 */
export const redirectToLoginOnAuthFailure = (response: Response): void => {
	if (response.status === 401) {
		redirect(303, '/login?reason=session_expired');
	}
};
