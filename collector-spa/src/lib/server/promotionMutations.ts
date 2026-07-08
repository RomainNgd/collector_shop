import { PROMOTION_TYPE_FIXED, PROMOTION_TYPE_PERCENTAGE, type ApiPromotion } from '$lib/types';
import { getFormString, getFormStrings } from '$lib/server/forms';

export type PromotionMutationApiData = ApiPromotion | { ID?: number; id?: number } | null;

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
