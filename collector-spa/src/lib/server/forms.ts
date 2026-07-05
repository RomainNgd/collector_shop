export const getFormString = (formData: FormData, key: string, fallback = ''): string => {
	const value = formData.get(key);
	return typeof value === 'string' ? value : fallback;
};

export const getFormStrings = (formData: FormData, key: string): string[] =>
	formData.getAll(key).filter((value): value is string => typeof value === 'string');
