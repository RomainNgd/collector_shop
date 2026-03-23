import { API_BASE_URL, API_PUBLIC_BASE_URL, readApiResponse } from '$lib/server/api';
import { error } from '@sveltejs/kit';
import { mapApiProduct, type ApiProduct } from '$lib/types';

export const load = async ({ fetch }) => {
	const response = await fetch(`${API_BASE_URL}/products`);

	if (!response.ok) {
		throw error(response.status, 'Impossible de charger les produits');
	}

	const { payload } = await readApiResponse<ApiProduct[]>(response);

	if (!Array.isArray(payload?.data)) {
		throw error(502, 'Format de reponse API invalide');
	}

	const products = payload.data.map((item) => mapApiProduct(item, API_PUBLIC_BASE_URL));

	return { products };
};
