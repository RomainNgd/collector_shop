import { describe, expect, it, vi } from 'vitest';

import { USER_ROLE } from '$lib/types';
import { actions as loginActions, load as loginLoad } from '../login/+page.server';
import { actions as registerActions, load as registerLoad } from './register/+page.server';

const cookieJar = () => ({
	get: vi.fn<(name: string) => string | undefined>(),
	set: vi.fn(),
	delete: vi.fn()
});

const actionRequest = (path: string, values: Record<string, string>) =>
	new Request(`http://localhost${path}`, {
		method: 'POST',
		body: new URLSearchParams(values)
	});

const tokenFor = (claims: Record<string, unknown>) =>
	`header.${Buffer.from(JSON.stringify(claims)).toString('base64url')}.signature`;

describe('registration page', () => {
	it('redirects authenticated users and validates registration fields', async () => {
		expect(() => registerLoad({ locals: { user: { id: 1, role: USER_ROLE } } } as never)).toThrow(
			expect.objectContaining({ status: 303 })
		);

		const action = registerActions.default!;
		const invoke = (values: Record<string, string>) =>
			action({
				request: actionRequest('/auth/register', values),
				fetch: vi.fn(),
				cookies: cookieJar(),
				url: new URL('https://collector.test/auth/register')
			} as never);

		await expect(invoke({ email: '' })).resolves.toMatchObject({ status: 400 });
		await expect(
			invoke({ email: 'user@test.local', password: 'short', confirmPassword: 'short' })
		).resolves.toMatchObject({ status: 400 });
		await expect(
			invoke({ email: 'user@test.local', password: 'password1', confirmPassword: 'password2' })
		).resolves.toMatchObject({ status: 400 });
	});

	it('handles registration API errors and success', async () => {
		const action = registerActions.default!;
		const values = {
			email: 'user@test.local',
			password: 'password1',
			confirmPassword: 'password1'
		};
		const conflict = await action({
			request: actionRequest('/auth/register', values),
			fetch: vi.fn(() => Promise.resolve(new Response('{}', { status: 409 }))),
			cookies: cookieJar(),
			url: new URL('http://localhost/auth/register')
		} as never);
		expect(conflict).toMatchObject({ status: 409 });

		const cookies = cookieJar();
		const success = await action({
			request: actionRequest('/auth/register', values),
			fetch: vi.fn(() =>
				Promise.resolve(new Response(JSON.stringify({ success: true, data: { id: 1 } })))
			),
			cookies,
			url: new URL('https://collector.test/auth/register')
		} as never);
		expect(success).toMatchObject({ success: true, email: values.email });
		expect(cookies.set).toHaveBeenCalledWith(
			'register_success_email',
			values.email,
			expect.objectContaining({ secure: true })
		);
	});
});

describe('login page', () => {
	it('reads and clears the registration success cookie', () => {
		const cookies = cookieJar();
		cookies.get.mockReturnValue(' user@test.local ');
		expect(
			loginLoad({
				locals: { user: null },
				cookies,
				url: new URL('http://localhost/login')
			} as never)
		).toEqual({
			registered: true,
			registeredEmail: 'user@test.local',
			sessionExpired: false
		});
		expect(cookies.delete).toHaveBeenCalledOnce();
	});

	it('shows a session-expired message when redirected with that reason', () => {
		const cookies = cookieJar();
		expect(
			loginLoad({
				locals: { user: null },
				cookies,
				url: new URL('http://localhost/login?reason=session_expired')
			} as never)
		).toMatchObject({ sessionExpired: true });
	});

	it('validates credentials and logs in a user', async () => {
		const action = loginActions.default!;
		await expect(
			action({
				request: actionRequest('/login', { email: '', password: '' }),
				fetch: vi.fn(),
				cookies: cookieJar(),
				url: new URL('http://localhost/login')
			} as never)
		).resolves.toMatchObject({ status: 400 });

		const token = tokenFor({ sub: 4, role: USER_ROLE, exp: Math.floor(Date.now() / 1000) + 60 });
		const refreshToken = 'refresh-token-value';
		const cookies = cookieJar();
		const result = action({
			request: actionRequest('/login', { email: 'user@test.local', password: 'password1' }),
			fetch: vi.fn(() =>
				Promise.resolve(
					new Response(
						JSON.stringify({ success: true, data: { token, refresh_token: refreshToken } })
					)
				)
			),
			cookies,
			url: new URL('https://collector.test/login')
		} as never);

		await expect(result).rejects.toMatchObject({ status: 303, location: '/' });
		expect(cookies.set).toHaveBeenCalledWith(
			'auth_token',
			token,
			expect.objectContaining({ secure: true })
		);
		expect(cookies.set).toHaveBeenCalledWith(
			'refresh_token',
			refreshToken,
			expect.objectContaining({ secure: true, path: '/api/auth' })
		);
	});
});
