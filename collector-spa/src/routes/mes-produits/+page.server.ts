import {
	buildApiHeaders,
	buildInternalApiPath,
	getApiErrorMessage,
	readApiResponse
} from '$lib/server/api';
import { loadCategories } from '$lib/server/categories';
import { getFormString } from '$lib/server/forms';
import { loadSellerProducts } from '$lib/server/products';
import {
	PROMOTION_TYPE_FIXED,
	PROMOTION_TYPE_PERCENTAGE,
	ADMIN_ROLE,
	type ApiProduct
} from '$lib/types';
import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';

type ProductFormValues = {
	id: string;
	name: string;
	description: string;
	price: string;
	stock: string;
	categoryId: string;
	isActive: 'true' | 'false';
	promotionActive: 'true' | 'false';
	promotionType: string;
	promotionValue: string;
	imageName: string;
};

const failProductAction = (status: number, message: string, values?: ProductFormValues) =>
	fail(status, { error: message, values });

const readProductForm = async (request: Request): Promise<ProductFormValues> => {
	const formData = await request.formData();
	return {
		id: getFormString(formData, 'id').trim(),
		name: getFormString(formData, 'name').trim(),
		description: getFormString(formData, 'description').trim(),
		price: getFormString(formData, 'price').trim(),
		stock: getFormString(formData, 'stock').trim(),
		categoryId: getFormString(formData, 'category_id').trim(),
		isActive: getFormString(formData, 'is_active', 'false') === 'true' ? 'true' : 'false',
		promotionActive:
			getFormString(formData, 'promotion_active', 'false') === 'true' ? 'true' : 'false',
		promotionType: getFormString(formData, 'promotion_type').trim(),
		promotionValue: getFormString(formData, 'promotion_value').trim(),
		imageName: getFormString(formData, 'image').trim()
	};
};

const validateProductForm = (values: ProductFormValues) => {
	const price = Number(values.price);
	const stock = Number(values.stock);
	const categoryId = Number(values.categoryId);
	const promotionValue = Number(values.promotionValue);

	if (!values.id || !values.name || !values.description) {
		return 'Produit introuvable ou incomplet';
	}
	if (!Number.isFinite(price) || price <= 0) {
		return 'Le prix doit etre positif';
	}
	if (!Number.isInteger(stock) || stock <= 0) {
		return 'Le stock doit etre un entier positif';
	}
	if (!Number.isInteger(categoryId) || categoryId <= 0) {
		return 'Selectionne une categorie existante';
	}
	if (values.promotionActive === 'true') {
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

export const load: PageServerLoad = async ({ locals, fetch }) => {
	if (!locals.user) {
		redirect(303, '/login');
	}
	if (locals.user.role === ADMIN_ROLE) {
		redirect(303, '/administration');
	}

	const [products, categories] = await Promise.all([
		loadSellerProducts(fetch),
		loadCategories(fetch)
	]);

	return { products, categories };
};

export const actions: Actions = {
	updateProduct: async ({ request, fetch, locals }) => {
		if (!locals.user) {
			redirect(303, '/login');
		}
		if (locals.user.role === ADMIN_ROLE) {
			redirect(303, '/administration');
		}

		const values = await readProductForm(request);
		const validationError = validateProductForm(values);
		if (validationError) {
			return failProductAction(400, validationError, values);
		}

		const promotionActive = values.promotionActive === 'true';
		const response = await fetch(buildInternalApiPath(`/products/${values.id}`), {
			method: 'PUT',
			headers: buildApiHeaders({ contentType: 'application/json' }),
			body: JSON.stringify({
				name: values.name,
				description: values.description,
				image: values.imageName,
				price: Number(values.price),
				stock: Number(values.stock),
				is_active: values.isActive === 'true',
				category_id: Number(values.categoryId),
				promotion_active: promotionActive,
				promotion_type: promotionActive ? values.promotionType : '',
				promotion_value: promotionActive ? Number(values.promotionValue) : 0
			})
		});

		const result = await readApiResponse<ApiProduct>(response);
		const apiError = getApiErrorMessage(response, result, 'Impossible de modifier le produit');
		if (apiError) {
			return failProductAction(response.status || 500, apiError, values);
		}

		return { success: 'Produit mis a jour' };
	},

	deleteProduct: async ({ request, fetch, locals }) => {
		if (!locals.user) {
			redirect(303, '/login');
		}
		if (locals.user.role === ADMIN_ROLE) {
			redirect(303, '/administration');
		}

		const productId = getFormString(await request.formData(), 'id').trim();
		if (!productId) {
			return failProductAction(400, 'Produit introuvable');
		}

		const response = await fetch(buildInternalApiPath(`/products/${productId}`), {
			method: 'DELETE'
		});
		const result = await readApiResponse<unknown>(response);
		const apiError = getApiErrorMessage(response, result, 'Impossible de supprimer le produit');
		if (apiError) {
			return failProductAction(response.status || 500, apiError);
		}

		return { success: 'Produit supprime' };
	}
};
