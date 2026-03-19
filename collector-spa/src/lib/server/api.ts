import { env } from '$env/dynamic/private';

export const API_BASE_URL = env.API_BASE_URL ?? 'http://localhost:8080';

export type ApiErrorPayload = {
	message?: string;
};

export type ApiResponsePayload<T> = {
	success?: boolean;
	data?: T;
	error?: ApiErrorPayload;
};

export const buildApiHeaders = (options?: {
	token?: string;
	contentType?: string;
}): Record<string, string> => {
	const headers: Record<string, string> = {};

	if (options?.contentType) {
		headers['Content-Type'] = options.contentType;
	}

	if (options?.token) {
		headers.Authorization = `Bearer ${options.token}`;
	}

	return headers;
};

export const readApiResponse = async <T>(response: Response) => {
	const rawText = await response.text();

	if (!rawText) {
		return {
			rawText: null,
			payload: null as ApiResponsePayload<T> | null
		};
	}

	try {
		return {
			rawText,
			payload: JSON.parse(rawText) as ApiResponsePayload<T>
		};
	} catch {
		return {
			rawText,
			payload: null as ApiResponsePayload<T> | null
		};
	}
};

export const getApiErrorMessage = <T>(
	response: Response,
	result: { rawText: string | null; payload: ApiResponsePayload<T> | null },
	fallbackMessage: string
) => {
	if (response.ok && result.payload?.success !== false) {
		return null;
	}

	return result.payload?.error?.message ?? result.rawText ?? fallbackMessage;
};
