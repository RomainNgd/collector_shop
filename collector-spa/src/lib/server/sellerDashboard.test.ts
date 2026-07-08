import { describe, expect, it, vi } from 'vitest';

import { ADMIN_ROLE, USER_ROLE } from '$lib/types';
import { loadSellerDashboardData, requireSeller } from '$lib/server/sellerDashboard';

const apiResponse = (data: unknown, status = 200) =>
	new Response(JSON.stringify({ success: status < 400, data }), { status });

const fetchByPath = (responses: Record<string, Response>) =>
	vi.fn((input: RequestInfo | URL) => {
		const url = typeof input === 'string' ? input : input.toString();
		const match = Object.entries(responses).find(([path]) => url.endsWith(path));
		return Promise.resolve(match ? match[1] : apiResponse(null, 404));
	}) as unknown as typeof fetch;

describe('requireSeller', () => {
	it('redirects anonymous users to login', () => {
		expect(() => requireSeller(null)).toThrowError(
			expect.objectContaining({ status: 303, location: '/login' })
		);
	});

	it('redirects admins to the administration page', () => {
		expect(() => requireSeller({ id: 1, role: ADMIN_ROLE })).toThrowError(
			expect.objectContaining({ status: 303, location: '/administration' })
		);
	});

	it('allows sellers through', () => {
		expect(() => requireSeller({ id: 1, role: USER_ROLE })).not.toThrow();
	});
});

describe('loadSellerDashboardData', () => {
	const okFetch = () =>
		fetchByPath({
			'/seller/products': apiResponse([
				{ ID: 4, name: 'Console', description: 'Retro', price: 10, image: '' }
			]),
			'/categories': apiResponse([{ ID: 3, name: 'Consoles' }]),
			'/promotions': apiResponse([{ ID: 9, name: 'Promo', type: 'percentage', value: 10 }]),
			'/seller/stats': apiResponse({ total_revenue: 42.5, total_sales: 3, product_count: 1 })
		});

	it('loads and combines products, categories, promotions and stats', async () => {
		await expect(loadSellerDashboardData(okFetch())).resolves.toMatchObject({
			products: [{ id: 4 }],
			categories: [{ id: 3 }],
			promotions: [{ id: 9 }],
			stats: { totalRevenue: 42.5, totalSales: 3, productCount: 1 }
		});
	});

	it('throws when promotions request fails', async () => {
		const fetchFn = fetchByPath({
			'/seller/products': apiResponse([]),
			'/categories': apiResponse([]),
			'/promotions': new Response(null, { status: 500 }),
			'/seller/stats': apiResponse({ total_revenue: 0, total_sales: 0, product_count: 0 })
		});

		await expect(loadSellerDashboardData(fetchFn)).rejects.toMatchObject({ status: 500 });
	});

	it('throws when promotions payload is not an array', async () => {
		const fetchFn = fetchByPath({
			'/seller/products': apiResponse([]),
			'/categories': apiResponse([]),
			'/promotions': apiResponse(null),
			'/seller/stats': apiResponse({ total_revenue: 0, total_sales: 0, product_count: 0 })
		});

		await expect(loadSellerDashboardData(fetchFn)).rejects.toMatchObject({ status: 502 });
	});

	it('throws when stats request fails', async () => {
		const fetchFn = fetchByPath({
			'/seller/products': apiResponse([]),
			'/categories': apiResponse([]),
			'/promotions': apiResponse([]),
			'/seller/stats': new Response(null, { status: 503 })
		});

		await expect(loadSellerDashboardData(fetchFn)).rejects.toMatchObject({ status: 503 });
	});

	it('throws when stats payload is missing', async () => {
		const fetchFn = fetchByPath({
			'/seller/products': apiResponse([]),
			'/categories': apiResponse([]),
			'/promotions': apiResponse([]),
			'/seller/stats': apiResponse(null)
		});

		await expect(loadSellerDashboardData(fetchFn)).rejects.toMatchObject({ status: 502 });
	});
});
