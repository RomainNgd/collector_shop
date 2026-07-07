import { buildInternalApiPath, readApiResponse } from '$lib/server/api';
import { mapApiCategory, type ApiCategory, type Category } from '$lib/types';
import { error } from '@sveltejs/kit';

export const loadCategories = async (fetchFn: typeof fetch): Promise<Category[]> => {
	const response = await fetchFn(buildInternalApiPath('/categories'));

	if (!response.ok) {
		throw error(response.status, 'Impossible de charger les categories');
	}

	const { payload } = await readApiResponse<ApiCategory[]>(response);

	if (!Array.isArray(payload?.data)) {
		throw error(502, 'Format de reponse API invalide pour les categories');
	}

	return payload.data.map(mapApiCategory);
};
