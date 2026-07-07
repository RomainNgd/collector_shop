import { get } from 'svelte/store';
import { beforeEach, describe, expect, it } from 'vitest';

import type { Product } from '$lib/types';
import {
	addToCart,
	cartAddPulse,
	cartCount,
	cartItems,
	cartTotal,
	clearCart,
	removeFromCart,
	updateQuantity
} from './cart';

const product = (id: number, price = 10): Product => ({
	id,
	name: `Product ${id}`,
	description: 'Description',
	price,
	basePrice: price,
	imageUrl: '/image.png',
	imageName: 'image.png',
	categoryId: 1,
	category: 'Category',
	sellerId: 7,
	sellerEmail: 'seller@test.local',
	stock: 4,
	isActive: true,
	promotion: null
});

describe('cart store', () => {
	beforeEach(() => {
		clearCart();
		cartAddPulse.set(0);
	});

	it('adds products and increments an existing line', () => {
		addToCart(product(1, 12));
		addToCart(product(2, 5));
		addToCart(product(1, 12));

		expect(get(cartItems)).toEqual([
			{ product: product(1, 12), quantity: 2 },
			{ product: product(2, 5), quantity: 1 }
		]);
		expect(get(cartCount)).toBe(3);
		expect(get(cartTotal)).toBe(29);
		expect(get(cartAddPulse)).toBe(3);
	});

	it('updates, removes and clears lines', () => {
		addToCart(product(1));
		addToCart(product(2));
		updateQuantity(1, 4);
		expect(get(cartCount)).toBe(5);

		updateQuantity(1, 0);
		expect(get(cartItems).map((item) => item.product.id)).toEqual([2]);
		removeFromCart(2);
		expect(get(cartItems)).toEqual([]);

		addToCart(product(3));
		clearCart();
		expect(get(cartItems)).toEqual([]);
	});
});
