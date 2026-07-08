import {
	createProductActions,
	type ProductAction,
	type ProductFormValues
} from '$lib/server/productMutations';
import {
	createPromotionActions,
	type PromotionAction,
	type PromotionFormValues
} from '$lib/server/promotionMutations';
import { loadSellerDashboardData, requireSeller } from '$lib/server/sellerDashboard';
import { fail } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';

type SellerAction = ProductAction | PromotionAction;

const failSellerAction = (
	status: number,
	action: SellerAction,
	message: string,
	options?: {
		productValues?: ProductFormValues;
		productId?: string;
		promotionValues?: PromotionFormValues;
		promotionId?: string;
	}
) =>
	fail(status, {
		action,
		error: message,
		values: options?.productValues,
		productId: options?.productId,
		promotionValues: options?.promotionValues,
		promotionId: options?.promotionId
	});

export const load: PageServerLoad = async ({ locals, fetch }) => {
	requireSeller(locals.user);

	return loadSellerDashboardData(fetch);
};

export const actions: Actions = {
	...createProductActions(requireSeller, failSellerAction),
	...createPromotionActions(requireSeller, failSellerAction, {
		allowAppliesToAll: false,
		emptyProductsMessage: 'Selectionne au moins un de tes produits'
	})
};
