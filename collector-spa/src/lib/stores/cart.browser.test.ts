import { get } from 'svelte/store';
import { describe, expect, it, vi } from 'vitest';

vi.mock('$app/environment', () => ({ browser: true }));

describe('cart browser persistence', () => {
	it('loads, normalizes and persists valid local storage entries', async () => {
		const stored = JSON.stringify([
			null,
			{ product: null, quantity: 1 },
			{ product: { id: 1 }, quantity: 0 },
			{
				product: {
					id: 4,
					name: 'Console',
					description: 'Retro',
					price: 20,
					imageUrl: '/console.png',
					imageName: 'console.png',
					categoryId: 1,
					category: 'Consoles'
				},
				quantity: 2
			}
		]);
		const localStorage = {
			getItem: vi.fn(() => stored),
			setItem: vi.fn()
		};
		vi.stubGlobal('localStorage', localStorage);

		const { cartItems } = await import('./cart');
		expect(get(cartItems)).toEqual([
			expect.objectContaining({
				quantity: 2,
				product: expect.objectContaining({ id: 4, basePrice: 20, promotion: null })
			})
		]);
		expect(localStorage.getItem).toHaveBeenCalledWith('collector-shop-cart-v1');
		expect(localStorage.setItem).toHaveBeenCalled();
	});
});
