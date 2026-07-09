import { buildInternalApiPath, readApiResponse } from '$lib/server/api';
import { error } from '@sveltejs/kit';

export type ApiProfileStats = {
	email?: string;
	products_bought?: number;
	listings_posted?: number;
	products_sold?: number;
};

export type ProfileStats = {
	email: string;
	productsBought: number;
	listingsPosted: number;
	productsSold: number;
};

export const mapApiProfileStats = (item: ApiProfileStats): ProfileStats => ({
	email: item.email ?? '',
	productsBought: item.products_bought ?? 0,
	listingsPosted: item.listings_posted ?? 0,
	productsSold: item.products_sold ?? 0
});

export const loadProfileStats = async (fetchFn: typeof fetch): Promise<ProfileStats> => {
	const response = await fetchFn(buildInternalApiPath('/profile'));

	if (!response.ok) {
		throw error(response.status, 'Impossible de charger le profil');
	}

	const { payload } = await readApiResponse<ApiProfileStats>(response);

	if (!payload?.data || typeof payload.data !== 'object') {
		throw error(502, 'Format de reponse API invalide pour le profil');
	}

	return mapApiProfileStats(payload.data);
};
