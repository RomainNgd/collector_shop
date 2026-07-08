import { describe, expect, it, vi } from 'vitest';

import { USER_ROLE } from '$lib/types';
import { load as loadHome } from './+page.server';
import { load as loadCatalogue } from './catalogue/+page.server';
import { load as loadOrdersPage } from './mes-commandes/+page.server';
import { actions as orderActions, load as loadOrderPage } from './mes-commandes/[id]/+page.server';
import { load as loadProductPage } from './produit/[id]/+page.server';

const user = { id: 1, role: USER_ROLE };
const apiResponse = (data: unknown, status = 200) =>
	new Response(JSON.stringify({ success: status < 400, data }), { status });
const fetchReturning = (response: Response) =>
	vi.fn(() => Promise.resolve(response)) as unknown as typeof fetch;

const apiProduct = {
	ID: 4,
	name: 'Console',
	description: 'Retro',
	price: 50,
	image: ''
};
const apiOrder = {
	ID: 8,
	CreatedAt: '2026-07-05T10:00:00Z',
	status: 'awaiting_payment',
	currency: 'EUR',
	item_count: 0,
	subtotal: 0,
	discount_total: 0,
	total: 0,
	items: []
};

describe('catalogue pages', () => {
	it('loads home, catalogue and product detail data', async () => {
		await expect(
			loadHome({ fetch: fetchReturning(apiResponse([apiProduct])) } as never)
		).resolves.toMatchObject({
			products: [{ id: 4 }]
		});
		await expect(
			loadCatalogue({ fetch: fetchReturning(apiResponse([apiProduct])) } as never)
		).resolves.toMatchObject({ products: [{ id: 4 }] });
		await expect(
			loadProductPage({
				params: { id: '4' },
				fetch: fetchReturning(apiResponse(apiProduct)),
				locals: { user: null }
			} as never)
		).resolves.toMatchObject({ product: { id: 4 } });
	});

	it('rejects missing product details', async () => {
		await expect(
			loadProductPage({
				params: { id: '99' },
				fetch: fetchReturning(apiResponse(null, 404))
			} as never)
		).rejects.toMatchObject({ status: 404 });
		await expect(
			loadProductPage({ params: { id: '99' }, fetch: fetchReturning(apiResponse(null)) } as never)
		).rejects.toMatchObject({ status: 404 });
	});

	it('flags a product as owned by the logged-in seller', async () => {
		const ownProduct = { ...apiProduct, seller_id: 1 };
		await expect(
			loadProductPage({
				params: { id: '4' },
				fetch: fetchReturning(apiResponse(ownProduct)),
				locals: { user }
			} as never)
		).resolves.toMatchObject({ isOwnProduct: true });

		await expect(
			loadProductPage({
				params: { id: '4' },
				fetch: fetchReturning(apiResponse(ownProduct)),
				locals: { user: { id: 2, role: USER_ROLE } }
			} as never)
		).resolves.toMatchObject({ isOwnProduct: false });

		await expect(
			loadProductPage({
				params: { id: '4' },
				fetch: fetchReturning(apiResponse(ownProduct)),
				locals: { user: null }
			} as never)
		).resolves.toMatchObject({ isOwnProduct: false });
	});
});

describe('order pages', () => {
	it('protects and loads the order list', async () => {
		await expect(
			loadOrdersPage({ locals: { user: null }, fetch: vi.fn() } as never)
		).rejects.toMatchObject({ status: 303, location: '/login' });
		await expect(
			loadOrdersPage({ locals: { user }, fetch: fetchReturning(apiResponse([apiOrder])) } as never)
		).resolves.toMatchObject({ orders: [{ id: 8 }] });
	});

	it('loads an order and consumes the cart clearing cookie', async () => {
		const cookies = { get: vi.fn(() => '1'), delete: vi.fn() };
		const result = await loadOrderPage({
			locals: { user },
			fetch: fetchReturning(apiResponse(apiOrder)),
			params: { id: '8' },
			cookies,
			url: new URL('http://localhost/mes-commandes/8?payment=processing')
		} as never);

		expect(result).toMatchObject({
			order: { id: 8 },
			shouldClearCart: true,
			paymentFlow: 'processing'
		});
		expect(cookies.delete).toHaveBeenCalledOnce();
	});

	it('starts Stripe checkout and handles invalid responses', async () => {
		const pay = orderActions.pay!;
		const event = {
			locals: { user },
			params: { id: '8' },
			url: new URL('https://collector.test/mes-commandes/8')
		};

		await expect(
			pay({
				...event,
				fetch: fetchReturning(
					apiResponse({ url: 'https://checkout.stripe.test/session', session_id: 'cs_test' })
				)
			} as never)
		).rejects.toMatchObject({ status: 303, location: 'https://checkout.stripe.test/session' });

		await expect(
			pay({ ...event, fetch: fetchReturning(apiResponse({})) } as never)
		).resolves.toMatchObject({ status: 502 });
		await expect(
			pay({
				...event,
				fetch: fetchReturning(
					new Response(
						JSON.stringify({ success: false, error: { message: 'Stripe unavailable' } }),
						{
							status: 503
						}
					)
				)
			} as never)
		).resolves.toMatchObject({ status: 503 });
	});
});
