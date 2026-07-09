import { describe, expect, it, vi } from 'vitest';

import { USER_ROLE } from '$lib/types';
import { load } from './+page.server';

const user = { id: 1, role: USER_ROLE };
const apiResponse = (data: unknown, status = 200) =>
	new Response(JSON.stringify({ success: status < 400, data }), { status });
const fetchReturning = (response: Response) =>
	vi.fn(() => Promise.resolve(response)) as unknown as typeof fetch;

describe('profil page', () => {
	it('redirects anonymous visitors to the login page', async () => {
		await expect(load({ locals: { user: null }, fetch: vi.fn() } as never)).rejects.toMatchObject({
			status: 303,
			location: '/login'
		});
	});

	it('loads and maps the profile stats', async () => {
		const result = await load({
			locals: { user },
			fetch: fetchReturning(
				apiResponse({
					email: 'collector@example.com',
					products_bought: 3,
					listings_posted: 2,
					products_sold: 4
				})
			)
		} as never);

		expect(result).toEqual({
			profile: {
				email: 'collector@example.com',
				productsBought: 3,
				listingsPosted: 2,
				productsSold: 4
			}
		});
	});

	it('defaults missing stats to zero', async () => {
		await expect(
			load({ locals: { user }, fetch: fetchReturning(apiResponse({})) } as never)
		).resolves.toEqual({
			profile: { email: '', productsBought: 0, listingsPosted: 0, productsSold: 0 }
		});
	});

	it('propagates API errors', async () => {
		await expect(
			load({ locals: { user }, fetch: fetchReturning(apiResponse(null, 500)) } as never)
		).rejects.toMatchObject({ status: 500 });
	});

	it('rejects invalid API payloads', async () => {
		await expect(
			load({ locals: { user }, fetch: fetchReturning(apiResponse(null)) } as never)
		).rejects.toMatchObject({ status: 502 });
	});
});
