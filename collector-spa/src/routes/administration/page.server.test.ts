import { describe, expect, it, vi } from 'vitest';

import { ADMIN_ROLE, PROMOTION_TYPE_FIXED, PROMOTION_TYPE_PERCENTAGE } from '$lib/types';
import { actions, load } from './+page.server';

const adminLocals: App.Locals = { user: { id: 1, role: ADMIN_ROLE } };

const requestWithForm = (values: Record<string, string>) =>
	new Request('http://localhost/administration', {
		method: 'POST',
		body: new URLSearchParams(values)
	});

const requestWithMultipart = (values: Record<string, string>, file?: File) => {
	const formData = new FormData();
	for (const [key, value] of Object.entries(values)) {
		formData.set(key, value);
	}
	if (file) {
		formData.set('image', file);
	}
	return new Request('http://localhost/administration', { method: 'POST', body: formData });
};

const successfulFetch = vi.fn(() =>
	Promise.resolve(
		new Response(JSON.stringify({ success: true, data: { ID: 10 } }), {
			status: 200,
			headers: { 'Content-Type': 'application/json' }
		})
	)
) as unknown as typeof fetch;

const fetchByPath = (responses: Record<string, Response | (() => Response)>) =>
	vi.fn((input: RequestInfo | URL) => {
		const url = typeof input === 'string' ? input : input.toString();
		const match = Object.entries(responses).find(([path]) => url.endsWith(path));
		if (!match) {
			return Promise.resolve(new Response(JSON.stringify({ success: false }), { status: 404 }));
		}
		const value = match[1];
		return Promise.resolve(typeof value === 'function' ? value() : value);
	}) as unknown as typeof fetch;

const apiResponse = (data: unknown, status = 200) =>
	new Response(JSON.stringify({ success: status < 400, data }), { status });

const invokeAction = async (
	name: keyof typeof actions,
	values: Record<string, string>,
	fetch: typeof globalThis.fetch = successfulFetch,
	request?: Request
) => {
	const action = actions[name];
	if (!action) {
		throw new Error(`Missing action ${name}`);
	}

	return action({
		request: request ?? requestWithForm(values),
		fetch,
		locals: adminLocals
	} as never);
};

describe('administration actions', () => {
	it('creates, updates and deletes products', async () => {
		const product = {
			name: 'Console',
			description: 'Console retro',
			price: '99.90',
			category_id: '2'
		};

		await expect(invokeAction('createProduct', product)).resolves.toMatchObject({
			action: 'create-product'
		});
		await expect(
			invokeAction('updateProduct', { ...product, id: '10', currentImageName: 'console.png' })
		).resolves.toMatchObject({ action: 'edit-product' });
		await expect(invokeAction('deleteProduct', { id: '10' })).resolves.toMatchObject({
			action: 'delete-product'
		});
	});

	it('creates, updates and deletes categories', async () => {
		const category = { name: 'Consoles', description: 'Consoles retro' };

		await expect(invokeAction('createCategory', category)).resolves.toMatchObject({
			action: 'create-category'
		});
		await expect(invokeAction('updateCategory', { ...category, id: '2' })).resolves.toMatchObject({
			action: 'edit-category'
		});
		await expect(invokeAction('deleteCategory', { id: '2' })).resolves.toMatchObject({
			action: 'delete-category'
		});
	});

	it('creates, updates and deletes promotions', async () => {
		const promotion = {
			name: 'Summer',
			description: 'Summer sale',
			type: PROMOTION_TYPE_PERCENTAGE,
			value: '15',
			is_active: 'true',
			applies_to_all: 'true'
		};

		await expect(invokeAction('createPromotion', promotion)).resolves.toMatchObject({
			action: 'create-promotion'
		});
		await expect(invokeAction('updatePromotion', { ...promotion, id: '3' })).resolves.toMatchObject(
			{ action: 'edit-promotion' }
		);
		await expect(invokeAction('deletePromotion', { id: '3' })).resolves.toMatchObject({
			action: 'delete-promotion'
		});
	});

	it('returns validation and API errors', async () => {
		await expect(invokeAction('createProduct', {})).resolves.toMatchObject({
			status: 400,
			data: { action: 'create-product' }
		});
		await expect(
			invokeAction('createProduct', {
				name: 'Console',
				description: 'Retro',
				price: '-1',
				category_id: 'invalid'
			})
		).resolves.toMatchObject({ status: 400 });
		await expect(
			invokeAction('createProduct', {
				name: 'Console',
				description: 'Retro',
				price: '10',
				category_id: 'invalid'
			})
		).resolves.toMatchObject({ status: 400 });
		await expect(invokeAction('updateCategory', { name: 'Category' })).resolves.toMatchObject({
			status: 400,
			data: { action: 'edit-category' }
		});
		await expect(
			invokeAction('createPromotion', {
				name: 'Invalid',
				type: 'unknown',
				value: '10'
			})
		).resolves.toMatchObject({ status: 400 });
		await expect(
			invokeAction('createPromotion', {
				name: 'Invalid',
				type: PROMOTION_TYPE_PERCENTAGE,
				value: '101'
			})
		).resolves.toMatchObject({ status: 400, data: { action: 'create-promotion' } });

		const failedFetch = vi.fn(() =>
			Promise.resolve(
				new Response(JSON.stringify({ success: false, error: { message: 'API unavailable' } }), {
					status: 503
				})
			)
		) as unknown as typeof fetch;
		await expect(
			invokeAction('createCategory', { name: 'Consoles', description: 'Retro' }, failedFetch)
		).resolves.toMatchObject({ status: 503, data: { error: 'API unavailable' } });
	});

	it('loads admin data for an admin user', async () => {
		const fetch = fetchByPath({
			'/products': apiResponse([]),
			'/categories': apiResponse([]),
			'/promotions': apiResponse([])
		});

		await expect(load({ locals: adminLocals, fetch } as never)).resolves.toMatchObject({
			products: [],
			categories: [],
			promotions: []
		});
	});

	it('validates product price, stock, category and image edge cases', async () => {
		const base = {
			name: 'Console',
			description: 'Console retro',
			price: '10',
			stock: '2',
			category_id: '2'
		};

		await expect(invokeAction('createProduct', { ...base, stock: '0' })).resolves.toMatchObject({
			status: 400
		});

		await expect(
			invokeAction('createProduct', { ...base, is_active: 'false' })
		).resolves.toMatchObject({ action: 'create-product' });

		const invalidImage = new File(['data'], 'note.txt', { type: 'text/plain' });
		await expect(
			invokeAction('createProduct', base, successfulFetch, requestWithMultipart(base, invalidImage))
		).resolves.toMatchObject({ status: 400 });
	});

	it('validates product-level promotion fields', async () => {
		const base = {
			name: 'Console',
			description: 'Console retro',
			price: '10',
			stock: '2',
			category_id: '2',
			promotion_active: 'true'
		};

		await expect(
			invokeAction('createProduct', { ...base, promotion_type: 'unknown', promotion_value: '5' })
		).resolves.toMatchObject({ status: 400 });

		await expect(
			invokeAction('createProduct', {
				...base,
				promotion_type: PROMOTION_TYPE_FIXED,
				promotion_value: '0'
			})
		).resolves.toMatchObject({ status: 400 });

		await expect(
			invokeAction('createProduct', {
				...base,
				promotion_type: PROMOTION_TYPE_PERCENTAGE,
				promotion_value: '101'
			})
		).resolves.toMatchObject({ status: 400 });

		await expect(
			invokeAction('createProduct', {
				...base,
				promotion_type: PROMOTION_TYPE_PERCENTAGE,
				promotion_value: '20'
			})
		).resolves.toMatchObject({ action: 'create-product' });
	});

	it('creates a product with an image and links it to the created id', async () => {
		const image = new File(['fake'], 'console.png', { type: 'image/png' });
		const fetch = fetchByPath({
			'/products': apiResponse({ ID: 42 }, 201),
			'/products/42/image': apiResponse({ ok: true })
		});

		await expect(
			invokeAction(
				'createProduct',
				{},
				fetch,
				requestWithMultipart(
					{ name: 'Console', description: 'Retro', price: '10', category_id: '2' },
					image
				)
			)
		).resolves.toMatchObject({
			action: 'create-product',
			success: 'Produit et image ajoutes avec succes'
		});
	});

	it('fails when the created product has no usable id for the image upload', async () => {
		const image = new File(['fake'], 'console.png', { type: 'image/png' });
		const fetch = fetchByPath({
			'/products': apiResponse({}, 201)
		});

		await expect(
			invokeAction(
				'createProduct',
				{},
				fetch,
				requestWithMultipart(
					{ name: 'Console', description: 'Retro', price: '10', category_id: '2' },
					image
				)
			)
		).resolves.toMatchObject({ status: 500 });
	});

	it('reports an error when the product image upload fails', async () => {
		const image = new File(['fake'], 'console.png', { type: 'image/png' });
		const fetch = fetchByPath({
			'/products': apiResponse({ ID: 42 }, 201),
			'/products/42/image': () =>
				new Response(JSON.stringify({ success: false, error: { message: 'Upload down' } }), {
					status: 500
				})
		});

		await expect(
			invokeAction(
				'createProduct',
				{},
				fetch,
				requestWithMultipart(
					{ name: 'Console', description: 'Retro', price: '10', category_id: '2' },
					image
				)
			)
		).resolves.toMatchObject({ status: 500 });
	});

	it('updates a product, removing and replacing its image', async () => {
		const product = {
			id: '10',
			name: 'Console',
			description: 'Retro',
			price: '10',
			stock: '2',
			category_id: '2',
			currentImageName: 'console.png'
		};

		await expect(invokeAction('updateProduct', {})).resolves.toMatchObject({
			status: 400,
			data: { action: 'edit-product' }
		});

		const removeImageFetch = fetchByPath({
			'/products/10': apiResponse({ ID: 10 }),
			'/products/10/image': apiResponse({ ok: true })
		});
		await expect(
			invokeAction(
				'updateProduct',
				{},
				removeImageFetch,
				requestWithMultipart({ ...product, removeImage: 'true' })
			)
		).resolves.toMatchObject({ success: 'Produit et image mis a jour avec succes' });

		const removeImageFailFetch = fetchByPath({
			'/products/10': apiResponse({ ID: 10 }),
			'/products/10/image': () =>
				new Response(JSON.stringify({ success: false, error: { message: 'Delete down' } }), {
					status: 500
				})
		});
		await expect(
			invokeAction(
				'updateProduct',
				{},
				removeImageFailFetch,
				requestWithMultipart({ ...product, removeImage: 'true' })
			)
		).resolves.toMatchObject({ status: 500 });

		const newImage = new File(['fake'], 'new.png', { type: 'image/png' });
		const uploadFailFetch = fetchByPath({
			'/products/10': apiResponse({ ID: 10 }),
			'/products/10/image': () =>
				new Response(JSON.stringify({ success: false, error: { message: 'Upload down' } }), {
					status: 500
				})
		});
		await expect(
			invokeAction('updateProduct', {}, uploadFailFetch, requestWithMultipart(product, newImage))
		).resolves.toMatchObject({ status: 500 });

		const uploadOkFetch = fetchByPath({
			'/products/10': apiResponse({ ID: 10 }),
			'/products/10/image': apiResponse({ ok: true })
		});
		await expect(
			invokeAction('updateProduct', {}, uploadOkFetch, requestWithMultipart(product, newImage))
		).resolves.toMatchObject({ success: 'Produit et image mis a jour avec succes' });
	});

	it('fails to delete a product without an id', async () => {
		await expect(invokeAction('deleteProduct', {})).resolves.toMatchObject({ status: 400 });
	});

	it('fails to create a category without a name', async () => {
		await expect(
			invokeAction('createCategory', { description: 'Missing name' })
		).resolves.toMatchObject({ status: 400, data: { action: 'create-category' } });
	});

	it('updates and reports API errors for categories', async () => {
		await expect(
			invokeAction(
				'updateCategory',
				{ id: '2', name: 'Consoles', description: 'Retro' },
				vi.fn(() =>
					Promise.resolve(
						new Response(JSON.stringify({ success: false, error: { message: 'down' } }), {
							status: 503
						})
					)
				) as unknown as typeof fetch
			)
		).resolves.toMatchObject({ status: 503, data: { action: 'edit-category' } });
	});

	it('fails to delete a category without an id and reports API errors', async () => {
		await expect(invokeAction('deleteCategory', {})).resolves.toMatchObject({ status: 400 });

		await expect(
			invokeAction(
				'deleteCategory',
				{ id: '2' },
				vi.fn(() =>
					Promise.resolve(
						new Response(JSON.stringify({ success: false, error: { message: 'down' } }), {
							status: 500
						})
					)
				) as unknown as typeof fetch
			)
		).resolves.toMatchObject({ status: 500, data: { action: 'delete-category' } });
	});

	it('validates required promotion fields and product scope', async () => {
		await expect(
			invokeAction('createPromotion', {
				name: 'Summer',
				type: PROMOTION_TYPE_PERCENTAGE
			})
		).resolves.toMatchObject({ status: 400 });

		await expect(
			invokeAction('createPromotion', {
				name: 'Summer',
				type: PROMOTION_TYPE_PERCENTAGE,
				value: '10',
				applies_to_all: 'false'
			})
		).resolves.toMatchObject({ status: 400, data: { action: 'create-promotion' } });

		const scoped = new URLSearchParams();
		scoped.set('name', 'Summer');
		scoped.set('type', PROMOTION_TYPE_PERCENTAGE);
		scoped.set('value', '10');
		scoped.set('applies_to_all', 'false');
		scoped.append('product_ids', '3');
		scoped.append('product_ids', 'not-a-number');
		scoped.append('product_ids', '-1');
		const scopedRequest = new Request('http://localhost/administration', {
			method: 'POST',
			body: scoped
		});
		await expect(
			invokeAction('createPromotion', {}, successfulFetch, scopedRequest)
		).resolves.toMatchObject({
			action: 'create-promotion'
		});
	});

	it('fails to update or delete a promotion without an id and reports API errors', async () => {
		await expect(
			invokeAction('updatePromotion', {
				name: 'Summer',
				type: PROMOTION_TYPE_PERCENTAGE,
				value: '10',
				applies_to_all: 'true'
			})
		).resolves.toMatchObject({ status: 400, data: { action: 'edit-promotion' } });

		await expect(
			invokeAction(
				'updatePromotion',
				{
					id: '3',
					name: 'Summer',
					type: PROMOTION_TYPE_PERCENTAGE,
					value: '10',
					applies_to_all: 'true'
				},
				vi.fn(() =>
					Promise.resolve(
						new Response(JSON.stringify({ success: false, error: { message: 'down' } }), {
							status: 500
						})
					)
				) as unknown as typeof fetch
			)
		).resolves.toMatchObject({ status: 500, data: { action: 'edit-promotion' } });

		await expect(invokeAction('deletePromotion', {})).resolves.toMatchObject({ status: 400 });

		await expect(
			invokeAction(
				'deletePromotion',
				{ id: '3' },
				vi.fn(() =>
					Promise.resolve(
						new Response(JSON.stringify({ success: false, error: { message: 'down' } }), {
							status: 500
						})
					)
				) as unknown as typeof fetch
			)
		).resolves.toMatchObject({ status: 500, data: { action: 'delete-promotion' } });
	});
});
