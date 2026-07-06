import { describe, expect, it } from 'vitest';

import { buildApiHeaders, buildInternalApiPath, getApiErrorMessage, readApiResponse } from './api';

describe('API helpers', () => {
	it('normalizes internal paths and builds optional headers', () => {
		expect(buildInternalApiPath('products')).toBe('/api/products');
		expect(buildInternalApiPath('/products')).toBe('/api/products');
		expect(buildInternalApiPath('')).toBe('/api/');
		expect(buildApiHeaders()).toEqual({});
		expect(buildApiHeaders({ token: 'token', contentType: 'application/json' })).toEqual({
			Authorization: 'Bearer token',
			'Content-Type': 'application/json'
		});
	});

	it('reads JSON, empty and invalid API responses', async () => {
		await expect(
			readApiResponse<{ id: number }>(
				new Response(JSON.stringify({ success: true, data: { id: 4 } }))
			)
		).resolves.toMatchObject({ payload: { data: { id: 4 } } });
		await expect(readApiResponse(new Response(null))).resolves.toEqual({
			rawText: null,
			payload: null
		});
		await expect(readApiResponse(new Response('not-json'))).resolves.toEqual({
			rawText: 'not-json',
			payload: null
		});
	});

	it('extracts API errors with sensible fallbacks', () => {
		const ok = new Response('{}');
		expect(
			getApiErrorMessage(ok, { rawText: '{}', payload: { success: true } }, 'fallback')
		).toBeNull();

		const failed = new Response('', { status: 400 });
		expect(
			getApiErrorMessage(
				failed,
				{ rawText: null, payload: { error: { message: 'API error' } } },
				'fallback'
			)
		).toBe('API error');
		expect(getApiErrorMessage(failed, { rawText: 'raw error', payload: null }, 'fallback')).toBe(
			'raw error'
		);
		expect(getApiErrorMessage(failed, { rawText: null, payload: null }, 'fallback')).toBe(
			'fallback'
		);
	});
});
