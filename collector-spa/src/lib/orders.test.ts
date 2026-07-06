import { describe, expect, it } from 'vitest';

import {
	ORDER_STATUS_AWAITING_PAYMENT,
	ORDER_STATUS_CANCELLED,
	ORDER_STATUS_DELIVERED,
	ORDER_STATUS_PREPARATION,
	ORDER_STATUS_SHIPPING
} from '$lib/types';
import { canOrderBePaid, getOrderStatusLabel, getOrderStatusTone } from './orders';

describe('order presentation helpers', () => {
	it.each([
		[ORDER_STATUS_AWAITING_PAYMENT, 'En attente de paiement', 'status-pill-awaiting-payment'],
		[ORDER_STATUS_PREPARATION, 'Preparation', 'status-pill-preparation'],
		[ORDER_STATUS_SHIPPING, 'En cours de livraison', 'status-pill-shipping'],
		[ORDER_STATUS_DELIVERED, 'Livree', 'status-pill-delivered'],
		[ORDER_STATUS_CANCELLED, 'Annulee', 'status-pill-cancelled']
	] as const)('maps %s to its label and tone', (status, label, tone) => {
		expect(getOrderStatusLabel(status)).toBe(label);
		expect(getOrderStatusTone(status)).toBe(tone);
	});

	it('allows payment only for awaiting orders', () => {
		expect(canOrderBePaid(ORDER_STATUS_AWAITING_PAYMENT)).toBe(true);
		expect(canOrderBePaid(ORDER_STATUS_PREPARATION)).toBe(false);
	});
});
