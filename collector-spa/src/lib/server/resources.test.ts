import { describe, expect, it, vi } from 'vitest';

import { ADMIN_ROLE, USER_ROLE } from '$lib/types';
import { loadAdminData, requireAdmin } from './admin';
import { loadOrderById, loadOrders } from './orders';
import { loadProducts } from './products';

const apiResponse = (data: unknown, status = 200) =>
	new Response(JSON.stringify({ success: status < 400, data }), {
		status,
		headers: { 'Content-Type': 'application/json' }
	});

const fetchMock = (implementation: (url: string) => Response) =>
	vi.fn((input: RequestInfo | URL) =>
		Promise.resolve(implementation(String(input)))
	) as unknown as typeof fetch;

describe('server resources', () => {
	it('requires an authenticated administrator', () => {
		expect(() => requireAdmin({ id: 1, role: ADMIN_ROLE })).not.toThrow();
		expect(() => requireAdmin({ id: 2, role: USER_ROLE })).toThrowError(
			expect.objectContaining({ status: 403 })
		);
		expect(() => requireAdmin(null)).toThrowError(
			expect.objectContaining({ status: 303, location: '/login' })
		);
	});

	it('loads and maps all administration resources', async () => {
		const fetch = fetchMock((url) => {
			if (url.endsWith('/products')) {
				return apiResponse([
					{ ID: 1, name: 'Console', description: 'Retro', price: 20, image: '', category: 'Games' }
				]);
			}
			if (url.endsWith('/categories')) {
				return apiResponse([{ ID: 2, name: 'Games', description: 'Retro games' }]);
			}
			return apiResponse([
				{
					ID: 3,
					name: 'Promo',
					description: '',
					type: 'fixed',
					value: 2,
					is_active: true,
					applies_to_all: true
				}
			]);
		});

		const result = await loadAdminData(fetch);
		expect(result.products[0]).toMatchObject({ id: 1, name: 'Console' });
		expect(result.categories[0]).toMatchObject({ id: 2, name: 'Games' });
		expect(result.promotions[0]).toMatchObject({ id: 3, name: 'Promo' });
		expect(fetch).toHaveBeenCalledTimes(3);
	});

	it('rejects invalid administration API responses', async () => {
		const fetch = fetchMock((url) =>
			url.endsWith('/products') ? apiResponse(null, 503) : apiResponse([])
		);
		await expect(loadAdminData(fetch)).rejects.toMatchObject({ status: 503 });
	});

	it('loads products and orders', async () => {
		const productFetch = fetchMock(() =>
			apiResponse([{ ID: 1, name: 'Console', description: 'Retro', price: 20, image: '' }])
		);
		expect(await loadProducts(productFetch)).toHaveLength(1);

		const order = {
			ID: 7,
			CreatedAt: '2026-07-05T10:00:00Z',
			status: 'awaiting_payment',
			currency: 'EUR',
			item_count: 0,
			subtotal: 0,
			discount_total: 0,
			total: 0,
			items: []
		};
		const orderFetch = fetchMock((url) => apiResponse(url.endsWith('/orders') ? [order] : order));
		expect(await loadOrders(orderFetch)).toHaveLength(1);
		expect(await loadOrderById(orderFetch, '7')).toMatchObject({ id: 7 });
	});

	it('reports missing and malformed resources', async () => {
		await expect(loadProducts(fetchMock(() => apiResponse(null)))).rejects.toMatchObject({
			status: 502
		});
		await expect(loadOrders(fetchMock(() => apiResponse(null)))).rejects.toMatchObject({
			status: 502
		});
		await expect(
			loadOrderById(
				fetchMock(() => apiResponse(null, 404)),
				'9'
			)
		).rejects.toMatchObject({
			status: 404
		});
		await expect(
			loadOrderById(
				fetchMock(() => apiResponse(null, 500)),
				'9'
			)
		).rejects.toMatchObject({
			status: 500
		});
	});
});
