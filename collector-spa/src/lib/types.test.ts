import { describe, expect, it } from 'vitest';

import {
	ORDER_STATUS_AWAITING_PAYMENT,
	PROMOTION_TYPE_FIXED,
	PROMOTION_TYPE_PERCENTAGE,
	mapApiOrder,
	mapApiProduct,
	mapApiPromotion,
	type ApiOrder,
	type ApiProduct,
	type ApiPromotion
} from '$lib/types';

describe('API mapping helpers', () => {
	it('maps API products to UI products with effective pricing and promotion summary', () => {
		const apiProduct: ApiProduct = {
			ID: 12,
			name: 'Game Boy',
			description: 'Console portable',
			price: 100,
			effective_price: 85,
			image: 'game-boy.png',
			category_id: 4,
			category: {
				ID: 4,
				name: 'Consoles'
			},
			applied_promotion: {
				id: 7,
				name: 'Collectors',
				type: PROMOTION_TYPE_PERCENTAGE,
				value: 15,
				discount_amount: 15,
				applies_to_all: false
			}
		};

		const product = mapApiProduct(apiProduct, 'https://api.collector.local');

		expect(product).toMatchObject({
			id: 12,
			price: 85,
			basePrice: 100,
			categoryId: 4,
			category: 'Consoles'
		});
		expect(product.imageUrl).toBe('https://api.collector.local/upload/game-boy.png');
		expect(product.promotion).toEqual({
			id: 7,
			name: 'Collectors',
			type: PROMOTION_TYPE_PERCENTAGE,
			value: 15,
			discountAmount: 15,
			appliesToAll: false
		});
	});

	it('falls back to placeholder image and base price when no promotion exists', () => {
		const product = mapApiProduct(
			{
				ID: 1,
				name: 'Binder',
				description: 'Top loader',
				price: 19.5,
				image: '',
				category: 'Accessoires'
			},
			'https://api.collector.local'
		);

		expect(product.basePrice).toBe(19.5);
		expect(product.price).toBe(19.5);
		expect(product.promotion).toBeNull();
		expect(product.imageUrl.startsWith('data:image/svg+xml')).toBe(true);
	});

	it('maps API promotions for the admin UI', () => {
		const apiPromotion: ApiPromotion = {
			ID: 9,
			name: 'Global fixed',
			description: 'Launch campaign',
			type: PROMOTION_TYPE_FIXED,
			value: 5,
			is_active: true,
			applies_to_all: true,
			product_count: 0,
			product_ids: []
		};

		expect(mapApiPromotion(apiPromotion)).toEqual({
			id: 9,
			name: 'Global fixed',
			description: 'Launch campaign',
			type: PROMOTION_TYPE_FIXED,
			value: 5,
			isActive: true,
			appliesToAll: true,
			productIds: [],
			productCount: 0
		});
	});

	it('maps API orders with frozen pricing lines', () => {
		const apiOrder: ApiOrder = {
			ID: 42,
			CreatedAt: '2026-04-03T10:30:00Z',
			status: ORDER_STATUS_AWAITING_PAYMENT,
			currency: 'EUR',
			item_count: 2,
			subtotal: 100,
			discount_total: 15,
			total: 85,
			payment_provider: 'stripe',
			payment_status: 'checkout_open',
			paid_at: null,
			stripe_checkout_expires_at: '2026-04-03T11:00:00Z',
			items: [
				{
					ID: 1,
					product_id: 7,
					product_name: 'Game Boy',
					product_description: 'Console portable',
					product_image: 'game-boy.png',
					category_name: 'Consoles',
					quantity: 2,
					unit_base_price: 50,
					unit_price: 42.5,
					unit_discount: 7.5,
					line_base_total: 100,
					line_discount_total: 15,
					line_total: 85,
					promotion_id: 11,
					promotion_name: 'Promo',
					promotion_type: PROMOTION_TYPE_PERCENTAGE,
					promotion_value: 15,
					promotion_applies_to_all: false
				}
			]
		};

		expect(mapApiOrder(apiOrder, 'https://api.collector.local')).toMatchObject({
			id: 42,
			status: ORDER_STATUS_AWAITING_PAYMENT,
			currency: 'EUR',
			itemCount: 2,
			total: 85,
			paymentProvider: 'stripe',
			paymentStatus: 'checkout_open',
			stripeCheckoutExpiresAt: '2026-04-03T11:00:00Z',
			items: [
				{
					productId: 7,
					productName: 'Game Boy',
					categoryName: 'Consoles',
					lineTotal: 85,
					promotionId: 11,
					promotionType: PROMOTION_TYPE_PERCENTAGE
				}
			]
		});
	});

	it('maps fallback product, promotion and order shapes', () => {
		const productWithObjectCategory = mapApiProduct(
			{
				ID: 2,
				name: 'Cards',
				description: 'Trading cards',
				price: 12,
				image: ' cards.png ',
				CategoryID: 8,
				category: { ID: 8, name: 'TCG' },
				applied_promotion: {
					ID: 3,
					name: 'Fixed',
					type: PROMOTION_TYPE_FIXED,
					value: 2
				}
			},
			'https://api.collector.local'
		);
		expect(productWithObjectCategory).toMatchObject({
			categoryId: 8,
			category: 'TCG',
			imageName: 'cards.png',
			promotion: { id: 3, discountAmount: 0, appliesToAll: false }
		});

		const productWithInvalidPromotion = mapApiProduct(
			{
				ID: 3,
				name: 'Binder',
				description: 'Binder',
				price: 10,
				image: '',
				category_id: 4,
				applied_promotion: { id: 0, name: 'Invalid', type: 'fixed', value: 1 }
			},
			'https://api.collector.local'
		);
		expect(productWithInvalidPromotion).toMatchObject({ categoryId: 4, promotion: null });

		expect(mapApiPromotion({ ID: 4, name: 'Defaults', type: 'unknown', value: 1 })).toMatchObject({
			description: '',
			type: PROMOTION_TYPE_PERCENTAGE,
			isActive: false,
			appliesToAll: false,
			productIds: [],
			productCount: 0
		});

		const fallbackOrder = mapApiOrder(
			{
				ID: 5,
				status: 'unknown',
				subtotal: 20,
				discount_total: 0,
				total: 20,
				payment_provider: ' ',
				payment_status: '',
				paid_at: '',
				stripe_checkout_expires_at: '',
				items: [
					{
						ID: 1,
						product_name: 'Binder',
						product_description: 'Binder',
						product_image: '',
						quantity: 2,
						unit_base_price: 10,
						unit_price: 10,
						unit_discount: 0,
						line_base_total: 20,
						line_discount_total: 0,
						line_total: 20,
						promotion_id: Number.NaN,
						promotion_value: Number.NaN
					}
				]
			},
			'https://api.collector.local'
		);
		expect(fallbackOrder).toMatchObject({
			createdAt: '',
			status: ORDER_STATUS_AWAITING_PAYMENT,
			currency: 'EUR',
			itemCount: 2,
			paymentProvider: null,
			paymentStatus: null,
			paidAt: null,
			stripeCheckoutExpiresAt: null,
			items: [{ productId: 0, categoryName: 'non-classe', promotionId: null }]
		});
	});
});
