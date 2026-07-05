import {
	buildApiHeaders,
	buildInternalApiPath,
	getApiErrorMessage,
	readApiResponse
} from '$lib/server/api';
import { getFormString } from '$lib/server/forms';
import type { ApiOrder } from '$lib/types';
import { fail, redirect } from '@sveltejs/kit';
import type { Actions } from './$types';

type CheckoutItem = {
	product_id: number;
	quantity: number;
};

const parseCheckoutItems = (value: string): CheckoutItem[] => {
	if (!value) {
		return [];
	}

	try {
		const parsed = JSON.parse(value) as unknown;
		if (!Array.isArray(parsed)) {
			return [];
		}

		return parsed.flatMap((item) => {
			if (!item || typeof item !== 'object') {
				return [];
			}

			const productId = Number((item as { product_id?: unknown }).product_id);
			const quantity = Number((item as { quantity?: unknown }).quantity);

			if (
				!Number.isInteger(productId) ||
				productId <= 0 ||
				!Number.isInteger(quantity) ||
				quantity <= 0
			) {
				return [];
			}

			return [{ product_id: productId, quantity }];
		});
	} catch {
		return [];
	}
};

export const actions: Actions = {
	createOrder: async ({ request, fetch, locals, cookies }) => {
		if (!locals.user) {
			redirect(303, '/login');
		}

		const formData = await request.formData();
		const items = parseCheckoutItems(getFormString(formData, 'items'));

		if (items.length === 0) {
			return fail(400, {
				error: 'Le panier est vide ou invalide'
			});
		}

		const response = await fetch(buildInternalApiPath('/orders'), {
			method: 'POST',
			headers: buildApiHeaders({ contentType: 'application/json' }),
			body: JSON.stringify({ items })
		});

		const result = await readApiResponse<ApiOrder>(response);
		const apiError = getApiErrorMessage(response, result, 'Impossible de creer la commande');
		if (apiError) {
			return fail(response.status || 500, {
				error: apiError
			});
		}

		const createdOrderId = result.payload?.data?.ID;
		if (typeof createdOrderId !== 'number' || !Number.isFinite(createdOrderId)) {
			return fail(500, {
				error: 'Commande creee, mais impossible de charger le recapitulatif'
			});
		}

		cookies.set('clear_cart_once', '1', {
			path: '/',
			httpOnly: true,
			sameSite: 'lax',
			maxAge: 60
		});

		redirect(303, `/mes-commandes/${createdOrderId}`);
	}
};
