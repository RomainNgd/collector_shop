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

describe('seller pages', () => {
	it('protects seller pages from anonymous and admin users', async () => {
		await expect(
			loadSellPage({ locals: { user: null }, fetch: vi.fn() } as never)
		).rejects.toMatchObject({ status: 303, location: '/login' });
		await expect(
			loadSellerProductsPage({
				locals: { user: { id: 2, role: ADMIN_ROLE } },
				fetch: vi.fn()
			} as never)
		).rejects.toMatchObject({ status: 303, location: '/administration' });
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
					'/categories': apiResponse([{ ID: 3, name: 'Consoles' }])
				})
			} as never)
		).resolves.toMatchObject({ products: [{ id: 4 }], categories: [{ id: 3 }] });
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
		).resolves.toMatchObject({ success: 'Produit mis a jour' });

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
		).resolves.toMatchObject({ success: 'Produit supprime' });

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
});
