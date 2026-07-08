import { describe, expect, it, vi } from 'vitest';

import { POST } from './+server';

const cookieJar = (values: Record<string, string | undefined> = {}) => ({
	get: vi.fn((name: string) => values[name]),
	set: vi.fn(),
	delete: vi.fn()
});

describe('logout endpoint', () => {
	it('calls the backend logout endpoint and clears both cookies', async () => {
		const cookies = cookieJar({ refresh_token: 'some-refresh-token' });
		const fetchMock = vi.fn<typeof fetch>(() =>
			Promise.resolve(new Response(JSON.stringify({ success: true, data: { success: true } })))
		);

		const result = POST({ cookies, fetch: fetchMock } as never);
		await expect(result).rejects.toMatchObject({ status: 303, location: '/' });

		expect(fetchMock).toHaveBeenCalledOnce();
		const [, init] = fetchMock.mock.calls[0];
		expect(JSON.parse((init as RequestInit).body as string)).toEqual({
			refresh_token: 'some-refresh-token'
		});

		expect(cookies.delete).toHaveBeenCalledWith('auth_token', { path: '/' });
		expect(cookies.delete).toHaveBeenCalledWith('refresh_token', { path: '/api/auth' });
	});

	it('clears cookies even if the backend call fails', async () => {
		const cookies = cookieJar({ refresh_token: 'some-refresh-token' });
		const fetchMock = vi.fn(() => Promise.reject(new Error('network down')));

		const result = POST({ cookies, fetch: fetchMock } as never);
		await expect(result).rejects.toMatchObject({ status: 303, location: '/' });

		expect(cookies.delete).toHaveBeenCalledWith('auth_token', { path: '/' });
		expect(cookies.delete).toHaveBeenCalledWith('refresh_token', { path: '/api/auth' });
	});

	it('skips the backend call and still clears cookies when there is no refresh token', async () => {
		const cookies = cookieJar();
		const fetchMock = vi.fn();

		const result = POST({ cookies, fetch: fetchMock } as never);
		await expect(result).rejects.toMatchObject({ status: 303, location: '/' });

		expect(fetchMock).not.toHaveBeenCalled();
		expect(cookies.delete).toHaveBeenCalledWith('auth_token', { path: '/' });
		expect(cookies.delete).toHaveBeenCalledWith('refresh_token', { path: '/api/auth' });
	});
});
