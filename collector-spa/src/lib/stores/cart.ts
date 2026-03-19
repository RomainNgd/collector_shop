import { browser } from '$app/environment';
import { derived, writable } from 'svelte/store';
import type { CartItem, Product } from '$lib/types';

const STORAGE_KEY = 'collector-shop-cart-v1';

const loadInitialCart = (): CartItem[] => {
	if (!browser) {
		return [];
	}

	try {
		const saved = localStorage.getItem(STORAGE_KEY);
		if (!saved) {
			return [];
		}

		const parsed = JSON.parse(saved) as unknown;
		return Array.isArray(parsed) ? (parsed as CartItem[]) : [];
	} catch {
		return [];
	}
};

const cart = writable<CartItem[]>(loadInitialCart());
export const cartAddPulse = writable(0);

if (browser) {
	cart.subscribe((items) => {
		localStorage.setItem(STORAGE_KEY, JSON.stringify(items));
	});
}

export const cartItems = cart;

export const cartCount = derived(cart, (items) =>
	items.reduce((total, item) => total + item.quantity, 0)
);

export const cartTotal = derived(cart, (items) =>
	items.reduce((total, item) => total + item.product.price * item.quantity, 0)
);

export const addToCart = (product: Product) => {
	cart.update((items) => {
		const existing = items.find((item) => item.product.id === product.id);
		if (!existing) {
			return [...items, { product, quantity: 1 }];
		}

		return items.map((item) =>
			item.product.id === product.id ? { ...item, quantity: item.quantity + 1 } : item
		);
	});

	cartAddPulse.update((n) => n + 1);
};

export const updateQuantity = (productId: number, quantity: number) => {
	cart.update((items) => {
		if (quantity <= 0) {
			return items.filter((item) => item.product.id !== productId);
		}

		return items.map((item) => (item.product.id === productId ? { ...item, quantity } : item));
	});
};

export const removeFromCart = (productId: number) => {
	cart.update((items) => items.filter((item) => item.product.id !== productId));
};

export const clearCart = () => {
	cart.set([]);
};
