import { browser } from '$app/environment';
import { derived, writable } from 'svelte/store';
import type { CartItem, Product } from '$lib/types';

const STORAGE_KEY = 'collector-shop-cart-v1';

const normalizeProduct = (product: Product): Product => ({
	...product,
	basePrice: typeof product.basePrice === 'number' ? product.basePrice : product.price,
	promotion: product.promotion ?? null
});

const normalizeCartItems = (items: unknown): CartItem[] => {
	if (!Array.isArray(items)) {
		return [];
	}

	return items.flatMap((item) => {
		if (!item || typeof item !== 'object') {
			return [];
		}

		const quantity = Number((item as { quantity?: unknown }).quantity);
		const product = (item as { product?: Product }).product;

		if (!product || typeof product !== 'object' || !Number.isFinite(quantity) || quantity <= 0) {
			return [];
		}

		return [{ product: normalizeProduct(product), quantity }];
	});
};

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
		return normalizeCartItems(parsed);
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
	const normalizedProduct = normalizeProduct(product);

	cart.update((items) => {
		const existing = items.some((item) => item.product.id === normalizedProduct.id);
		if (!existing) {
			return [...items, { product: normalizedProduct, quantity: 1 }];
		}

		return items.map((item) =>
			item.product.id === normalizedProduct.id ? { ...item, quantity: item.quantity + 1 } : item
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
