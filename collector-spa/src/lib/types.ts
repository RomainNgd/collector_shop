export const ADMIN_ROLE = 'ROLE_ADMIN' as const;
export const USER_ROLE = 'ROLE_USER' as const;

export type UserRole = typeof ADMIN_ROLE | typeof USER_ROLE;

export interface Product {
	id: number;
	name: string;
	description: string;
	price: number;
	imageUrl: string;
	imageName: string | null;
	categoryId: number | null;
	category: string;
}

export interface CartItem {
	product: Product;
	quantity: number;
}

export interface ApiProduct {
	ID: number;
	name: string;
	description: string;
	price: number;
	image: string;
	category?: string | ApiCategory;
	category_id?: number;
	CategoryID?: number;
}

export interface Category {
	id: number;
	name: string;
	description: string;
}

export interface ApiCategory {
	ID: number;
	name: string;
	description?: string;
}

export interface AuthUser {
	id: number;
	role: UserRole;
	email?: string;
}

const PRODUCT_IMAGE_PLACEHOLDER =
	"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 320 320'%3E%3Crect width='320' height='320' rx='32' fill='%23e2e8f0'/%3E%3Cpath d='M88 218l44-52c6-7 17-8 24-1l20 20 40-50c7-9 21-10 30-1l26 28v56H88z' fill='%2394a3b8'/%3E%3Ccircle cx='120' cy='110' r='24' fill='%23cbd5e1'/%3E%3C/svg%3E";

const normalizeImageName = (imageName: string): string | null => {
	const normalizedImageName = imageName.trim();
	return normalizedImageName === '' ? null : normalizedImageName;
};

const buildProductImageUrl = (imageName: string, apiBaseUrl: string): string => {
	const normalizedImageName = normalizeImageName(imageName);

	if (!normalizedImageName) {
		return PRODUCT_IMAGE_PLACEHOLDER;
	}

	return `${apiBaseUrl}/upload/${normalizedImageName}`;
};

export const mapApiProduct = (item: ApiProduct, apiBaseUrl: string): Product => ({
	id: item.ID,
	name: item.name,
	description: item.description,
	price: item.price,
	imageUrl: buildProductImageUrl(item.image, apiBaseUrl),
	imageName: normalizeImageName(item.image),
	categoryId:
		typeof item.CategoryID === 'number'
			? item.CategoryID
			: typeof item.category_id === 'number'
				? item.category_id
				: null,
	category:
		typeof item.category === 'string'
			? item.category
			: item.category?.name ?? 'non-classe'
});

export const mapApiCategory = (item: ApiCategory): Category => ({
	id: item.ID,
	name: item.name,
	description: item.description ?? ''
});
