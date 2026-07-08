import { API_PUBLIC_BASE_URL, buildInternalApiPath, readApiResponse } from '$lib/server/api';
import { redirectToLoginOnAuthFailure } from '$lib/server/auth';
import { mapApiOrder, type ApiOrder, type Order } from '$lib/types';
import { error } from '@sveltejs/kit';

export const loadOrders = async (fetchFn: typeof fetch): Promise<Order[]> => {
	const response = await fetchFn(buildInternalApiPath('/orders'));

	if (!response.ok) {
		redirectToLoginOnAuthFailure(response);
		throw error(response.status, 'Impossible de charger les commandes');
	}

	const { payload } = await readApiResponse<ApiOrder[]>(response);

	if (!Array.isArray(payload?.data)) {
		throw error(502, 'Format de reponse API invalide pour les commandes');
	}

	return payload.data.map((item) => mapApiOrder(item, API_PUBLIC_BASE_URL));
};

export const loadOrderById = async (fetchFn: typeof fetch, orderId: string): Promise<Order> => {
	const response = await fetchFn(buildInternalApiPath(`/orders/${orderId}`));

	if (response.status === 404) {
		throw error(404, 'Commande introuvable');
	}

	if (!response.ok) {
		redirectToLoginOnAuthFailure(response);
		throw error(response.status, 'Impossible de charger la commande');
	}

	const { payload } = await readApiResponse<ApiOrder>(response);

	if (!payload?.data || typeof payload.data !== 'object') {
		throw error(502, 'Format de reponse API invalide pour la commande');
	}

	return mapApiOrder(payload.data, API_PUBLIC_BASE_URL);
};
