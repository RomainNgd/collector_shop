import { createHmac } from 'node:crypto';
import { describe, expect, it } from 'vitest';

import { ADMIN_ROLE } from '$lib/types';
import { claimsToUser, verifyJwtPayload } from './jwt';

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
});
