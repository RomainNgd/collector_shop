import { createHmac } from 'node:crypto';
import { describe, expect, it, vi } from 'vitest';

import { USER_ROLE } from '$lib/types';

const secret = 'hooks-test-secret';

vi.mock('$env/dynamic/private', () => ({
	env: {
		JWT_SECRET: 'hooks-test-secret',
		API_BASE_URL: 'http://go-api:8080/base/'
	}
}));

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

describe('server hooks', () => {
	it('resolves anonymous, invalid and authenticated requests', async () => {
		const { handle } = await import('./hooks.server');
		const resolve = vi.fn((event) =>
			Promise.resolve(new Response(String(event.locals.user?.id ?? 'anonymous')))
		);
		const noopFetch = vi.fn();
		const url = new URL('https://collector.test/');

		const anonymousEvent = {
			cookies: cookieJar(),
			locals: { user: { id: 99, role: USER_ROLE } },
			fetch: noopFetch,
			url
		};
		const anonymousResponse = await handle({ event: anonymousEvent, resolve } as never);
		expect(await anonymousResponse.text()).toBe('anonymous');

		const invalidEvent = {
			cookies: cookieJar({ auth_token: 'invalid' }),
			locals: { user: null },
			fetch: noopFetch,
			url
		};
		await handle({ event: invalidEvent, resolve } as never);
		expect(invalidEvent.cookies.delete).toHaveBeenCalledWith('auth_token', { path: '/' });
		expect(invalidEvent.cookies.delete).toHaveBeenCalledWith('refresh_token', {
			path: '/api/auth'
		});

		const token = signToken({ sub: 7, role: USER_ROLE, exp: Math.floor(Date.now() / 1000) + 60 });
		const authenticatedEvent = {
			cookies: cookieJar({ auth_token: token }),
			locals: { user: null },
			fetch: noopFetch,
			url
		};
		await handle({ event: authenticatedEvent, resolve } as never);
		expect(authenticatedEvent.locals.user).toEqual({ id: 7, role: USER_ROLE, email: undefined });
	});

	it('transparently refreshes an expired access token using the refresh cookie', async () => {
		const { handle } = await import('./hooks.server');
		const resolve = vi.fn((event) =>
			Promise.resolve(new Response(String(event.locals.user?.id ?? 'anonymous')))
		);
		const url = new URL('https://collector.test/');

		const newToken = signToken({
			sub: 9,
			role: USER_ROLE,
			exp: Math.floor(Date.now() / 1000) + 60
		});
		const refreshFetch = vi.fn(() =>
			Promise.resolve(
				new Response(
					JSON.stringify({
						success: true,
						data: { token: newToken, refresh_token: 'new-refresh-token' }
					})
				)
			)
		);

		const event = {
			cookies: cookieJar({ refresh_token: 'old-refresh-token' }),
			locals: { user: null },
			fetch: refreshFetch,
			url
		};
		await handle({ event, resolve } as never);

		expect(refreshFetch).toHaveBeenCalledOnce();
		expect(event.locals.user).toEqual({ id: 9, role: USER_ROLE, email: undefined });
		expect(event.cookies.set).toHaveBeenCalledWith(
			'auth_token',
			newToken,
			expect.objectContaining({ path: '/' })
		);
		expect(event.cookies.set).toHaveBeenCalledWith(
			'refresh_token',
			'new-refresh-token',
			expect.objectContaining({ path: '/api/auth' })
		);
	});

	it('clears both cookies when the transparent refresh fails', async () => {
		const { handle } = await import('./hooks.server');
		const resolve = vi.fn((event) =>
			Promise.resolve(new Response(String(event.locals.user?.id ?? 'anonymous')))
		);
		const url = new URL('https://collector.test/');

		const failingFetch = vi.fn(() =>
			Promise.resolve(new Response(JSON.stringify({ success: false }), { status: 401 }))
		);

		const event = {
			cookies: cookieJar({ refresh_token: 'stale-refresh-token' }),
			locals: { user: null },
			fetch: failingFetch,
			url
		};
		await handle({ event, resolve } as never);

		expect(event.locals.user).toBeNull();
		expect(event.cookies.delete).toHaveBeenCalledWith('auth_token', { path: '/' });
		expect(event.cookies.delete).toHaveBeenCalledWith('refresh_token', { path: '/api/auth' });
	});

	it('passes external fetches through and proxies internal API requests', async () => {
		const { handleFetch } = await import('./hooks.server');
		const externalRequest = new Request('https://example.com/image.png');
		const passthrough = vi.fn(() => Promise.resolve(new Response('external')));
		await handleFetch({
			event: { url: new URL('https://collector.test/'), cookies: { get: vi.fn() } },
			request: externalRequest,
			fetch: passthrough
		} as never);
		expect(passthrough).toHaveBeenCalledWith(externalRequest);

		const proxiedFetch = vi.fn((request: Request) => Promise.resolve(new Response(request.url)));
		await handleFetch({
			event: {
				url: new URL('https://collector.test/'),
				cookies: { get: vi.fn(() => 'auth-token') }
			},
			request: new Request('https://collector.test/api/orders?page=2'),
			fetch: proxiedFetch
		} as never);
		const proxiedRequest = proxiedFetch.mock.calls[0][0];
		expect(proxiedRequest.url).toBe('http://go-api:8080/base/orders?page=2');
		expect(proxiedRequest.headers.get('authorization')).toBe('Bearer auth-token');
	});
});
