import { describe, expect, it, vi } from 'vitest';

import { USER_ROLE } from '$lib/types';
import { actions } from './+page.server';

const invokeCreateOrder = (options: {
	items?: string;
	user?: App.Locals['user'];
	response?: Response;
}) => {
	const action = actions.createOrder;
	if (!action) {
		throw new Error('Missing createOrder action');
	}

	const cookies = { set: vi.fn() };
	const fetch = vi.fn(() =>
		Promise.resolve(
			options.response ??
				new Response(JSON.stringify({ success: true, data: { ID: 12 } }), { status: 201 })
		)
	) as unknown as typeof globalThis.fetch;
	const request = new Request('http://localhost/panier', {
		method: 'POST',
		body: new URLSearchParams({ items: options.items ?? '' })
	});

	return {
		cookies,
		fetch,
		result: action({
			request,
			fetch,
			cookies,
			locals: { user: options.user === undefined ? { id: 1, role: USER_ROLE } : options.user }
		} as never)
	};
};

describe('cart checkout action', () => {
	it('rejects unauthenticated and invalid carts', async () => {
		await expect(invokeCreateOrder({ user: null }).result).rejects.toMatchObject({
			status: 303,
			location: '/login'
		});
		await expect(invokeCreateOrder({ items: 'not-json' }).result).resolves.toMatchObject({
			status: 400
		});
		await expect(
			invokeCreateOrder({ items: JSON.stringify([{ product_id: 0, quantity: -1 }]) }).result
		).resolves.toMatchObject({ status: 400 });
	});

	it('creates an order, marks the cart for clearing and redirects', async () => {
		const execution = invokeCreateOrder({
			items: JSON.stringify([{ product_id: 4, quantity: 2 }])
		});

		await expect(execution.result).rejects.toMatchObject({
			status: 303,
			location: '/mes-commandes/12'
		});
		expect(execution.fetch).toHaveBeenCalledOnce();
		expect(execution.cookies.set).toHaveBeenCalledWith(
			'clear_cart_once',
			'1',
			expect.objectContaining({ path: '/' })
		);
	});

	it('returns API and malformed success errors', async () => {
		await expect(
			invokeCreateOrder({
				items: JSON.stringify([{ product_id: 4, quantity: 1 }]),
				response: new Response(
					JSON.stringify({ success: false, error: { message: 'Stock insuffisant' } }),
					{ status: 409 }
				)
			}).result
		).resolves.toMatchObject({ status: 409, data: { error: 'Stock insuffisant' } });

		await expect(
			invokeCreateOrder({
				items: JSON.stringify([{ product_id: 4, quantity: 1 }]),
				response: new Response(JSON.stringify({ success: true, data: {} }), { status: 201 })
			}).result
		).resolves.toMatchObject({ status: 500 });
	});
});
