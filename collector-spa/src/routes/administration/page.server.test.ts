import { describe, expect, it, vi } from 'vitest';

import { ADMIN_ROLE, PROMOTION_TYPE_PERCENTAGE } from '$lib/types';
import { actions } from './+page.server';

const adminLocals: App.Locals = { user: { id: 1, role: ADMIN_ROLE } };

const requestWithForm = (values: Record<string, string>) =>
	new Request('http://localhost/administration', {
		method: 'POST',
		body: new URLSearchParams(values)
	});

const successfulFetch = vi.fn(() =>
	Promise.resolve(
		new Response(JSON.stringify({ success: true, data: { ID: 10 } }), {
			status: 200,
			headers: { 'Content-Type': 'application/json' }
		})
	)
) as unknown as typeof fetch;

const invokeAction = async (
	name: keyof typeof actions,
	values: Record<string, string>,
	fetch: typeof globalThis.fetch = successfulFetch
) => {
	const action = actions[name];
	if (!action) {
		throw new Error(`Missing action ${name}`);
	}

	return action({
		request: requestWithForm(values),
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
});
