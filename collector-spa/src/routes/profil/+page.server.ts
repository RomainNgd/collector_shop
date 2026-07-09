import { loadProfileStats } from '$lib/server/profile';
import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, fetch }) => {
	if (!locals.user) {
		redirect(303, '/login');
	}

	return {
		profile: await loadProfileStats(fetch)
	};
};
