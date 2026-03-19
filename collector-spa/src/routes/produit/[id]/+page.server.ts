import { API_BASE_URL, readApiResponse } from '$lib/server/api';
import { error } from '@sveltejs/kit';
import { mapApiProduct, type ApiProduct } from '$lib/types';

export const load = async ({ params, fetch }) => {
	const response = await fetch(`${API_BASE_URL}/products/${params.id}`);

	if (!response.ok) {
		throw error(404, 'Produit introuvable');
	}

	const { payload } = await readApiResponse<ApiProduct>(response);

	if (!payload?.data) {
		throw error(404, 'Produit introuvable');
	}

	const product = mapApiProduct(payload.data, API_BASE_URL);

	return { product };
};
