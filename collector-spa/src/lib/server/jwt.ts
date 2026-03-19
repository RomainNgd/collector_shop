import { ADMIN_ROLE, USER_ROLE, type AuthUser, type UserRole } from '$lib/types';

type JwtClaims = Record<string, unknown> & {
	sub?: string | number;
	exp?: number;
	email?: string;
	role?: string;
	roles?: string[] | string;
	is_admin?: boolean;
	isAdmin?: boolean;
};

const decodeBase64Url = (value: string): string | null => {
	try {
		const normalized = value.replace(/-/g, '+').replace(/_/g, '/');
		const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '=');
		return Buffer.from(padded, 'base64').toString('utf-8');
	} catch {
		return null;
	}
};

export const decodeJwtPayload = (token: string): JwtClaims | null => {
	const parts = token.split('.');
	if (parts.length !== 3) {
		return null;
	}

	const payloadJson = decodeBase64Url(parts[1]);
	if (!payloadJson) {
		return null;
	}

	try {
		return JSON.parse(payloadJson) as JwtClaims;
	} catch {
		return null;
	}
};

const isKnownRole = (role: string): role is UserRole => role === ADMIN_ROLE || role === USER_ROLE;

const normalizeRole = (claims: JwtClaims): UserRole => {
	if (typeof claims.role === 'string') {
		const role = claims.role.trim();
		if (isKnownRole(role)) {
			return role;
		}
	}

	if (Array.isArray(claims.roles)) {
		for (const value of claims.roles) {
			if (isKnownRole(value)) {
				return value;
			}
		}
	}

	if (typeof claims.roles === 'string') {
		const role = claims.roles.trim();
		if (isKnownRole(role)) {
			return role;
		}
	}

	if (claims.is_admin === true || claims.isAdmin === true) {
		return ADMIN_ROLE;
	}

	return USER_ROLE;
};

const toNumber = (value: unknown): number | null => {
	if (typeof value === 'number' && Number.isFinite(value)) {
		return value;
	}

	if (typeof value === 'string') {
		const parsed = Number(value);
		return Number.isFinite(parsed) ? parsed : null;
	}

	return null;
};

export const tokenIsExpired = (claims: JwtClaims): boolean => {
	if (typeof claims.exp !== 'number') {
		return true;
	}
	const now = Math.floor(Date.now() / 1000);
	return claims.exp <= now;
};

export const claimsToUser = (claims: JwtClaims): AuthUser | null => {
	const userId = toNumber(claims.sub);
	if (userId === null) {
		return null;
	}

	if (tokenIsExpired(claims)) {
		return null;
	}

	return {
		id: userId,
		role: normalizeRole(claims),
		email: typeof claims.email === 'string' ? claims.email : undefined
	};
};
