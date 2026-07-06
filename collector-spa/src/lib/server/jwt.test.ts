import { createHmac } from 'node:crypto';
import { describe, expect, it } from 'vitest';

import { ADMIN_ROLE } from '$lib/types';
import { claimsToUser, decodeJwtPayload, tokenIsExpired, verifyJwtPayload } from './jwt';

const base64Url = (value: string | Buffer) => Buffer.from(value).toString('base64url');

const signToken = (claims: Record<string, unknown>, secret: string) => {
	const header = base64Url(JSON.stringify({ alg: 'HS256', typ: 'JWT' }));
	const payload = base64Url(JSON.stringify(claims));
	const signature = createHmac('sha256', secret).update(`${header}.${payload}`).digest('base64url');

	return `${header}.${payload}.${signature}`;
};

describe('JWT helpers', () => {
	it('accepts a valid HS256 token and maps it to a user', () => {
		const token = signToken(
			{
				sub: 12,
				role: ADMIN_ROLE,
				exp: Math.floor(Date.now() / 1000) + 60
			},
			'test-secret'
		);

		const claims = verifyJwtPayload(token, 'test-secret');

		expect(claims).not.toBeNull();
		expect(claimsToUser(claims!)).toEqual({
			id: 12,
			role: ADMIN_ROLE,
			email: undefined
		});
	});

	it('rejects a token signed with another secret', () => {
		const token = signToken(
			{
				sub: 12,
				role: ADMIN_ROLE,
				exp: Math.floor(Date.now() / 1000) + 60
			},
			'test-secret'
		);

		expect(verifyJwtPayload(token, 'other-secret')).toBeNull();
	});

	it('rejects malformed, unsigned and expired tokens', () => {
		expect(decodeJwtPayload('invalid')).toBeNull();
		expect(decodeJwtPayload('header.invalid-json.signature')).toBeNull();
		expect(verifyJwtPayload('header.payload.signature', '')).toBeNull();
		expect(verifyJwtPayload('invalid', 'secret')).toBeNull();
		expect(verifyJwtPayload('header.payload.short', 'secret')).toBeNull();
		expect(tokenIsExpired({})).toBe(true);
		expect(tokenIsExpired({ exp: Math.floor(Date.now() / 1000) - 1 })).toBe(true);
		expect(
			claimsToUser({ sub: 1, role: ADMIN_ROLE, exp: Math.floor(Date.now() / 1000) - 1 })
		).toBeNull();
		expect(claimsToUser({ sub: 'invalid', exp: Math.floor(Date.now() / 1000) + 60 })).toBeNull();
	});

	it('normalizes supported role claim variants and string identifiers', () => {
		const exp = Math.floor(Date.now() / 1000) + 60;

		expect(claimsToUser({ sub: '9', roles: [ADMIN_ROLE], exp })).toEqual({
			id: 9,
			role: ADMIN_ROLE,
			email: undefined
		});
		expect(claimsToUser({ sub: 10, roles: ADMIN_ROLE, exp })).toMatchObject({
			id: 10,
			role: ADMIN_ROLE
		});
		expect(claimsToUser({ sub: 11, is_admin: true, exp, email: 'admin@test.local' })).toEqual({
			id: 11,
			role: ADMIN_ROLE,
			email: 'admin@test.local'
		});
		expect(claimsToUser({ sub: 12, isAdmin: true, exp })).toMatchObject({
			role: ADMIN_ROLE
		});
		expect(claimsToUser({ sub: 13, role: 'unknown', exp })).toMatchObject({
			role: 'ROLE_USER'
		});
	});
});
