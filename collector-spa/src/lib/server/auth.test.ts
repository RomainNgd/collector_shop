import { createHmac } from 'node:crypto';
import { describe, expect, it, vi } from 'vitest';

import { USER_ROLE } from '$lib/types';
import { attemptRefresh, redirectToLoginOnAuthFailure, requireSession } from './auth';

const secret = 'auth-test-secret';

const signToken = (claims: Record<string, unknown>) => {
	const header = Buffer.from(JSON.stringify({ alg: 'HS256', typ: 'JWT' })).toString('base64url');
	const payload = Buffer.from(JSON.stringify(claims)).toString('base64url');
	const signature = createHmac('sha256', secret).update(`${header}.${payload}`).digest('base64url');
	return `${header}.${payload}.${signature}`;
};

const cookieJar = (values: Record<string, string | undefined> = {}) => ({
	get: vi.fn((name: string) => values[name]),
	set: vi.fn(),
	delete: vi.fn()
});

describe('attemptRefresh', () => {
	const url = new URL('https://collector.test/');

	it('returns null without calling the API when there is no refresh cookie', async () => {
		const cookies = cookieJar();
		const fetchMock = vi.fn();

		const user = await attemptRefresh(cookies as never, url, fetchMock, secret);

		expect(user).toBeNull();
		expect(fetchMock).not.toHaveBeenCalled();
	});

	it('re-sets both cookies and returns the user on success', async () => {
		const cookies = cookieJar({ refresh_token: 'old-refresh-token' });
		const newToken = signToken({
			sub: 3,
			role: USER_ROLE,
			exp: Math.floor(Date.now() / 1000) + 60
		});
		const fetchMock = vi.fn(() =>
			Promise.resolve(
				new Response(
					JSON.stringify({ success: true, data: { token: newToken, refresh_token: 'new-token' } })
				)
			)
		);

		const user = await attemptRefresh(cookies as never, url, fetchMock, secret);

		expect(user).toEqual({ id: 3, role: USER_ROLE, email: undefined });
		expect(cookies.set).toHaveBeenCalledWith(
			'auth_token',
			newToken,
			expect.objectContaining({ path: '/' })
		);
		expect(cookies.set).toHaveBeenCalledWith(
			'refresh_token',
			'new-token',
			expect.objectContaining({ path: '/api/auth' })
		);
	});

	it('clears both cookies and returns null when the API rejects the refresh token', async () => {
		const cookies = cookieJar({ refresh_token: 'revoked-refresh-token' });
		const fetchMock = vi.fn(() =>
			Promise.resolve(new Response(JSON.stringify({ success: false }), { status: 401 }))
		);

		const user = await attemptRefresh(cookies as never, url, fetchMock, secret);

		expect(user).toBeNull();
		expect(cookies.delete).toHaveBeenCalledWith('auth_token', { path: '/' });
		expect(cookies.delete).toHaveBeenCalledWith('refresh_token', { path: '/api/auth' });
	});

	it('clears both cookies and returns null when the fetch throws', async () => {
		const cookies = cookieJar({ refresh_token: 'refresh-token' });
		const fetchMock = vi.fn(() => Promise.reject(new Error('network down')));

		const user = await attemptRefresh(cookies as never, url, fetchMock, secret);

		expect(user).toBeNull();
		expect(cookies.delete).toHaveBeenCalledWith('auth_token', { path: '/' });
		expect(cookies.delete).toHaveBeenCalledWith('refresh_token', { path: '/api/auth' });
	});
});

describe('requireSession', () => {
	it('returns the user when present', () => {
		const user = { id: 1, role: USER_ROLE };
		expect(requireSession(user)).toBe(user);
	});

	it('redirects to /login with a session_expired reason when absent', () => {
		expect(() => requireSession(null)).toThrow(
			expect.objectContaining({ status: 303, location: '/login?reason=session_expired' })
		);
	});
});

describe('redirectToLoginOnAuthFailure', () => {
	it('redirects on a 401 response', () => {
		const response = new Response(null, { status: 401 });
		expect(() => redirectToLoginOnAuthFailure(response)).toThrow(
			expect.objectContaining({ status: 303, location: '/login?reason=session_expired' })
		);
	});

	it('does nothing for other statuses', () => {
		const response = new Response(null, { status: 500 });
		expect(() => redirectToLoginOnAuthFailure(response)).not.toThrow();
	});
});
