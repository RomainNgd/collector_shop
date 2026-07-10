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
	createProductActions,
	type ProductAction,
	type ProductFormValues
} from '$lib/server/productMutations';
import {
	createPromotionActions,
	type PromotionAction,
	type PromotionFormValues
} from '$lib/server/promotionMutations';
import { fail } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';

type CategoryAction = 'create-category' | 'edit-category' | 'delete-category';
type AdminAction = ProductAction | CategoryAction | PromotionAction;

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
	...createProductActions(requireAdmin, failAdminAction),
	...createPromotionActions(requireAdmin, failAdminAction, { allowAppliesToAll: true }),

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
	}
};
