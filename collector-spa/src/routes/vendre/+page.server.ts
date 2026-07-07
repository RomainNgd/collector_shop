import {
	buildApiHeaders,
	buildInternalApiPath,
	getApiErrorMessage,
	readApiResponse
} from '$lib/server/api';
import { loadCategories } from '$lib/server/categories';
import { getFormString } from '$lib/server/forms';
import {
	PROMOTION_TYPE_FIXED,
	PROMOTION_TYPE_PERCENTAGE,
	ADMIN_ROLE,
	type ApiProduct
} from '$lib/types';
import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';

type SellFormValues = {
	name: string;
	description: string;
	price: string;
	stock: string;
	categoryId: string;
	promotionActive: 'true' | 'false';
	promotionType: string;
	promotionValue: string;
};

const getImageFile = (entry: FormDataEntryValue | null): File | null =>
	entry instanceof File && entry.size > 0 ? entry : null;

const extractProductId = (item: ApiProduct | null | undefined) =>
	typeof item?.ID === 'number' && Number.isFinite(item.ID) ? item.ID : null;

const failSell = (status: number, message: string, values: SellFormValues) =>
	fail(status, { error: message, values });

const readSellForm = async (request: Request) => {
	const formData = await request.formData();
	const promotionActive =
		getFormString(formData, 'promotion_active', 'false') === 'true' ? 'true' : 'false';

	return {
		values: {
			name: getFormString(formData, 'name').trim(),
			description: getFormString(formData, 'description').trim(),
			price: getFormString(formData, 'price').trim(),
			stock: getFormString(formData, 'stock').trim(),
			categoryId: getFormString(formData, 'category_id').trim(),
			promotionActive,
			promotionType: getFormString(formData, 'promotion_type').trim(),
			promotionValue: getFormString(formData, 'promotion_value').trim()
		} satisfies SellFormValues,
		imageFile: getImageFile(formData.get('image'))
	};
};

const validateSellForm = (values: SellFormValues, imageFile: File | null) => {
	const price = Number(values.price);
	const stock = Number(values.stock);
	const categoryId = Number(values.categoryId);
	const promotionValue = Number(values.promotionValue);

	if (!values.name || !values.description || values.price === '' || values.stock === '') {
		return 'Tous les champs produit sont requis';
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
	if (imageFile && !imageFile.type.startsWith('image/')) {
		return 'Le fichier choisi doit etre une image';
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

const uploadProductImage = async (fetchFn: typeof fetch, productId: number, imageFile: File) => {
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

export const load: PageServerLoad = async ({ locals, fetch }) => {
	if (!locals.user) {
		redirect(303, '/login');
	}
	if (locals.user.role === ADMIN_ROLE) {
		redirect(303, '/administration');
	}

	return {
		categories: await loadCategories(fetch)
	};
};

export const actions: Actions = {
	default: async ({ request, fetch, locals }) => {
		if (!locals.user) {
			redirect(303, '/login');
		}
		if (locals.user.role === ADMIN_ROLE) {
			redirect(303, '/administration');
		}

		const { values, imageFile } = await readSellForm(request);
		const validationError = validateSellForm(values, imageFile);
		if (validationError) {
			return failSell(400, validationError, values);
		}

		const promotionActive = values.promotionActive === 'true';
		const response = await fetch(buildInternalApiPath('/products'), {
			method: 'POST',
			headers: buildApiHeaders({ contentType: 'application/json' }),
			body: JSON.stringify({
				name: values.name,
				description: values.description,
				price: Number(values.price),
				stock: Number(values.stock),
				category_id: Number(values.categoryId),
				promotion_active: promotionActive,
				promotion_type: promotionActive ? values.promotionType : '',
				promotion_value: promotionActive ? Number(values.promotionValue) : 0
			})
		});

		const result = await readApiResponse<ApiProduct>(response);
		const apiError = getApiErrorMessage(response, result, 'Impossible de creer le produit');
		if (apiError) {
			return failSell(response.status || 500, apiError, values);
		}

		const productId = extractProductId(result.payload?.data);
		if (imageFile && productId !== null) {
			const imageError = await uploadProductImage(fetch, productId, imageFile);
			if (imageError) {
				return failSell(500, `Produit cree, mais image non envoyee: ${imageError}`, values);
			}
		}

		redirect(303, '/mes-produits');
	}
};
