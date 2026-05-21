import {
	buildApiHeaders,
	buildInternalApiPath,
	getApiErrorMessage,
	readApiResponse
} from '$lib/server/api';
import { loadOrderById } from '$lib/server/orders';
import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';

type CheckoutSessionApiData = {
	url?: string;
	session_id?: string;
	reused?: boolean;
};

export const load: PageServerLoad = async ({ locals, fetch, params, cookies, url }) => {
	if (!locals.user) {
		redirect(303, '/login');
	}

	const shouldClearCart = cookies.get('clear_cart_once') === '1';
	if (shouldClearCart) {
		cookies.delete('clear_cart_once', { path: '/' });
	}

	return {
		order: await loadOrderById(fetch, params.id),
		shouldClearCart,
		paymentFlow: url.searchParams.get('payment')
	};
};

export const actions: Actions = {
	pay: async ({ locals, fetch, params, url }) => {
		if (!locals.user) {
			redirect(303, '/login');
		}

		const successURL = `${url.origin}/mes-commandes/${params.id}?payment=processing&session_id={CHECKOUT_SESSION_ID}`;
		const cancelURL = `${url.origin}/mes-commandes/${params.id}?payment=cancelled`;

		const response = await fetch(buildInternalApiPath(`/orders/${params.id}/checkout-session`), {
			method: 'POST',
			headers: buildApiHeaders({ contentType: 'application/json' }),
			body: JSON.stringify({
				success_url: successURL,
				cancel_url: cancelURL
			})
		});

		const result = await readApiResponse<CheckoutSessionApiData>(response);
		const apiError = getApiErrorMessage(
			response,
			result,
			'Impossible de demarrer le paiement Stripe'
		);
		if (apiError) {
			return fail(response.status || 500, {
				error: apiError
			});
		}

		const checkoutURL = result.payload?.data?.url;
		if (!checkoutURL) {
			return fail(502, {
				error: "Stripe n'a pas renvoye d'URL de paiement exploitable"
			});
		}

		redirect(303, checkoutURL);
	}
};
