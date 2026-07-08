import { describe, expect, it, vi } from 'vitest';

import { ADMIN_ROLE, PROMOTION_TYPE_PERCENTAGE, USER_ROLE } from '$lib/types';
import {
	actions as sellerProductActions,
	load as loadSellerProductsPage
} from './mes-produits/+page.server';
import { actions as sellActions, load as loadSellPage } from './vendre/+page.server';

const user = { id: 1, role: USER_ROLE };
const apiResponse = (data: unknown, status = 200) =>
	new Response(JSON.stringify({ success: status < 400, data }), { status });
const requestWithForm = (path: string, values: Record<string, string>) =>
	new Request(`http://localhost${path}`, {
		method: 'POST',
		body: new URLSearchParams(values)
	});
const requestWithMultipart = (path: string, values: Record<string, string>, file: File) => {
	const formData = new FormData();
	for (const [key, value] of Object.entries(values)) {
		formData.set(key, value);
	}
	formData.set('image', file);
	return new Request(`http://localhost${path}`, { method: 'POST', body: formData });
};
const fetchByPath = (responses: Record<string, Response>) =>
	vi.fn((input: RequestInfo | URL) => {
		const url = typeof input === 'string' ? input : input.toString();
		const match = Object.entries(responses).find(([path]) => url.endsWith(path));
		return Promise.resolve(match ? match[1] : apiResponse(null, 404));
	}) as unknown as typeof fetch;

const productForm = {
	id: '4',
	name: 'Console',
	description: 'Console retro',
	price: '99.90',
	stock: '2',
	category_id: '3',
	is_active: 'true',
	promotion_active: 'true',
	promotion_type: PROMOTION_TYPE_PERCENTAGE,
	promotion_value: '10',
	image: ''
};

const promotionForm = {
	id: '9',
	name: 'Promo Test',
	description: 'Remise test',
	type: PROMOTION_TYPE_PERCENTAGE,
	value: '15',
	is_active: 'true',
	product_ids: '4'
};

describe('seller pages', () => {
	it('protects seller pages from anonymous and admin users', async () => {
		await expect(
			loadSellPage({ locals: { user: null }, fetch: vi.fn() } as never)
		).rejects.toMatchObject({ status: 303, location: '/login' });
		await expect(
			loadSellPage({
				locals: { user: { id: 2, role: ADMIN_ROLE } },
				fetch: vi.fn()
			} as never)
		).rejects.toMatchObject({ status: 303, location: '/administration' });
		await expect(
			loadSellerProductsPage({ locals: { user: null }, fetch: vi.fn() } as never)
		).rejects.toMatchObject({ status: 303, location: '/login' });
		await expect(
			loadSellerProductsPage({
				locals: { user: { id: 2, role: ADMIN_ROLE } },
				fetch: vi.fn()
			} as never)
		).rejects.toMatchObject({ status: 303, location: '/administration' });
	});

	it('protects seller product mutation actions from anonymous and admin users', async () => {
		const update = sellerProductActions.updateProduct!;
		await expect(
			update({
				locals: { user: null },
				request: requestWithForm('/mes-produits', {}),
				fetch: vi.fn()
			} as never)
		).rejects.toMatchObject({ status: 303, location: '/login' });
		await expect(
			update({
				locals: { user: { id: 2, role: ADMIN_ROLE } },
				request: requestWithForm('/mes-produits', {}),
				fetch: vi.fn()
			} as never)
		).rejects.toMatchObject({ status: 303, location: '/administration' });

		const deleteAction = sellerProductActions.deleteProduct!;
		await expect(
			deleteAction({
				locals: { user: { id: 2, role: ADMIN_ROLE } },
				request: requestWithForm('/mes-produits', { id: '4' }),
				fetch: vi.fn()
			} as never)
		).rejects.toMatchObject({ status: 303, location: '/administration' });

		const action = sellActions.default!;
		await expect(
			action({
				locals: { user: null },
				request: requestWithForm('/vendre', {}),
				fetch: vi.fn()
			} as never)
		).rejects.toMatchObject({ status: 303, location: '/login' });
		await expect(
			action({
				locals: { user: { id: 2, role: ADMIN_ROLE } },
				request: requestWithForm('/vendre', {}),
				fetch: vi.fn()
			} as never)
		).rejects.toMatchObject({ status: 303, location: '/administration' });
	});

	it('validates seller product edge cases for price, category and promotion fields', async () => {
		const update = sellerProductActions.updateProduct!;
		const invalidPrice = { ...productForm, price: '-5' };
		await expect(
			update({
				locals: { user },
				request: requestWithForm('/mes-produits', invalidPrice),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });

		const invalidCategory = { ...productForm, category_id: '0' };
		await expect(
			update({
				locals: { user },
				request: requestWithForm('/mes-produits', invalidCategory),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });

		const invalidPromotionType = {
			...productForm,
			promotion_active: 'true',
			promotion_type: 'unknown'
		};
		await expect(
			update({
				locals: { user },
				request: requestWithForm('/mes-produits', invalidPromotionType),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });

		const invalidPromotionValue = {
			...productForm,
			promotion_active: 'true',
			promotion_value: '0'
		};
		await expect(
			update({
				locals: { user },
				request: requestWithForm('/mes-produits', invalidPromotionValue),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });

		const promotionOverHundred = {
			...productForm,
			promotion_active: 'true',
			promotion_type: PROMOTION_TYPE_PERCENTAGE,
			promotion_value: '150'
		};
		await expect(
			update({
				locals: { user },
				request: requestWithForm('/mes-produits', promotionOverHundred),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });
	});

	it('validates sell form edge cases for price, category and promotion fields', async () => {
		const action = sellActions.default!;

		await expect(
			action({
				locals: { user },
				request: requestWithForm('/vendre', { ...productForm, price: '-5' }),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });

		await expect(
			action({
				locals: { user },
				request: requestWithForm('/vendre', { ...productForm, category_id: '0' }),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });

		await expect(
			action({
				locals: { user },
				request: requestWithForm('/vendre', {
					...productForm,
					promotion_active: 'true',
					promotion_type: 'unknown'
				}),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });

		await expect(
			action({
				locals: { user },
				request: requestWithForm('/vendre', {
					...productForm,
					promotion_active: 'true',
					promotion_value: '0'
				}),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });

		await expect(
			action({
				locals: { user },
				request: requestWithForm('/vendre', {
					...productForm,
					promotion_active: 'true',
					promotion_type: PROMOTION_TYPE_PERCENTAGE,
					promotion_value: '150'
				}),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });
	});

	it('loads sell and seller products pages for authenticated sellers', async () => {
		await expect(
			loadSellPage({
				locals: { user },
				fetch: fetchByPath({ '/categories': apiResponse([{ ID: 3, name: 'Consoles' }]) })
			} as never)
		).resolves.toMatchObject({ categories: [{ id: 3 }] });

		await expect(
			loadSellerProductsPage({
				locals: { user },
				fetch: fetchByPath({
					'/seller/products': apiResponse([
						{ ID: 4, name: 'Console', description: 'Retro', price: 10, image: '' }
					]),
					'/categories': apiResponse([{ ID: 3, name: 'Consoles' }]),
					'/promotions': apiResponse([]),
					'/seller/stats': apiResponse({ total_revenue: 0, total_sales: 0, product_count: 1 })
				})
			} as never)
		).resolves.toMatchObject({
			products: [{ id: 4 }],
			categories: [{ id: 3 }],
			promotions: [],
			stats: { productCount: 1 }
		});
	});

	it('validates and creates a seller product', async () => {
		const action = sellActions.default!;
		await expect(
			action({
				locals: { user },
				request: requestWithForm('/vendre', { name: '' }),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });

		const fetchMock = vi.fn(() =>
			Promise.resolve(apiResponse({ ID: 12 }, 201))
		) as unknown as typeof globalThis.fetch;
		await expect(
			action({
				locals: { user },
				request: requestWithForm('/vendre', productForm),
				fetch: fetchMock
			} as never)
		).rejects.toMatchObject({ status: 303, location: '/mes-produits' });
		expect(fetchMock).toHaveBeenCalledOnce();
	});

	it('handles seller product API and image upload errors', async () => {
		const action = sellActions.default!;
		await expect(
			action({
				locals: { user },
				request: requestWithForm('/vendre', { ...productForm, price: '-1' }),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });

		await expect(
			action({
				locals: { user },
				request: requestWithForm('/vendre', productForm),
				fetch: vi.fn(() =>
					Promise.resolve(
						new Response(JSON.stringify({ success: false, error: { message: 'API down' } }), {
							status: 503
						})
					)
				)
			} as never)
		).resolves.toMatchObject({ status: 503 });

		const image = new File(['fake'], 'console.png', { type: 'image/png' });
		const uploadFails = vi
			.fn()
			.mockResolvedValueOnce(apiResponse({ ID: 12 }, 201))
			.mockResolvedValueOnce(
				new Response(JSON.stringify({ success: false, error: { message: 'Upload down' } }), {
					status: 500
				})
			) as unknown as typeof globalThis.fetch;

		await expect(
			action({
				locals: { user },
				request: requestWithMultipart('/vendre', productForm, image),
				fetch: uploadFails
			} as never)
		).resolves.toMatchObject({ status: 500 });
	});

	it('updates and deletes seller products', async () => {
		const update = sellerProductActions.updateProduct!;
		await expect(
			update({
				locals: { user },
				request: requestWithForm('/mes-produits', { ...productForm, stock: '0' }),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });

		await expect(
			update({
				locals: { user },
				request: requestWithForm('/mes-produits', productForm),
				fetch: vi.fn(() => Promise.resolve(apiResponse({ ID: 4 })))
			} as never)
		).resolves.toMatchObject({ success: 'Produit modifie avec succes' });

		await expect(
			update({
				locals: { user },
				request: requestWithForm('/mes-produits', productForm),
				fetch: vi.fn(() =>
					Promise.resolve(
						new Response(JSON.stringify({ success: false, error: { message: 'Update down' } }), {
							status: 503
						})
					)
				)
			} as never)
		).resolves.toMatchObject({ status: 503 });

		const deleteAction = sellerProductActions.deleteProduct!;
		await expect(
			deleteAction({
				locals: { user: null },
				request: requestWithForm('/mes-produits', { id: '4' }),
				fetch: vi.fn()
			} as never)
		).rejects.toMatchObject({ status: 303, location: '/login' });

		await expect(
			deleteAction({
				locals: { user },
				request: requestWithForm('/mes-produits', { id: '' }),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });

		await expect(
			deleteAction({
				locals: { user },
				request: requestWithForm('/mes-produits', { id: '4' }),
				fetch: vi.fn(() => Promise.resolve(apiResponse({ message: 'ok' })))
			} as never)
		).resolves.toMatchObject({ success: 'Produit supprime avec succes' });

		await expect(
			deleteAction({
				locals: { user },
				request: requestWithForm('/mes-produits', { id: '4' }),
				fetch: vi.fn(() =>
					Promise.resolve(
						new Response(JSON.stringify({ success: false, error: { message: 'Delete down' } }), {
							status: 500
						})
					)
				)
			} as never)
		).resolves.toMatchObject({ status: 500 });
	});

	it('removes the current image when updating a seller product', async () => {
		const update = sellerProductActions.updateProduct!;

		const removeOk = vi
			.fn()
			.mockResolvedValueOnce(apiResponse({ ID: 4 }))
			.mockResolvedValueOnce(apiResponse({ message: 'ok' })) as unknown as typeof globalThis.fetch;
		await expect(
			update({
				locals: { user },
				request: requestWithForm('/mes-produits', {
					...productForm,
					removeImage: 'true',
					currentImageName: 'console.png'
				}),
				fetch: removeOk
			} as never)
		).resolves.toMatchObject({ success: 'Produit et image mis a jour avec succes' });

		const removeFails = vi
			.fn()
			.mockResolvedValueOnce(apiResponse({ ID: 4 }))
			.mockResolvedValueOnce(
				new Response(JSON.stringify({ success: false, error: { message: 'Delete image down' } }), {
					status: 500
				})
			) as unknown as typeof globalThis.fetch;
		await expect(
			update({
				locals: { user },
				request: requestWithForm('/mes-produits', {
					...productForm,
					removeImage: 'true',
					currentImageName: 'console.png'
				}),
				fetch: removeFails
			} as never)
		).resolves.toMatchObject({ status: 500 });

		const replaceImage = new File(['fake'], 'new-console.png', { type: 'image/png' });
		const replaceOk = vi
			.fn()
			.mockResolvedValueOnce(apiResponse({ ID: 4 }))
			.mockResolvedValueOnce(apiResponse({ ID: 4 })) as unknown as typeof globalThis.fetch;
		await expect(
			update({
				locals: { user },
				request: requestWithMultipart('/mes-produits', productForm, replaceImage),
				fetch: replaceOk
			} as never)
		).resolves.toMatchObject({ success: 'Produit et image mis a jour avec succes' });

		const replaceFails = vi
			.fn()
			.mockResolvedValueOnce(apiResponse({ ID: 4 }))
			.mockResolvedValueOnce(
				new Response(JSON.stringify({ success: false, error: { message: 'Upload down' } }), {
					status: 500
				})
			) as unknown as typeof globalThis.fetch;
		await expect(
			update({
				locals: { user },
				request: requestWithMultipart('/mes-produits', productForm, replaceImage),
				fetch: replaceFails
			} as never)
		).resolves.toMatchObject({ status: 500 });
	});

	it('protects seller create-product and promotion mutation actions', async () => {
		const create = sellerProductActions.createProduct!;
		await expect(
			create({
				locals: { user: null },
				request: requestWithForm('/mes-produits', {}),
				fetch: vi.fn()
			} as never)
		).rejects.toMatchObject({ status: 303, location: '/login' });
		await expect(
			create({
				locals: { user: { id: 2, role: ADMIN_ROLE } },
				request: requestWithForm('/mes-produits', {}),
				fetch: vi.fn()
			} as never)
		).rejects.toMatchObject({ status: 303, location: '/administration' });

		const createPromotion = sellerProductActions.createPromotion!;
		await expect(
			createPromotion({
				locals: { user: null },
				request: requestWithForm('/mes-produits', {}),
				fetch: vi.fn()
			} as never)
		).rejects.toMatchObject({ status: 303, location: '/login' });
		await expect(
			createPromotion({
				locals: { user: { id: 2, role: ADMIN_ROLE } },
				request: requestWithForm('/mes-produits', {}),
				fetch: vi.fn()
			} as never)
		).rejects.toMatchObject({ status: 303, location: '/administration' });

		const updatePromotion = sellerProductActions.updatePromotion!;
		await expect(
			updatePromotion({
				locals: { user: { id: 2, role: ADMIN_ROLE } },
				request: requestWithForm('/mes-produits', { id: '9' }),
				fetch: vi.fn()
			} as never)
		).rejects.toMatchObject({ status: 303, location: '/administration' });

		const deletePromotion = sellerProductActions.deletePromotion!;
		await expect(
			deletePromotion({
				locals: { user: { id: 2, role: ADMIN_ROLE } },
				request: requestWithForm('/mes-produits', { id: '9' }),
				fetch: vi.fn()
			} as never)
		).rejects.toMatchObject({ status: 303, location: '/administration' });
	});

	it('validates and creates a seller product from the dashboard', async () => {
		const create = sellerProductActions.createProduct!;

		await expect(
			create({
				locals: { user },
				request: requestWithForm('/mes-produits', { ...productForm, name: '' }),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });

		await expect(
			create({
				locals: { user },
				request: requestWithForm('/mes-produits', { ...productForm, category_id: '0' }),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });

		await expect(
			create({
				locals: { user },
				request: requestWithForm('/mes-produits', {
					...productForm,
					promotion_active: 'true',
					promotion_type: 'unknown'
				}),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });

		const fetchMock = vi.fn(() =>
			Promise.resolve(apiResponse({ ID: 12 }, 201))
		) as unknown as typeof globalThis.fetch;
		await expect(
			create({
				locals: { user },
				request: requestWithForm('/mes-produits', productForm),
				fetch: fetchMock
			} as never)
		).resolves.toMatchObject({ success: 'Produit ajoute avec succes' });
		expect(fetchMock).toHaveBeenCalledOnce();

		await expect(
			create({
				locals: { user },
				request: requestWithForm('/mes-produits', productForm),
				fetch: vi.fn(() =>
					Promise.resolve(
						new Response(JSON.stringify({ success: false, error: { message: 'API down' } }), {
							status: 503
						})
					)
				)
			} as never)
		).resolves.toMatchObject({ status: 503 });
	});

	it('creates a seller product with an image and handles upload failure', async () => {
		const create = sellerProductActions.createProduct!;
		const image = new File(['fake'], 'console.png', { type: 'image/png' });

		const uploadOk = vi
			.fn()
			.mockResolvedValueOnce(apiResponse({ ID: 12 }, 201))
			.mockResolvedValueOnce(apiResponse({ ID: 12 })) as unknown as typeof globalThis.fetch;
		await expect(
			create({
				locals: { user },
				request: requestWithMultipart('/mes-produits', productForm, image),
				fetch: uploadOk
			} as never)
		).resolves.toMatchObject({ success: 'Produit et image ajoutes avec succes' });

		const uploadFails = vi
			.fn()
			.mockResolvedValueOnce(apiResponse({ ID: 12 }, 201))
			.mockResolvedValueOnce(
				new Response(JSON.stringify({ success: false, error: { message: 'Upload down' } }), {
					status: 500
				})
			) as unknown as typeof globalThis.fetch;
		await expect(
			create({
				locals: { user },
				request: requestWithMultipart('/mes-produits', productForm, image),
				fetch: uploadFails
			} as never)
		).resolves.toMatchObject({ status: 500 });

		const invalidImage = new File(['fake'], 'notes.txt', { type: 'text/plain' });
		await expect(
			create({
				locals: { user },
				request: requestWithMultipart('/mes-produits', productForm, invalidImage),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });
	});

	it('creates, updates and deletes a seller promotion', async () => {
		const createPromotion = sellerProductActions.createPromotion!;

		await expect(
			createPromotion({
				locals: { user },
				request: requestWithForm('/mes-produits', { ...promotionForm, name: '' }),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });

		await expect(
			createPromotion({
				locals: { user },
				request: requestWithForm('/mes-produits', { ...promotionForm, type: 'unknown' }),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });

		await expect(
			createPromotion({
				locals: { user },
				request: requestWithForm('/mes-produits', {
					...promotionForm,
					type: PROMOTION_TYPE_PERCENTAGE,
					value: '150'
				}),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });

		await expect(
			createPromotion({
				locals: { user },
				request: requestWithForm('/mes-produits', { ...promotionForm, product_ids: '' }),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });

		let capturedBody: unknown;
		const createFetch = vi.fn((_input: RequestInfo | URL, init?: RequestInit) => {
			capturedBody = init?.body ? JSON.parse(String(init.body)) : null;
			return Promise.resolve(apiResponse({ ID: 9 }, 201));
		}) as unknown as typeof globalThis.fetch;
		await expect(
			createPromotion({
				locals: { user },
				request: requestWithForm('/mes-produits', promotionForm),
				fetch: createFetch
			} as never)
		).resolves.toMatchObject({ success: 'Promotion ajoutee avec succes' });
		expect(capturedBody).toMatchObject({ applies_to_all: false, product_ids: [4] });

		await expect(
			createPromotion({
				locals: { user },
				request: requestWithForm('/mes-produits', promotionForm),
				fetch: vi.fn(() =>
					Promise.resolve(
						new Response(JSON.stringify({ success: false, error: { message: 'API down' } }), {
							status: 503
						})
					)
				)
			} as never)
		).resolves.toMatchObject({ status: 503 });

		const updatePromotion = sellerProductActions.updatePromotion!;
		await expect(
			updatePromotion({
				locals: { user },
				request: requestWithForm('/mes-produits', { ...promotionForm, id: '' }),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });

		await expect(
			updatePromotion({
				locals: { user },
				request: requestWithForm('/mes-produits', { ...promotionForm, name: '' }),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });

		await expect(
			updatePromotion({
				locals: { user },
				request: requestWithForm('/mes-produits', promotionForm),
				fetch: vi.fn(() => Promise.resolve(apiResponse({ ID: 9 })))
			} as never)
		).resolves.toMatchObject({ success: 'Promotion modifiee avec succes' });

		await expect(
			updatePromotion({
				locals: { user },
				request: requestWithForm('/mes-produits', promotionForm),
				fetch: vi.fn(() =>
					Promise.resolve(
						new Response(JSON.stringify({ success: false, error: { message: 'Update down' } }), {
							status: 503
						})
					)
				)
			} as never)
		).resolves.toMatchObject({ status: 503 });

		const deletePromotion = sellerProductActions.deletePromotion!;
		await expect(
			deletePromotion({
				locals: { user },
				request: requestWithForm('/mes-produits', { id: '' }),
				fetch: vi.fn()
			} as never)
		).resolves.toMatchObject({ status: 400 });

		await expect(
			deletePromotion({
				locals: { user },
				request: requestWithForm('/mes-produits', { id: '9' }),
				fetch: vi.fn(() => Promise.resolve(apiResponse({ message: 'ok' })))
			} as never)
		).resolves.toMatchObject({ success: 'Promotion supprimee avec succes' });

		await expect(
			deletePromotion({
				locals: { user },
				request: requestWithForm('/mes-produits', { id: '9' }),
				fetch: vi.fn(() =>
					Promise.resolve(
						new Response(JSON.stringify({ success: false, error: { message: 'Delete down' } }), {
							status: 500
						})
					)
				)
			} as never)
		).resolves.toMatchObject({ status: 500 });
	});
});
