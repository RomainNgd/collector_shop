import {
	ADMIN_ROLE,
	mapApiCategory,
	mapApiProduct,
	mapApiPromotion,
	type ApiCategory,
	type ApiProduct,
	type ApiPromotion,
	type AuthUser
} from '$lib/types';
import { API_PUBLIC_BASE_URL, buildInternalApiPath, readApiResponse } from '$lib/server/api';
import { error, redirect } from '@sveltejs/kit';

export const requireAdmin = (user: AuthUser | null) => {
	if (!user) {
		redirect(303, '/login');
	}

	if (user.role !== ADMIN_ROLE) {
		throw error(403, 'Acces refuse');
	}
};

const loadProducts = async (fetchFn: typeof fetch) => {
	const response = await fetchFn(buildInternalApiPath('/products'));

	if (!response.ok) {
		throw error(response.status, 'Impossible de charger les produits');
	}

	const result = await readApiResponse<ApiProduct[]>(response);

	if (!Array.isArray(result.payload?.data)) {
		throw error(502, 'Format de reponse API invalide pour les produits');
	}

	return result.payload.data.map((item) => mapApiProduct(item, API_PUBLIC_BASE_URL));
};

const loadCategories = async (fetchFn: typeof fetch) => {
	const response = await fetchFn(buildInternalApiPath('/categories'));

	if (!response.ok) {
		throw error(response.status, 'Impossible de charger les categories');
	}

	const result = await readApiResponse<ApiCategory[]>(response);

	if (!Array.isArray(result.payload?.data)) {
		throw error(502, 'Format de reponse API invalide pour les categories');
	}

	return result.payload.data.map(mapApiCategory);
};

const loadPromotions = async (fetchFn: typeof fetch) => {
	const response = await fetchFn(buildInternalApiPath('/promotions'));

	if (!response.ok) {
		throw error(response.status, 'Impossible de charger les promotions');
	}

	const result = await readApiResponse<ApiPromotion[]>(response);

	if (!Array.isArray(result.payload?.data)) {
		throw error(502, 'Format de reponse API invalide pour les promotions');
	}

	return result.payload.data.map(mapApiPromotion);
};

export const loadAdminData = async (fetchFn: typeof fetch) => {
	const [products, categories, promotions] = await Promise.all([
		loadProducts(fetchFn),
		loadCategories(fetchFn),
		loadPromotions(fetchFn)
	]);

	return {
		products,
		categories,
		promotions
	};
};
