import {
	PROMOTION_TYPE_FIXED,
	PROMOTION_TYPE_PERCENTAGE,
	type ApiProduct,
	type ApiPromotion
} from '$lib/types';
import {
	buildApiHeaders,
	buildInternalApiPath,
	getApiErrorMessage,
	readApiResponse
} from '$lib/server/api';
import { getFormString, getFormStrings } from '$lib/server/forms';
import { loadSellerDashboardData, requireSeller } from '$lib/server/sellerDashboard';
import { fail } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';

type SellerAction =
	| 'create-product'
	| 'edit-product'
	| 'delete-product'
	| 'create-promotion'
	| 'edit-promotion'
	| 'delete-promotion';

type ProductMutationApiData = ApiProduct | { ID?: number; id?: number } | null;
type PromotionMutationApiData = ApiPromotion | { ID?: number; id?: number } | null;

type ProductFormValues = {
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

type PromotionFormValues = {
	id?: string;
	name: string;
	description: string;
	type: string;
	value: string;
	isActive: 'true' | 'false';
	productIds: string[];
};

type ParsedProductForm = {
	id: string;
	values: ProductFormValues;
	price: number;
	imageFile: File | null;
	removeImage: boolean;
};

type ParsedPromotionForm = {
	id: string;
	values: PromotionFormValues;
	value: number;
	productIds: number[];
};

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

const getImageFile = (entry: FormDataEntryValue | null): File | null => {
	if (!(entry instanceof File) || entry.size === 0) {
		return null;
	}

	return entry;
};

const validateImageFile = (file: File | null) => {
	if (!file) {
		return null;
	}

	return file.type.startsWith('image/') ? null : 'Le fichier choisi doit etre une image';
};

const extractEntityId = (
	payload: ProductMutationApiData | PromotionMutationApiData | undefined
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

const readProductForm = async (request: Request): Promise<ParsedProductForm> => {
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

const readPromotionForm = async (request: Request): Promise<ParsedPromotionForm> => {
	const formData = await request.formData();
	const id = getFormString(formData, 'id').trim();
	const name = getFormString(formData, 'name').trim();
	const description = getFormString(formData, 'description').trim();
	const type = getFormString(formData, 'type').trim();
	const valueText = getFormString(formData, 'value').trim();
	const isActive = getFormString(formData, 'is_active', 'false') === 'true' ? 'true' : 'false';
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
			productIds: rawProductIds
		},
		value: Number(valueText),
		productIds: rawProductIds.map(Number).filter((entry) => Number.isInteger(entry) && entry > 0)
	};
};

const validateProductForm = (
	values: ProductFormValues,
	action: Extract<SellerAction, 'create-product' | 'edit-product'>,
	price: number,
	imageFile: File | null
) => {
	if (!values.name || !values.description || values.price === '' || values.stock === '') {
		return failSellerAction(400, action, 'Tous les champs produit sont requis', {
			productValues: values
		});
	}

	if (values.categoryId === '') {
		return failSellerAction(400, action, 'Tous les champs produit sont requis', {
			productValues: values
		});
	}

	if (!Number.isFinite(price) || price <= 0) {
		return failSellerAction(400, action, 'Le prix doit etre un nombre positif', {
			productValues: values
		});
	}

	const numericStock = Number(values.stock);
	if (!Number.isInteger(numericStock) || numericStock <= 0) {
		return failSellerAction(400, action, 'Le stock doit etre un entier positif', {
			productValues: values
		});
	}

	const numericCategoryId = Number(values.categoryId);
	if (!Number.isInteger(numericCategoryId) || numericCategoryId <= 0) {
		return failSellerAction(400, action, 'La categorie selectionnee est invalide', {
			productValues: values
		});
	}

	const imageError = validateImageFile(imageFile);
	if (imageError) {
		return failSellerAction(400, action, imageError, {
			productValues: values
		});
	}

	if (values.promotionActive === 'true') {
		const promotionValue = Number(values.promotionValue);
		if (
			values.promotionType !== PROMOTION_TYPE_PERCENTAGE &&
			values.promotionType !== PROMOTION_TYPE_FIXED
		) {
			return failSellerAction(400, action, 'Le type de promotion est invalide', {
				productValues: values
			});
		}
		if (!Number.isFinite(promotionValue) || promotionValue <= 0) {
			return failSellerAction(400, action, 'La valeur de promotion doit etre positive', {
				productValues: values
			});
		}
		if (values.promotionType === PROMOTION_TYPE_PERCENTAGE && promotionValue > 100) {
			return failSellerAction(400, action, 'Le pourcentage ne peut pas depasser 100', {
				productValues: values
			});
		}
	}

	return null;
};

const validatePromotionForm = (
	values: PromotionFormValues,
	action: Extract<SellerAction, 'create-promotion' | 'edit-promotion'>,
	value: number,
	productIds: number[]
) => {
	if (!values.name || !values.type || values.value === '') {
		return failSellerAction(400, action, 'Tous les champs promotion sont requis', {
			promotionValues: values
		});
	}

	if (values.type !== PROMOTION_TYPE_PERCENTAGE && values.type !== PROMOTION_TYPE_FIXED) {
		return failSellerAction(400, action, 'Le type de promotion est invalide', {
			promotionValues: values
		});
	}

	if (!Number.isFinite(value) || value <= 0) {
		return failSellerAction(400, action, 'La valeur de promotion doit etre positive', {
			promotionValues: values
		});
	}

	if (values.type === PROMOTION_TYPE_PERCENTAGE && value > 100) {
		return failSellerAction(400, action, 'Le pourcentage ne peut pas depasser 100', {
			promotionValues: values
		});
	}

	if (productIds.length === 0) {
		return failSellerAction(400, action, 'Selectionne au moins un de tes produits', {
			promotionValues: values
		});
	}

	return null;
};

const uploadProductImage = async (
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

const deleteProductImage = async (fetchFn: typeof fetch, productId: string | number) => {
	const response = await fetchFn(buildInternalApiPath(`/products/${productId}/image`), {
		method: 'DELETE'
	});

	const result = await readApiResponse<unknown>(response);
	return getApiErrorMessage(response, result, "Impossible de supprimer l'image");
};

export const load: PageServerLoad = async ({ locals, fetch }) => {
	requireSeller(locals.user);

	return loadSellerDashboardData(fetch);
};

export const actions: Actions = {
	createProduct: async ({ request, fetch, locals }) => {
		requireSeller(locals.user);

		const { values, price, imageFile } = await readProductForm(request);
		const validationError = validateProductForm(values, 'create-product', price, imageFile);
		if (validationError) {
			return validationError;
		}

		const response = await fetch(buildInternalApiPath('/products'), {
			method: 'POST',
			headers: buildApiHeaders({ contentType: 'application/json' }),
			body: JSON.stringify({
				name: values.name,
				price,
				stock: Number(values.stock),
				description: values.description,
				category_id: Number(values.categoryId),
				promotion_active: values.promotionActive === 'true',
				promotion_type: values.promotionActive === 'true' ? values.promotionType : '',
				promotion_value: values.promotionActive === 'true' ? Number(values.promotionValue) : 0
			})
		});

		const result = await readApiResponse<ProductMutationApiData>(response);
		const apiError = getApiErrorMessage(response, result, 'Impossible de creer le produit');
		if (apiError) {
			return failSellerAction(response.status || 500, 'create-product', apiError, {
				productValues: values
			});
		}

		if (imageFile) {
			const createdProductId = extractEntityId(result.payload?.data);
			if (createdProductId === null) {
				return failSellerAction(
					500,
					'create-product',
					"Produit cree, mais l'image n'a pas pu etre associee automatiquement",
					{ productValues: values }
				);
			}

			const imageError = await uploadProductImage(fetch, createdProductId, imageFile);
			if (imageError) {
				return failSellerAction(
					500,
					'create-product',
					`Produit cree, mais l'image n'a pas pu etre envoyee: ${imageError}`,
					{ productValues: values }
				);
			}
		}

		return {
			action: 'create-product',
			success: imageFile ? 'Produit et image ajoutes avec succes' : 'Produit ajoute avec succes'
		};
	},

	updateProduct: async ({ request, fetch, locals }) => {
		requireSeller(locals.user);

		const { id, values, price, imageFile, removeImage } = await readProductForm(request);

		if (!id) {
			return failSellerAction(400, 'edit-product', 'Produit introuvable', {
				productValues: values
			});
		}

		const validationError = validateProductForm(values, 'edit-product', price, imageFile);
		if (validationError) {
			return validationError;
		}

		const response = await fetch(buildInternalApiPath(`/products/${id}`), {
			method: 'PUT',
			headers: buildApiHeaders({ contentType: 'application/json' }),
			body: JSON.stringify({
				name: values.name,
				price,
				stock: Number(values.stock),
				is_active: values.isActive === 'true',
				description: values.description,
				image: removeImage && !imageFile ? '' : (values.currentImageName ?? ''),
				category_id: Number(values.categoryId),
				promotion_active: values.promotionActive === 'true',
				promotion_type: values.promotionActive === 'true' ? values.promotionType : '',
				promotion_value: values.promotionActive === 'true' ? Number(values.promotionValue) : 0
			})
		});

		const result = await readApiResponse<ProductMutationApiData>(response);
		const apiError = getApiErrorMessage(response, result, 'Impossible de modifier le produit');
		if (apiError) {
			return failSellerAction(response.status || 500, 'edit-product', apiError, {
				productValues: values
			});
		}

		if (removeImage && !imageFile) {
			const imageDeleteError = await deleteProductImage(fetch, id);
			if (imageDeleteError) {
				return failSellerAction(
					500,
					'edit-product',
					`Produit modifie, mais l'image n'a pas pu etre supprimee: ${imageDeleteError}`,
					{ productValues: values }
				);
			}
		}

		if (imageFile) {
			const imageUploadError = await uploadProductImage(fetch, id, imageFile);
			if (imageUploadError) {
				return failSellerAction(
					500,
					'edit-product',
					`Produit modifie, mais l'image n'a pas pu etre envoyee: ${imageUploadError}`,
					{ productValues: values }
				);
			}
		}

		return {
			action: 'edit-product',
			success:
				imageFile || removeImage
					? 'Produit et image mis a jour avec succes'
					: 'Produit modifie avec succes'
		};
	},

	deleteProduct: async ({ request, fetch, locals }) => {
		requireSeller(locals.user);

		const productId = getFormString(await request.formData(), 'id').trim();

		if (!productId) {
			return failSellerAction(400, 'delete-product', 'Produit introuvable', {
				productId
			});
		}

		const response = await fetch(buildInternalApiPath(`/products/${productId}`), {
			method: 'DELETE'
		});

		const result = await readApiResponse<ProductMutationApiData>(response);
		const apiError = getApiErrorMessage(response, result, 'Impossible de supprimer le produit');
		if (apiError) {
			return failSellerAction(response.status || 500, 'delete-product', apiError, {
				productId
			});
		}

		return {
			action: 'delete-product',
			success: 'Produit supprime avec succes'
		};
	},

	createPromotion: async ({ request, fetch, locals }) => {
		requireSeller(locals.user);

		const { values, value, productIds } = await readPromotionForm(request);
		const validationError = validatePromotionForm(values, 'create-promotion', value, productIds);
		if (validationError) {
			return validationError;
		}

		const response = await fetch(buildInternalApiPath('/promotions'), {
			method: 'POST',
			headers: buildApiHeaders({ contentType: 'application/json' }),
			body: JSON.stringify({
				name: values.name,
				description: values.description,
				type: values.type,
				value,
				is_active: values.isActive === 'true',
				applies_to_all: false,
				product_ids: productIds
			})
		});

		const result = await readApiResponse<PromotionMutationApiData>(response);
		const apiError = getApiErrorMessage(response, result, 'Impossible de creer la promotion');
		if (apiError) {
			return failSellerAction(response.status || 500, 'create-promotion', apiError, {
				promotionValues: values
			});
		}

		return {
			action: 'create-promotion',
			success: 'Promotion ajoutee avec succes'
		};
	},

	updatePromotion: async ({ request, fetch, locals }) => {
		requireSeller(locals.user);

		const { id, values, value, productIds } = await readPromotionForm(request);
		if (!id) {
			return failSellerAction(400, 'edit-promotion', 'Promotion introuvable', {
				promotionValues: values
			});
		}

		const validationError = validatePromotionForm(values, 'edit-promotion', value, productIds);
		if (validationError) {
			return validationError;
		}

		const response = await fetch(buildInternalApiPath(`/promotions/${id}`), {
			method: 'PUT',
			headers: buildApiHeaders({ contentType: 'application/json' }),
			body: JSON.stringify({
				name: values.name,
				description: values.description,
				type: values.type,
				value,
				is_active: values.isActive === 'true',
				applies_to_all: false,
				product_ids: productIds
			})
		});

		const result = await readApiResponse<PromotionMutationApiData>(response);
		const apiError = getApiErrorMessage(response, result, 'Impossible de modifier la promotion');
		if (apiError) {
			return failSellerAction(response.status || 500, 'edit-promotion', apiError, {
				promotionValues: values,
				promotionId: id
			});
		}

		return {
			action: 'edit-promotion',
			success: 'Promotion modifiee avec succes'
		};
	},

	deletePromotion: async ({ request, fetch, locals }) => {
		requireSeller(locals.user);

		const promotionId = getFormString(await request.formData(), 'id').trim();
		if (!promotionId) {
			return failSellerAction(400, 'delete-promotion', 'Promotion introuvable', {
				promotionId
			});
		}

		const response = await fetch(buildInternalApiPath(`/promotions/${promotionId}`), {
			method: 'DELETE'
		});

		const result = await readApiResponse<PromotionMutationApiData>(response);
		const apiError = getApiErrorMessage(response, result, 'Impossible de supprimer la promotion');
		if (apiError) {
			return failSellerAction(response.status || 500, 'delete-promotion', apiError, {
				promotionId
			});
		}

		return {
			action: 'delete-promotion',
			success: 'Promotion supprimee avec succes'
		};
	}
};
