import {
	PROMOTION_TYPE_FIXED,
	PROMOTION_TYPE_PERCENTAGE,
	type ApiPromotion,
	type AuthUser
} from '$lib/types';
import {
	buildApiHeaders,
	buildInternalApiPath,
	getApiErrorMessage,
	readApiResponse
} from '$lib/server/api';
import { getFormString, getFormStrings } from '$lib/server/forms';

export type PromotionMutationApiData = ApiPromotion | { ID?: number; id?: number } | null;

export type PromotionAction = 'create-promotion' | 'edit-promotion' | 'delete-promotion';

export type PromotionFormValues = {
	id?: string;
	name: string;
	description: string;
	type: string;
	value: string;
	isActive: 'true' | 'false';
	appliesToAll: 'true' | 'false';
	productIds: string[];
};

export type ParsedPromotionForm = {
	id: string;
	values: PromotionFormValues;
	value: number;
	productIds: number[];
};

export const readPromotionForm = async (request: Request): Promise<ParsedPromotionForm> => {
	const formData = await request.formData();
	const id = getFormString(formData, 'id').trim();
	const name = getFormString(formData, 'name').trim();
	const description = getFormString(formData, 'description').trim();
	const type = getFormString(formData, 'type').trim();
	const valueText = getFormString(formData, 'value').trim();
	const isActive = getFormString(formData, 'is_active', 'false') === 'true' ? 'true' : 'false';
	const appliesToAll =
		getFormString(formData, 'applies_to_all', 'false') === 'true' ? 'true' : 'false';
	const rawProductIds = getFormStrings(formData, 'product_ids')
		.map((entry) => entry.trim())
		.filter((entry) => entry !== '');

	return {
		id,
		values: {
			id,
			name,
			description,
			type,
			value: valueText,
			isActive,
			appliesToAll,
			productIds: rawProductIds
		},
		value: Number(valueText),
		productIds: rawProductIds.map(Number).filter((entry) => Number.isInteger(entry) && entry > 0)
	};
};

export const validatePromotionFormError = (
	values: PromotionFormValues,
	value: number,
	productIds: number[],
	options?: { emptyProductsMessage?: string }
): string | null => {
	if (!values.name || !values.type || values.value === '') {
		return 'Tous les champs promotion sont requis';
	}

	if (values.type !== PROMOTION_TYPE_PERCENTAGE && values.type !== PROMOTION_TYPE_FIXED) {
		return 'Le type de promotion est invalide';
	}

	if (!Number.isFinite(value) || value <= 0) {
		return 'La valeur de promotion doit etre positive';
	}

	if (values.type === PROMOTION_TYPE_PERCENTAGE && value > 100) {
		return 'Le pourcentage ne peut pas depasser 100';
	}

	if (values.appliesToAll !== 'true' && productIds.length === 0) {
		return (
			options?.emptyProductsMessage ?? 'Selectionne au moins un produit ou active la portee globale'
		);
	}

	return null;
};

type MutationEvent = {
	request: Request;
	fetch: typeof fetch;
	locals: { user: AuthUser | null };
};

type PromotionFailAction<TResult> = (
	status: number,
	action: PromotionAction,
	message: string,
	options?: { promotionValues?: PromotionFormValues; promotionId?: string }
) => TResult;

export const createPromotionActions = <TResult>(
	requireActor: (user: AuthUser | null) => void,
	failAction: PromotionFailAction<TResult>,
	options?: { allowAppliesToAll?: boolean; emptyProductsMessage?: string }
) => {
	const allowAppliesToAll = options?.allowAppliesToAll ?? false;
	const emptyProductsMessage = options?.emptyProductsMessage;

	const buildPayload = (values: PromotionFormValues, value: number, productIds: number[]) => {
		const appliesToAll = allowAppliesToAll && values.appliesToAll === 'true';
		return {
			name: values.name,
			description: values.description,
			type: values.type,
			value,
			is_active: values.isActive === 'true',
			applies_to_all: appliesToAll,
			product_ids: appliesToAll ? [] : productIds
		};
	};

	return {
		createPromotion: async ({ request, fetch, locals }: MutationEvent) => {
			requireActor(locals.user);

			const { values, value, productIds } = await readPromotionForm(request);
			const validationMessage = validatePromotionFormError(values, value, productIds, {
				emptyProductsMessage
			});
			if (validationMessage) {
				return failAction(400, 'create-promotion', validationMessage, {
					promotionValues: values
				});
			}

			const response = await fetch(buildInternalApiPath('/promotions'), {
				method: 'POST',
				headers: buildApiHeaders({ contentType: 'application/json' }),
				body: JSON.stringify(buildPayload(values, value, productIds))
			});

			const result = await readApiResponse<PromotionMutationApiData>(response);
			const apiError = getApiErrorMessage(response, result, 'Impossible de creer la promotion');
			if (apiError) {
				return failAction(response.status || 500, 'create-promotion', apiError, {
					promotionValues: values
				});
			}

			return {
				action: 'create-promotion' as const,
				success: 'Promotion ajoutee avec succes'
			};
		},

		updatePromotion: async ({ request, fetch, locals }: MutationEvent) => {
			requireActor(locals.user);

			const { id, values, value, productIds } = await readPromotionForm(request);
			if (!id) {
				return failAction(400, 'edit-promotion', 'Promotion introuvable', {
					promotionValues: values
				});
			}

			const validationMessage = validatePromotionFormError(values, value, productIds, {
				emptyProductsMessage
			});
			if (validationMessage) {
				return failAction(400, 'edit-promotion', validationMessage, {
					promotionValues: values
				});
			}

			const response = await fetch(buildInternalApiPath(`/promotions/${id}`), {
				method: 'PUT',
				headers: buildApiHeaders({ contentType: 'application/json' }),
				body: JSON.stringify(buildPayload(values, value, productIds))
			});

			const result = await readApiResponse<PromotionMutationApiData>(response);
			const apiError = getApiErrorMessage(response, result, 'Impossible de modifier la promotion');
			if (apiError) {
				return failAction(response.status || 500, 'edit-promotion', apiError, {
					promotionValues: values,
					promotionId: id
				});
			}

			return {
				action: 'edit-promotion' as const,
				success: 'Promotion modifiee avec succes'
			};
		},

		deletePromotion: async ({ request, fetch, locals }: MutationEvent) => {
			requireActor(locals.user);

			const promotionId = getFormString(await request.formData(), 'id').trim();
			if (!promotionId) {
				return failAction(400, 'delete-promotion', 'Promotion introuvable', {
					promotionId
				});
			}

			const response = await fetch(buildInternalApiPath(`/promotions/${promotionId}`), {
				method: 'DELETE'
			});

			const result = await readApiResponse<PromotionMutationApiData>(response);
			const apiError = getApiErrorMessage(response, result, 'Impossible de supprimer la promotion');
			if (apiError) {
				return failAction(response.status || 500, 'delete-promotion', apiError, {
					promotionId
				});
			}

			return {
				action: 'delete-promotion' as const,
				success: 'Promotion supprimee avec succes'
			};
		}
	};
};
