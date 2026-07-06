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

describe('server hooks', () => {
	it('resolves anonymous, invalid and authenticated requests', async () => {
		const { handle } = await import('./hooks.server');
		const resolve = vi.fn((event) =>
			Promise.resolve(new Response(String(event.locals.user?.id ?? 'anonymous')))
		);

		const anonymousEvent = {
			cookies: { get: vi.fn(() => undefined), delete: vi.fn() },
			locals: { user: { id: 99, role: USER_ROLE } }
		};
		const anonymousResponse = await handle({ event: anonymousEvent, resolve } as never);
		expect(await anonymousResponse.text()).toBe('anonymous');

		const invalidEvent = {
			cookies: { get: vi.fn(() => 'invalid'), delete: vi.fn() },
			locals: { user: null }
		};
		await handle({ event: invalidEvent, resolve } as never);
		expect(invalidEvent.cookies.delete).toHaveBeenCalledWith('auth_token', { path: '/' });

		const token = signToken({ sub: 7, role: USER_ROLE, exp: Math.floor(Date.now() / 1000) + 60 });
		const authenticatedEvent = {
			cookies: { get: vi.fn(() => token), delete: vi.fn() },
			locals: { user: null }
		};
		await handle({ event: authenticatedEvent, resolve } as never);
		expect(authenticatedEvent.locals.user).toEqual({ id: 7, role: USER_ROLE, email: undefined });
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
