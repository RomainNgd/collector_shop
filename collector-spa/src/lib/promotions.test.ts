import { describe, expect, it } from 'vitest';

import {
	formatPrice,
	formatPromotionScope,
	formatPromotionValue,
	getPromotionBadgeLabel,
	hasProductPromotion
} from '$lib/promotions';
import {
	PROMOTION_TYPE_FIXED,
	PROMOTION_TYPE_PERCENTAGE,
	type Product,
	type Promotion
} from '$lib/types';

const buildProduct = (overrides: Partial<Product> = {}): Product => ({
	id: 1,
	name: 'Console',
	description: 'Edition collector',
	price: 90,
	basePrice: 100,
	imageUrl: '/console.png',
	imageName: 'console.png',
	categoryId: 2,
	category: 'Consoles',
	promotion: {
		id: 3,
		name: 'Spring',
		type: PROMOTION_TYPE_PERCENTAGE,
		value: 10,
		discountAmount: 10,
		appliesToAll: false
	},
	...overrides
});

describe('promotion helpers', () => {
	it('detects when a product has a visible promotion', () => {
		expect(hasProductPromotion(buildProduct())).toBe(true);
		expect(
			hasProductPromotion(
				buildProduct({
					price: 100,
					basePrice: 100,
					promotion: null
				})
			)
		).toBe(false);
	});

	it('formats promotion values for percentage and fixed amounts', () => {
		expect(
			formatPromotionValue({
				type: PROMOTION_TYPE_PERCENTAGE,
				value: 12.5
			})
		).toBe('-12,5%');
		expect(
			formatPromotionValue({
				type: PROMOTION_TYPE_FIXED,
				value: 5
			})
		).toBe('-5,00 €');
	});

	it('builds a badge label only when a promotion exists', () => {
		expect(getPromotionBadgeLabel(buildProduct().promotion)).toBe('-10%');
		expect(getPromotionBadgeLabel(null)).toBeNull();
	});

	it('formats promotion scope labels', () => {
		const globalPromotion: Pick<Promotion, 'appliesToAll' | 'productCount'> = {
			appliesToAll: true,
			productCount: 0
		};
		const targetedPromotion: Pick<Promotion, 'appliesToAll' | 'productCount'> = {
			appliesToAll: false,
			productCount: 2
		};

		expect(formatPromotionScope(globalPromotion)).toBe('Globale');
		expect(formatPromotionScope(targetedPromotion)).toBe('2 produits');
	});

	it('formats prices in euro', () => {
		expect(formatPrice(19.99)).toBe('19,99 €');
	});
});
