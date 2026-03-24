import { API_BASE_URL, API_PUBLIC_BASE_URL, readApiResponse } from '$lib/server/api';
import { mapApiProduct, type ApiProduct, type Product } from '$lib/types';
import { error } from '@sveltejs/kit';

export const loadProducts = async (fetchFn: typeof fetch): Promise<Product[]> => {
	const response = await fetchFn(`${API_BASE_URL}/products`);

	if (!response.ok) {
		throw error(response.status, 'Impossible de charger les produits');
	}

	const { payload } = await readApiResponse<ApiProduct[]>(response);

	if (!Array.isArray(payload?.data)) {
		throw error(502, 'Format de reponse API invalide');
	}

	return payload.data.map((item) => mapApiProduct(item, API_PUBLIC_BASE_URL));
};
