import { describe, expect, it, vi } from 'vitest';

import { loadCategories } from '$lib/server/categories';

describe('loadCategories', () => {
	it('loads and maps categories from the API', async () => {
		const fetchFn = vi.fn(() =>
			Promise.resolve(
				new Response(JSON.stringify({ success: true, data: [{ ID: 1, name: 'Consoles' }] }), {
					status: 200
				})
			)
		) as unknown as typeof fetch;

		await expect(loadCategories(fetchFn)).resolves.toEqual([
			{ id: 1, name: 'Consoles', description: '' }
		]);
	});

	it('throws when the API response is not ok', async () => {
		const fetchFn = vi.fn(() =>
			Promise.resolve(new Response(null, { status: 500 }))
		) as unknown as typeof fetch;

		await expect(loadCategories(fetchFn)).rejects.toMatchObject({ status: 500 });
	});

	it('throws when the API payload is not an array', async () => {
		const fetchFn = vi.fn(() =>
			Promise.resolve(new Response(JSON.stringify({ success: true, data: null }), { status: 200 }))
		) as unknown as typeof fetch;

		await expect(loadCategories(fetchFn)).rejects.toMatchObject({ status: 502 });
	});
});
