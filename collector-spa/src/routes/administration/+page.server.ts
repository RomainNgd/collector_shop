import type { ApiCategory } from '$lib/types';
import {
	buildApiHeaders,
	buildInternalApiPath,
	getApiErrorMessage,
	readApiResponse
} from '$lib/server/api';
import { loadAdminData, requireAdmin } from '$lib/server/admin';
import { getFormString } from '$lib/server/forms';
import {
	deleteProductImage,
	extractEntityId,
	readProductForm,
	uploadProductImage,
	validateProductFormError,
	type ProductFormValues,
	type ProductMutationApiData
} from '$lib/server/productMutations';
import {
	readPromotionForm,
	validatePromotionFormError,
	type PromotionFormValues,
	type PromotionMutationApiData
} from '$lib/server/promotionMutations';
import { fail } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';

type AdminAction =
	| 'create-product'
	| 'edit-product'
	| 'delete-product'
	| 'create-category'
	| 'edit-category'
	| 'delete-category'
	| 'create-promotion'
	| 'edit-promotion'
	| 'delete-promotion';

type CategoryMutationApiData = ApiCategory | { ID?: number; id?: number } | null;

type CategoryFormValues = {
	id?: string;
	name: string;
	description: string;
};

const failAdminAction = (
	status: number,
	action: AdminAction,
	message: string,
	options?: {
		productValues?: ProductFormValues;
		productId?: string;
		categoryValues?: CategoryFormValues;
		categoryId?: string;
		promotionValues?: PromotionFormValues;
		promotionId?: string;
	}
) =>
	fail(status, {
		action,
		error: message,
		values: options?.productValues,
		productId: options?.productId,
		categoryValues: options?.categoryValues,
		categoryId: options?.categoryId,
		promotionValues: options?.promotionValues,
		promotionId: options?.promotionId
	});

const readCategoryForm = async (request: Request) => {
	const formData = await request.formData();
	const id = getFormString(formData, 'id').trim();
	const name = getFormString(formData, 'name').trim();
	const description = getFormString(formData, 'description').trim();

	return {
		id,
		values: {
			id,
			name,
			description
		} satisfies CategoryFormValues
	};
};

const validateCategoryForm = (
	values: CategoryFormValues,
	action: Extract<AdminAction, 'create-category' | 'edit-category'>
) => {
	if (!values.name) {
		return failAdminAction(400, action, 'Le nom de categorie est requis', {
			categoryValues: values
		});
	}

	return null;
};

export const load: PageServerLoad = async ({ locals, fetch }) => {
	requireAdmin(locals.user);

	return loadAdminData(fetch);
};

export const actions: Actions = {
	createProduct: async ({ request, fetch, locals }) => {
		requireAdmin(locals.user);

		const { values, price, imageFile } = await readProductForm(request);
		const validationMessage = validateProductFormError(values, price, imageFile);
		if (validationMessage) {
			return failAdminAction(400, 'create-product', validationMessage, { productValues: values });
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
			return failAdminAction(response.status || 500, 'create-product', apiError, {
				productValues: values
			});
		}

		if (imageFile) {
			const createdProductId = extractEntityId(result.payload?.data);
			if (createdProductId === null) {
				return failAdminAction(
					500,
					'create-product',
					"Produit cree, mais l'image n'a pas pu etre associee automatiquement",
					{ productValues: values }
				);
			}

			const imageError = await uploadProductImage(fetch, createdProductId, imageFile);
			if (imageError) {
				return failAdminAction(
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
		requireAdmin(locals.user);

		const { id, values, price, imageFile, removeImage } = await readProductForm(request);

		if (!id) {
			return failAdminAction(400, 'edit-product', 'Produit introuvable', {
				productValues: values
			});
		}

		const validationMessage = validateProductFormError(values, price, imageFile);
		if (validationMessage) {
			return failAdminAction(400, 'edit-product', validationMessage, { productValues: values });
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
			return failAdminAction(response.status || 500, 'edit-product', apiError, {
				productValues: values
			});
		}

		if (removeImage && !imageFile) {
			const imageDeleteError = await deleteProductImage(fetch, id);
			if (imageDeleteError) {
				return failAdminAction(
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
				return failAdminAction(
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
		requireAdmin(locals.user);

		const productId = getFormString(await request.formData(), 'id').trim();

		if (!productId) {
			return failAdminAction(400, 'delete-product', 'Produit introuvable', {
				productId
			});
		}

		const response = await fetch(buildInternalApiPath(`/products/${productId}`), {
			method: 'DELETE'
		});

		const result = await readApiResponse<ProductMutationApiData>(response);
		const apiError = getApiErrorMessage(response, result, 'Impossible de supprimer le produit');
		if (apiError) {
			return failAdminAction(response.status || 500, 'delete-product', apiError, {
				productId
			});
		}

		return {
			action: 'delete-product',
			success: 'Produit supprime avec succes'
		};
	},

	createCategory: async ({ request, fetch, locals }) => {
		requireAdmin(locals.user);

		const { values } = await readCategoryForm(request);
		const validationError = validateCategoryForm(values, 'create-category');
		if (validationError) {
			return validationError;
		}

		const response = await fetch(buildInternalApiPath('/categories'), {
			method: 'POST',
			headers: buildApiHeaders({ contentType: 'application/json' }),
			body: JSON.stringify({
				name: values.name,
				description: values.description
			})
		});

		const result = await readApiResponse<CategoryMutationApiData>(response);
		const apiError = getApiErrorMessage(response, result, 'Impossible de creer la categorie');
		if (apiError) {
			return failAdminAction(response.status || 500, 'create-category', apiError, {
				categoryValues: values
			});
		}

		return {
			action: 'create-category',
			success: 'Categorie ajoutee avec succes'
		};
	},

	updateCategory: async ({ request, fetch, locals }) => {
		requireAdmin(locals.user);

		const { id, values } = await readCategoryForm(request);
		if (!id) {
			return failAdminAction(400, 'edit-category', 'Categorie introuvable', {
				categoryValues: values
			});
		}

		const validationError = validateCategoryForm(values, 'edit-category');
		if (validationError) {
			return validationError;
		}

		const response = await fetch(buildInternalApiPath(`/categories/${id}`), {
			method: 'PUT',
			headers: buildApiHeaders({ contentType: 'application/json' }),
			body: JSON.stringify({
				name: values.name,
				description: values.description
			})
		});

		const result = await readApiResponse<CategoryMutationApiData>(response);
		const apiError = getApiErrorMessage(response, result, 'Impossible de modifier la categorie');
		if (apiError) {
			return failAdminAction(response.status || 500, 'edit-category', apiError, {
				categoryValues: values
			});
		}

		return {
			action: 'edit-category',
			success: 'Categorie modifiee avec succes'
		};
	},

	deleteCategory: async ({ request, fetch, locals }) => {
		requireAdmin(locals.user);

		const categoryId = getFormString(await request.formData(), 'id').trim();
		if (!categoryId) {
			return failAdminAction(400, 'delete-category', 'Categorie introuvable', {
				categoryId
			});
		}

		const response = await fetch(buildInternalApiPath(`/categories/${categoryId}`), {
			method: 'DELETE'
		});

		const result = await readApiResponse<CategoryMutationApiData>(response);
		const apiError = getApiErrorMessage(response, result, 'Impossible de supprimer la categorie');
		if (apiError) {
			return failAdminAction(response.status || 500, 'delete-category', apiError, {
				categoryId
			});
		}

		return {
			action: 'delete-category',
			success: 'Categorie supprimee avec succes'
		};
	},

	createPromotion: async ({ request, fetch, locals }) => {
		requireAdmin(locals.user);

		const { values, value, productIds } = await readPromotionForm(request);
		const validationMessage = validatePromotionFormError(values, value, productIds);
		if (validationMessage) {
			return failAdminAction(400, 'create-promotion', validationMessage, {
				promotionValues: values
			});
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
				applies_to_all: values.appliesToAll === 'true',
				product_ids: values.appliesToAll === 'true' ? [] : productIds
			})
		});

		const result = await readApiResponse<PromotionMutationApiData>(response);
		const apiError = getApiErrorMessage(response, result, 'Impossible de creer la promotion');
		if (apiError) {
			return failAdminAction(response.status || 500, 'create-promotion', apiError, {
				promotionValues: values
			});
		}

		return {
			action: 'create-promotion',
			success: 'Promotion ajoutee avec succes'
		};
	},

	updatePromotion: async ({ request, fetch, locals }) => {
		requireAdmin(locals.user);

		const { id, values, value, productIds } = await readPromotionForm(request);
		if (!id) {
			return failAdminAction(400, 'edit-promotion', 'Promotion introuvable', {
				promotionValues: values
			});
		}

		const validationMessage = validatePromotionFormError(values, value, productIds);
		if (validationMessage) {
			return failAdminAction(400, 'edit-promotion', validationMessage, {
				promotionValues: values
			});
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
				applies_to_all: values.appliesToAll === 'true',
				product_ids: values.appliesToAll === 'true' ? [] : productIds
			})
		});

		const result = await readApiResponse<PromotionMutationApiData>(response);
		const apiError = getApiErrorMessage(response, result, 'Impossible de modifier la promotion');
		if (apiError) {
			return failAdminAction(response.status || 500, 'edit-promotion', apiError, {
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
		requireAdmin(locals.user);

		const promotionId = getFormString(await request.formData(), 'id').trim();
		if (!promotionId) {
			return failAdminAction(400, 'delete-promotion', 'Promotion introuvable', {
				promotionId
			});
		}

		const response = await fetch(buildInternalApiPath(`/promotions/${promotionId}`), {
			method: 'DELETE'
		});

		const result = await readApiResponse<PromotionMutationApiData>(response);
		const apiError = getApiErrorMessage(response, result, 'Impossible de supprimer la promotion');
		if (apiError) {
			return failAdminAction(response.status || 500, 'delete-promotion', apiError, {
				promotionId
			});
		}

		return {
			action: 'delete-promotion',
			success: 'Promotion supprimee avec succes'
		};
	}
};
