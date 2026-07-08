import { PROMOTION_TYPE_FIXED, PROMOTION_TYPE_PERCENTAGE, type ApiProduct } from '$lib/types';
import { buildInternalApiPath, getApiErrorMessage, readApiResponse } from '$lib/server/api';
import { getFormString } from '$lib/server/forms';

export type ProductMutationApiData = ApiProduct | { ID?: number; id?: number } | null;

export type ProductFormValues = {
	id?: string;
	name: string;
	description: string;
	price: string;
	stock: string;
	categoryId: string;
	isActive: 'true' | 'false';
	promotionActive: 'true' | 'false';
	promotionType: string;
	promotionValue: string;
	currentImageName?: string;
	removeImage?: 'true';
};

export type ParsedProductForm = {
	id: string;
	values: ProductFormValues;
	price: number;
	imageFile: File | null;
	removeImage: boolean;
};

export const getImageFile = (entry: FormDataEntryValue | null): File | null => {
	if (!(entry instanceof File) || entry.size === 0) {
		return null;
	}

	return entry;
};

export const validateImageFile = (file: File | null) => {
	if (!file) {
		return null;
	}

	return file.type.startsWith('image/') ? null : 'Le fichier choisi doit etre une image';
};

export const extractEntityId = (
	payload: { ID?: number; id?: number } | null | undefined
): number | null => {
	if (!payload || typeof payload !== 'object') {
		return null;
	}

	if ('ID' in payload && typeof payload.ID === 'number' && Number.isFinite(payload.ID)) {
		return payload.ID;
	}

	if ('id' in payload && typeof payload.id === 'number' && Number.isFinite(payload.id)) {
		return payload.id;
	}

	return null;
};

export const readProductForm = async (request: Request): Promise<ParsedProductForm> => {
	const formData = await request.formData();
	const id = getFormString(formData, 'id').trim();
	const name = getFormString(formData, 'name').trim();
	const description = getFormString(formData, 'description').trim();
	const priceValue = getFormString(formData, 'price').trim();
	const stockValue = getFormString(formData, 'stock', '1').trim();
	const categoryId = getFormString(formData, 'category_id').trim();
	const isActive = getFormString(formData, 'is_active', 'true') === 'true' ? 'true' : 'false';
	const promotionActive =
		getFormString(formData, 'promotion_active', 'false') === 'true' ? 'true' : 'false';
	const promotionType = getFormString(formData, 'promotion_type').trim();
	const promotionValue = getFormString(formData, 'promotion_value', '0').trim();
	const currentImageName = getFormString(formData, 'currentImageName').trim();
	const imageFile = getImageFile(formData.get('image'));
	const removeImage = getFormString(formData, 'removeImage') === 'true';

	return {
		id,
		values: {
			id,
			name,
			description,
			price: priceValue,
			stock: stockValue,
			categoryId,
			isActive,
			promotionActive,
			promotionType,
			promotionValue,
			currentImageName,
			removeImage: removeImage ? 'true' : undefined
		},
		price: Number(priceValue),
		imageFile,
		removeImage
	};
};

export const validateProductFormError = (
	values: ProductFormValues,
	price: number,
	imageFile: File | null
): string | null => {
	if (!values.name || !values.description || values.price === '' || values.stock === '') {
		return 'Tous les champs produit sont requis';
	}

	if (values.categoryId === '') {
		return 'Tous les champs produit sont requis';
	}

	if (!Number.isFinite(price) || price <= 0) {
		return 'Le prix doit etre un nombre positif';
	}

	const numericStock = Number(values.stock);
	if (!Number.isInteger(numericStock) || numericStock <= 0) {
		return 'Le stock doit etre un entier positif';
	}

	const numericCategoryId = Number(values.categoryId);
	if (!Number.isInteger(numericCategoryId) || numericCategoryId <= 0) {
		return 'La categorie selectionnee est invalide';
	}

	const imageError = validateImageFile(imageFile);
	if (imageError) {
		return imageError;
	}

	if (values.promotionActive === 'true') {
		const promotionValue = Number(values.promotionValue);
		if (
			values.promotionType !== PROMOTION_TYPE_PERCENTAGE &&
			values.promotionType !== PROMOTION_TYPE_FIXED
		) {
			return 'Le type de promotion est invalide';
		}
		if (!Number.isFinite(promotionValue) || promotionValue <= 0) {
			return 'La valeur de promotion doit etre positive';
		}
		if (values.promotionType === PROMOTION_TYPE_PERCENTAGE && promotionValue > 100) {
			return 'Le pourcentage ne peut pas depasser 100';
		}
	}

	return null;
};

export const uploadProductImage = async (
	fetchFn: typeof fetch,
	productId: string | number,
	imageFile: File
) => {
	const uploadBlob = new Blob([await imageFile.arrayBuffer()], {
		type: imageFile.type || 'application/octet-stream'
	});
	const imageFormData = new FormData();
	imageFormData.set('image', uploadBlob, imageFile.name || 'upload-image');

	const response = await fetchFn(buildInternalApiPath(`/products/${productId}/image`), {
		method: 'POST',
		body: imageFormData
	});

	const result = await readApiResponse<unknown>(response);
	return getApiErrorMessage(response, result, "Impossible d'envoyer l'image");
};

export const deleteProductImage = async (fetchFn: typeof fetch, productId: string | number) => {
	const response = await fetchFn(buildInternalApiPath(`/products/${productId}/image`), {
		method: 'DELETE'
	});

	const result = await readApiResponse<unknown>(response);
	return getApiErrorMessage(response, result, "Impossible de supprimer l'image");
};
