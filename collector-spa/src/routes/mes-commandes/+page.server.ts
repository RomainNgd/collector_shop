import { loadOrders } from '$lib/server/orders';
import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, fetch }) => {
	if (!locals.user) {
		redirect(303, '/login');
	}

	return {
		orders: await loadOrders(fetch)
	};
};
