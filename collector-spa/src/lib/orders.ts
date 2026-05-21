import {
	ORDER_STATUS_AWAITING_PAYMENT,
	ORDER_STATUS_CANCELLED,
	ORDER_STATUS_DELIVERED,
	ORDER_STATUS_PREPARATION,
	ORDER_STATUS_SHIPPING,
	type OrderStatus
} from '$lib/types';

export const getOrderStatusLabel = (status: OrderStatus) => {
	switch (status) {
		case ORDER_STATUS_PREPARATION:
			return 'Preparation';
		case ORDER_STATUS_SHIPPING:
			return 'En cours de livraison';
		case ORDER_STATUS_DELIVERED:
			return 'Livree';
		case ORDER_STATUS_CANCELLED:
			return 'Annulee';
		default:
			return 'En attente de paiement';
	}
};

export const getOrderStatusTone = (status: OrderStatus) => {
	switch (status) {
		case ORDER_STATUS_PREPARATION:
			return 'status-pill-preparation';
		case ORDER_STATUS_SHIPPING:
			return 'status-pill-shipping';
		case ORDER_STATUS_DELIVERED:
			return 'status-pill-delivered';
		case ORDER_STATUS_CANCELLED:
			return 'status-pill-cancelled';
		default:
			return 'status-pill-awaiting-payment';
	}
};

export const canOrderBePaid = (status: OrderStatus) => status === ORDER_STATUS_AWAITING_PAYMENT;
