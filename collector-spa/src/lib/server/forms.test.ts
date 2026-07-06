import { describe, expect, it } from 'vitest';

import { getFormString, getFormStrings } from './forms';

describe('form helpers', () => {
	it('returns strings, fallbacks and only string arrays', () => {
		const form = new FormData();
		form.append('name', 'Collector');
		form.append('tags', 'cards');
		form.append('tags', new Blob(['image']), 'image.png');
		form.append('tags', 'games');

		expect(getFormString(form, 'name')).toBe('Collector');
		expect(getFormString(form, 'missing', 'fallback')).toBe('fallback');
		expect(getFormString(form, 'tags', 'fallback')).toBe('cards');
		expect(getFormStrings(form, 'tags')).toEqual(['cards', 'games']);
	});
});
