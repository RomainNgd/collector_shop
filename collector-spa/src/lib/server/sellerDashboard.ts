import {
	ADMIN_ROLE,
	mapApiPromotion,
	mapApiSellerStats,
	type ApiPromotion,
	type ApiSellerStats,
	type AuthUser
} from '$lib/types';
import { buildInternalApiPath, readApiResponse } from '$lib/server/api';
import { loadCategories } from '$lib/server/categories';
import { loadSellerProducts } from '$lib/server/products';
import { error, redirect } from '@sveltejs/kit';

export const requireSeller = (user: AuthUser | null) => {
	if (!user) {
		redirect(303, '/login');
	}

	if (user.role === ADMIN_ROLE) {
		redirect(303, '/administration');
	}
};

const loadSellerPromotions = async (fetchFn: typeof fetch) => {
	const response = await fetchFn(buildInternalApiPath('/promotions'));

	if (!response.ok) {
		throw error(response.status, 'Impossible de charger tes promotions');
	}

	const result = await readApiResponse<ApiPromotion[]>(response);

	if (!Array.isArray(result.payload?.data)) {
		throw error(502, 'Format de reponse API invalide pour les promotions');
	}

	return result.payload.data.map(mapApiPromotion);
};

const loadSellerStats = async (fetchFn: typeof fetch) => {
	const response = await fetchFn(buildInternalApiPath('/seller/stats'));

	if (!response.ok) {
		throw error(response.status, 'Impossible de charger tes statistiques');
	}

	const result = await readApiResponse<ApiSellerStats>(response);

	if (!result.payload?.data) {
		throw error(502, 'Format de reponse API invalide pour les statistiques');
	}

	return mapApiSellerStats(result.payload.data);
};

export const loadSellerDashboardData = async (fetchFn: typeof fetch) => {
	const [products, categories, promotions, stats] = await Promise.all([
		loadSellerProducts(fetchFn),
		loadCategories(fetchFn),
		loadSellerPromotions(fetchFn),
		loadSellerStats(fetchFn)
	]);

	return { products, categories, promotions, stats };
};
