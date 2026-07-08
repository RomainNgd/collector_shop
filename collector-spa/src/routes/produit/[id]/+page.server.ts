import { API_PUBLIC_BASE_URL, buildInternalApiPath, readApiResponse } from '$lib/server/api';
import { error } from '@sveltejs/kit';
import { mapApiProduct, type ApiProduct } from '$lib/types';

export const load = async ({ params, fetch, locals }) => {
	const response = await fetch(buildInternalApiPath(`/products/${params.id}`));

	if (!response.ok) {
		throw error(404, 'Produit introuvable');
	}

	const { payload } = await readApiResponse<ApiProduct>(response);

	if (!payload?.data) {
		throw error(404, 'Produit introuvable');
	}

	const product = mapApiProduct(payload.data, API_PUBLIC_BASE_URL);
	const isOwnProduct = locals.user?.id != null && locals.user.id === product.sellerId;

	return { product, isOwnProduct };
};
