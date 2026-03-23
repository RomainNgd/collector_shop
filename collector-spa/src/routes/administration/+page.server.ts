import {
	ADMIN_ROLE,
	mapApiCategory,
	mapApiProduct,
	type ApiCategory,
	type ApiProduct,
	type AuthUser
} from '$lib/types';
import {
	API_BASE_URL,
	API_PUBLIC_BASE_URL,
	buildApiHeaders,
	getApiErrorMessage,
	readApiResponse
} from '$lib/server/api';
import { error, fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';

type AdminAction =
	| 'create-product'
	| 'edit-product'
	| 'delete-product'
	| 'create-category'
	| 'edit-category'
	| 'delete-category';

type ProductMutationApiData = ApiProduct | { ID?: number; id?: number } | null;
type CategoryMutationApiData = ApiCategory | { ID?: number; id?: number } | null;

type ProductFormValues = {
	id?: string;
	name: string;
	description: string;
	price: string;
	categoryId: string;
	currentImageName?: string;
	removeImage?: 'true';
};

type CategoryFormValues = {
	id?: string;
	name: string;
	description: string;
};

type ParsedProductForm = {
	id: string;
	values: ProductFormValues;
	price: number;
	imageFile: File | null;
	removeImage: boolean;
};

const requireAdmin = (user: AuthUser | null) => {
	if (!user) {
		redirect(303, '/login');
	}

	if (user.role !== ADMIN_ROLE) {
		throw error(403, 'Acces refuse');
	}
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
	}
) =>
	fail(status, {
		action,
		error: message,
		values: options?.productValues,
		productId: options?.productId,
		categoryValues: options?.categoryValues,
		categoryId: options?.categoryId
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

const extractEntityId = (payload: ProductMutationApiData | CategoryMutationApiData | undefined): number | null => {
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
	const id = String(formData.get('id') ?? '').trim();
	const name = String(formData.get('name') ?? '').trim();
	const description = String(formData.get('description') ?? '').trim();
	const priceValue = String(formData.get('price') ?? '').trim();
	const categoryId = String(formData.get('category_id') ?? '').trim();
	const currentImageName = String(formData.get('currentImageName') ?? '').trim();
	const imageFile = getImageFile(formData.get('image'));
	const removeImage = String(formData.get('removeImage') ?? '') === 'true';

	return {
		id,
		values: {
			id,
			name,
			description,
			price: priceValue,
			categoryId,
			currentImageName,
			removeImage: removeImage ? 'true' : undefined
		},
		price: Number(priceValue),
		imageFile,
		removeImage
	};
};

const readCategoryForm = async (request: Request) => {
	const formData = await request.formData();
	const id = String(formData.get('id') ?? '').trim();
	const name = String(formData.get('name') ?? '').trim();
	const description = String(formData.get('description') ?? '').trim();

	return {
		id,
		values: {
			id,
			name,
			description
		} satisfies CategoryFormValues
	};
};

const validateProductForm = (
	values: ProductFormValues,
	action: Extract<AdminAction, 'create-product' | 'edit-product'>,
	price: number,
	imageFile: File | null
) => {
	if (!values.name || !values.description || values.price === '') {
		return failAdminAction(400, action, 'Tous les champs produit sont requis', {
			productValues: values
		});
	}

	if (values.categoryId === '') {
		return failAdminAction(400, action, 'Tous les champs produit sont requis', {
			productValues: values
		});
	}

	if (!Number.isFinite(price) || price < 0) {
		return failAdminAction(400, action, 'Le prix doit etre un nombre valide', {
			productValues: values
		});
	}

	const numericCategoryId = Number(values.categoryId);
	if (!Number.isInteger(numericCategoryId) || numericCategoryId <= 0) {
		return failAdminAction(400, action, 'La categorie selectionnee est invalide', {
			productValues: values
		});
	}

	const imageError = validateImageFile(imageFile);
	if (imageError) {
		return failAdminAction(400, action, imageError, {
			productValues: values
		});
	}

	return null;
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

const uploadProductImage = async (
	fetchFn: typeof fetch,
	productId: string | number,
	token: string | undefined,
	imageFile: File
) => {
	const uploadBlob = new Blob([await imageFile.arrayBuffer()], {
		type: imageFile.type || 'application/octet-stream'
	});
	const imageFormData = new FormData();
	imageFormData.set('image', uploadBlob, imageFile.name || 'upload-image');

	const response = await fetchFn(`${API_BASE_URL}/products/${productId}/image`, {
		method: 'POST',
		headers: buildApiHeaders({ token }),
		body: imageFormData
	});

	const result = await readApiResponse<unknown>(response);
	return getApiErrorMessage(response, result, "Impossible d'envoyer l'image");
};

const deleteProductImage = async (
	fetchFn: typeof fetch,
	productId: string | number,
	token: string | undefined
) => {
	const response = await fetchFn(`${API_BASE_URL}/products/${productId}/image`, {
		method: 'DELETE',
		headers: buildApiHeaders({ token })
	});

	const result = await readApiResponse<unknown>(response);
	return getApiErrorMessage(response, result, "Impossible de supprimer l'image");
};

const loadProducts = async (fetchFn: typeof fetch) => {
	const response = await fetchFn(`${API_BASE_URL}/products`);

	if (!response.ok) {
		throw error(response.status, 'Impossible de charger les produits');
	}

	const result = await readApiResponse<ApiProduct[]>(response);

	if (!Array.isArray(result.payload?.data)) {
		throw error(502, 'Format de reponse API invalide pour les produits');
	}

	return result.payload.data.map((item) => mapApiProduct(item, API_PUBLIC_BASE_URL));
};

const loadCategories = async (fetchFn: typeof fetch) => {
	const response = await fetchFn(`${API_BASE_URL}/categories`);

	if (!response.ok) {
		throw error(response.status, 'Impossible de charger les categories');
	}

	const result = await readApiResponse<ApiCategory[]>(response);

	if (!Array.isArray(result.payload?.data)) {
		throw error(502, 'Format de reponse API invalide pour les categories');
	}

	return result.payload.data.map(mapApiCategory);
};

export const load: PageServerLoad = async ({ locals, fetch }) => {
	requireAdmin(locals.user);

	const [products, categories] = await Promise.all([loadProducts(fetch), loadCategories(fetch)]);

	return {
		products,
		categories
	};
};

export const actions: Actions = {
	createProduct: async ({ request, fetch, cookies, locals }) => {
		requireAdmin(locals.user);

		const token = cookies.get('auth_token');
		const { values, price, imageFile } = await readProductForm(request);
		const validationError = validateProductForm(values, 'create-product', price, imageFile);
		if (validationError) {
			return validationError;
		}

		const response = await fetch(`${API_BASE_URL}/products`, {
			method: 'POST',
			headers: buildApiHeaders({ token, contentType: 'application/json' }),
			body: JSON.stringify({
				name: values.name,
				price,
				description: values.description,
				image: 'truc.jpg',
				category_id: Number(values.categoryId)
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

			const imageError = await uploadProductImage(fetch, createdProductId, token, imageFile);
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

	updateProduct: async ({ request, fetch, cookies, locals }) => {
		requireAdmin(locals.user);

		const token = cookies.get('auth_token');
		const { id, values, price, imageFile, removeImage } = await readProductForm(request);

		if (!id) {
			return failAdminAction(400, 'edit-product', 'Produit introuvable', {
				productValues: values
			});
		}

		const validationError = validateProductForm(values, 'edit-product', price, imageFile);
		if (validationError) {
			return validationError;
		}

		const response = await fetch(`${API_BASE_URL}/products/${id}`, {
			method: 'PUT',
			headers: buildApiHeaders({ token, contentType: 'application/json' }),
			body: JSON.stringify({
				name: values.name,
				price,
				description: values.description,
				image: removeImage && !imageFile ? '' : (values.currentImageName ?? ''),
				category_id: Number(values.categoryId)
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
			const imageDeleteError = await deleteProductImage(fetch, id, token);
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
			const imageUploadError = await uploadProductImage(fetch, id, token, imageFile);
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

	deleteProduct: async ({ request, fetch, cookies, locals }) => {
		requireAdmin(locals.user);

		const productId = String((await request.formData()).get('id') ?? '').trim();

		if (!productId) {
			return failAdminAction(400, 'delete-product', 'Produit introuvable', {
				productId
			});
		}

		const response = await fetch(`${API_BASE_URL}/products/${productId}`, {
			method: 'DELETE',
			headers: buildApiHeaders({ token: cookies.get('auth_token') })
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

	createCategory: async ({ request, fetch, cookies, locals }) => {
		requireAdmin(locals.user);

		const token = cookies.get('auth_token');
		const { values } = await readCategoryForm(request);
		const validationError = validateCategoryForm(values, 'create-category');
		if (validationError) {
			return validationError;
		}

		const response = await fetch(`${API_BASE_URL}/categories`, {
			method: 'POST',
			headers: buildApiHeaders({ token, contentType: 'application/json' }),
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

	updateCategory: async ({ request, fetch, cookies, locals }) => {
		requireAdmin(locals.user);

		const token = cookies.get('auth_token');
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

		const response = await fetch(`${API_BASE_URL}/categories/${id}`, {
			method: 'PUT',
			headers: buildApiHeaders({ token, contentType: 'application/json' }),
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

	deleteCategory: async ({ request, fetch, cookies, locals }) => {
		requireAdmin(locals.user);

		const categoryId = String((await request.formData()).get('id') ?? '').trim();
		if (!categoryId) {
			return failAdminAction(400, 'delete-category', 'Categorie introuvable', {
				categoryId
			});
		}

		const response = await fetch(`${API_BASE_URL}/categories/${categoryId}`, {
			method: 'DELETE',
			headers: buildApiHeaders({ token: cookies.get('auth_token') })
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
