import { describe, expect, it, vi } from 'vitest';

import { loadProducts, loadSellerProducts } from '$lib/server/products';

const okResponse = (data: unknown) =>
	new Response(JSON.stringify({ success: true, data }), { status: 200 });

describe('loadProducts', () => {
	it('loads and maps products from the API', async () => {
		const fetchFn = vi.fn(() =>
			Promise.resolve(
				okResponse([{ ID: 1, name: 'Console', description: 'Retro', price: 10, image: '' }])
			)
		) as unknown as typeof fetch;

		await expect(loadProducts(fetchFn)).resolves.toMatchObject([{ id: 1, name: 'Console' }]);
	});

	it('throws when the API response is not ok', async () => {
		const fetchFn = vi.fn(() =>
			Promise.resolve(new Response(null, { status: 500 }))
		) as unknown as typeof fetch;

		await expect(loadProducts(fetchFn)).rejects.toMatchObject({ status: 500 });
	});

	it('throws when the API payload is not an array', async () => {
		const fetchFn = vi.fn(() => Promise.resolve(okResponse(null))) as unknown as typeof fetch;

		await expect(loadProducts(fetchFn)).rejects.toMatchObject({ status: 502 });
	});
});

describe('loadSellerProducts', () => {
	it('loads and maps seller products from the API', async () => {
		const fetchFn = vi.fn(() =>
			Promise.resolve(
				okResponse([{ ID: 2, name: 'Cards', description: 'TCG', price: 5, image: '' }])
			)
		) as unknown as typeof fetch;

		await expect(loadSellerProducts(fetchFn)).resolves.toMatchObject([{ id: 2, name: 'Cards' }]);
	});

	it('throws when the API response is not ok', async () => {
		const fetchFn = vi.fn(() =>
			Promise.resolve(new Response(null, { status: 503 }))
		) as unknown as typeof fetch;

		await expect(loadSellerProducts(fetchFn)).rejects.toMatchObject({ status: 503 });
	});

	it('throws when the API payload is not an array', async () => {
		const fetchFn = vi.fn(() => Promise.resolve(okResponse(undefined))) as unknown as typeof fetch;

		await expect(loadSellerProducts(fetchFn)).rejects.toMatchObject({ status: 502 });
	});
});
