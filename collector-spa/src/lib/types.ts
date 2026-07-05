export const ADMIN_ROLE = 'ROLE_ADMIN' as const;
export const USER_ROLE = 'ROLE_USER' as const;
export const PROMOTION_TYPE_PERCENTAGE = 'percentage' as const;
export const PROMOTION_TYPE_FIXED = 'fixed' as const;
export const ORDER_STATUS_AWAITING_PAYMENT = 'awaiting_payment' as const;
export const ORDER_STATUS_PREPARATION = 'preparation' as const;
export const ORDER_STATUS_SHIPPING = 'shipping' as const;
export const ORDER_STATUS_DELIVERED = 'delivered' as const;
export const ORDER_STATUS_CANCELLED = 'cancelled' as const;

export type UserRole = typeof ADMIN_ROLE | typeof USER_ROLE;
export type PromotionType = typeof PROMOTION_TYPE_PERCENTAGE | typeof PROMOTION_TYPE_FIXED;
export type OrderStatus =
	| typeof ORDER_STATUS_AWAITING_PAYMENT
	| typeof ORDER_STATUS_PREPARATION
	| typeof ORDER_STATUS_SHIPPING
	| typeof ORDER_STATUS_DELIVERED
	| typeof ORDER_STATUS_CANCELLED;

export interface PromotionSummary {
	id: number;
	name: string;
	type: PromotionType;
	value: number;
	discountAmount: number;
	appliesToAll: boolean;
}

export interface Promotion {
	id: number;
	name: string;
	description: string;
	type: PromotionType;
	value: number;
	isActive: boolean;
	appliesToAll: boolean;
	productIds: number[];
	productCount: number;
}

export interface Product {
	id: number;
	name: string;
	description: string;
	price: number;
	basePrice: number;
	imageUrl: string;
	imageName: string | null;
	categoryId: number | null;
	category: string;
	promotion: PromotionSummary | null;
}

export interface CartItem {
	product: Product;
	quantity: number;
}

export interface OrderItem {
	id: number;
	productId: number;
	productName: string;
	productDescription: string;
	productImageUrl: string;
	productImageName: string | null;
	categoryName: string;
	quantity: number;
	unitBasePrice: number;
	unitPrice: number;
	unitDiscount: number;
	lineBaseTotal: number;
	lineDiscountTotal: number;
	lineTotal: number;
	promotionId: number | null;
	promotionName: string | null;
	promotionType: PromotionType | null;
	promotionValue: number | null;
	promotionAppliesToAll: boolean;
}

export interface Order {
	id: number;
	createdAt: string;
	status: OrderStatus;
	currency: string;
	itemCount: number;
	subtotal: number;
	discountTotal: number;
	total: number;
	paymentProvider: string | null;
	paymentStatus: string | null;
	paidAt: string | null;
	stripeCheckoutExpiresAt: string | null;
	items: OrderItem[];
}

export interface ApiProduct {
	ID: number;
	name: string;
	description: string;
	price: number;
	effective_price?: number;
	image: string;
	category?: string | ApiCategory;
	category_id?: number;
	CategoryID?: number;
	applied_promotion?: ApiAppliedPromotion | null;
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

export interface ApiAppliedPromotion {
	id?: number;
	ID?: number;
	name: string;
	type: string;
	value: number;
	discount_amount?: number;
	applies_to_all?: boolean;
}

export interface ApiPromotion {
	ID: number;
	name: string;
	description?: string;
	type: string;
	value: number;
	is_active?: boolean;
	applies_to_all?: boolean;
	product_ids?: number[];
	product_count?: number;
}

export interface ApiOrderItem {
	ID: number;
	product_id?: number;
	product_name: string;
	product_description: string;
	product_image: string;
	category_name?: string;
	quantity: number;
	unit_base_price: number;
	unit_price: number;
	unit_discount: number;
	line_base_total: number;
	line_discount_total: number;
	line_total: number;
	promotion_id?: number | null;
	promotion_name?: string;
	promotion_type?: string;
	promotion_value?: number;
	promotion_applies_to_all?: boolean;
}

export interface ApiOrder {
	ID: number;
	CreatedAt?: string;
	status: string;
	currency?: string;
	item_count?: number;
	subtotal: number;
	discount_total: number;
	total: number;
	payment_provider?: string;
	payment_status?: string;
	paid_at?: string | null;
	stripe_checkout_expires_at?: string | null;
	items?: ApiOrderItem[];
}

export interface AuthUser {
	id: number;
	role: UserRole;
	email?: string;
}

const PRODUCT_IMAGE_PLACEHOLDER =
	"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 320 320'%3E%3Cdefs%3E%3ClinearGradient id='g' x1='32' x2='288' y1='20' y2='300' gradientUnits='userSpaceOnUse'%3E%3Cstop offset='0' stop-color='%23fbfdf9'/%3E%3Cstop offset='1' stop-color='%23e6f3ea'/%3E%3C/linearGradient%3E%3C/defs%3E%3Crect width='320' height='320' rx='32' fill='url(%23g)'/%3E%3Crect x='24' y='24' width='272' height='272' rx='28' fill='none' stroke='%23163329' stroke-opacity='.12'/%3E%3Cpath d='M88 218l44-52c6-7 17-8 24-1l20 20 40-50c7-9 21-10 30-1l26 28v56H88z' fill='%2398d7a9'/%3E%3Ccircle cx='120' cy='110' r='24' fill='%23163329' fill-opacity='.82'/%3E%3C/svg%3E";

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

const normalizePromotionType = (value: string): PromotionType =>
	value === PROMOTION_TYPE_FIXED ? PROMOTION_TYPE_FIXED : PROMOTION_TYPE_PERCENTAGE;

const normalizeOrderStatus = (value: string): OrderStatus => {
	switch (value) {
		case ORDER_STATUS_PREPARATION:
			return ORDER_STATUS_PREPARATION;
		case ORDER_STATUS_SHIPPING:
			return ORDER_STATUS_SHIPPING;
		case ORDER_STATUS_DELIVERED:
			return ORDER_STATUS_DELIVERED;
		case ORDER_STATUS_CANCELLED:
			return ORDER_STATUS_CANCELLED;
		default:
			return ORDER_STATUS_AWAITING_PAYMENT;
	}
};

const mapApiPromotionSummary = (
	item: ApiAppliedPromotion | null | undefined
): PromotionSummary | null => {
	if (!item) {
		return null;
	}

	const id = typeof item.id === 'number' ? item.id : (item.ID ?? 0);
	if (!Number.isFinite(id) || id <= 0) {
		return null;
	}

	return {
		id,
		name: item.name,
		type: normalizePromotionType(item.type),
		value: item.value,
		discountAmount: item.discount_amount ?? 0,
		appliesToAll: item.applies_to_all ?? false
	};
};

const getProductCategoryId = (item: ApiProduct): number | null => {
	if (typeof item.CategoryID === 'number') {
		return item.CategoryID;
	}

	return typeof item.category_id === 'number' ? item.category_id : null;
};

const getOrderItemCount = (item: ApiOrder): number => {
	if (typeof item.item_count === 'number' && Number.isFinite(item.item_count)) {
		return item.item_count;
	}

	if (!Array.isArray(item.items)) {
		return 0;
	}

	return item.items.reduce((total, orderItem) => total + orderItem.quantity, 0);
};

export const mapApiProduct = (item: ApiProduct, apiBaseUrl: string): Product => ({
	id: item.ID,
	name: item.name,
	description: item.description,
	price: typeof item.effective_price === 'number' ? item.effective_price : item.price,
	basePrice: item.price,
	imageUrl: buildProductImageUrl(item.image, apiBaseUrl),
	imageName: normalizeImageName(item.image),
	categoryId: getProductCategoryId(item),
	category:
		typeof item.category === 'string' ? item.category : (item.category?.name ?? 'non-classe'),
	promotion: mapApiPromotionSummary(item.applied_promotion)
});

export const mapApiCategory = (item: ApiCategory): Category => ({
	id: item.ID,
	name: item.name,
	description: item.description ?? ''
});

export const mapApiPromotion = (item: ApiPromotion): Promotion => ({
	id: item.ID,
	name: item.name,
	description: item.description ?? '',
	type: normalizePromotionType(item.type),
	value: item.value,
	isActive: item.is_active ?? false,
	appliesToAll: item.applies_to_all ?? false,
	productIds: Array.isArray(item.product_ids) ? item.product_ids : [],
	productCount:
		item.product_count ?? (Array.isArray(item.product_ids) ? item.product_ids.length : 0)
});

export const mapApiOrderItem = (item: ApiOrderItem, apiBaseUrl: string): OrderItem => ({
	id: item.ID,
	productId: item.product_id ?? 0,
	productName: item.product_name,
	productDescription: item.product_description,
	productImageUrl: buildProductImageUrl(item.product_image, apiBaseUrl),
	productImageName: normalizeImageName(item.product_image),
	categoryName: item.category_name ?? 'non-classe',
	quantity: item.quantity,
	unitBasePrice: item.unit_base_price,
	unitPrice: item.unit_price,
	unitDiscount: item.unit_discount,
	lineBaseTotal: item.line_base_total,
	lineDiscountTotal: item.line_discount_total,
	lineTotal: item.line_total,
	promotionId:
		typeof item.promotion_id === 'number' && Number.isFinite(item.promotion_id)
			? item.promotion_id
			: null,
	promotionName: item.promotion_name ?? null,
	promotionType:
		typeof item.promotion_type === 'string' ? normalizePromotionType(item.promotion_type) : null,
	promotionValue:
		typeof item.promotion_value === 'number' && Number.isFinite(item.promotion_value)
			? item.promotion_value
			: null,
	promotionAppliesToAll: item.promotion_applies_to_all ?? false
});

export const mapApiOrder = (item: ApiOrder, apiBaseUrl: string): Order => ({
	id: item.ID,
	createdAt: item.CreatedAt ?? '',
	status: normalizeOrderStatus(item.status),
	currency: item.currency ?? 'EUR',
	itemCount: getOrderItemCount(item),
	subtotal: item.subtotal,
	discountTotal: item.discount_total,
	total: item.total,
	paymentProvider:
		typeof item.payment_provider === 'string' && item.payment_provider.trim() !== ''
			? item.payment_provider
			: null,
	paymentStatus:
		typeof item.payment_status === 'string' && item.payment_status.trim() !== ''
			? item.payment_status
			: null,
	paidAt: typeof item.paid_at === 'string' && item.paid_at.trim() !== '' ? item.paid_at : null,
	stripeCheckoutExpiresAt:
		typeof item.stripe_checkout_expires_at === 'string' &&
		item.stripe_checkout_expires_at.trim() !== ''
			? item.stripe_checkout_expires_at
			: null,
	items: Array.isArray(item.items)
		? item.items.map((orderItem) => mapApiOrderItem(orderItem, apiBaseUrl))
		: []
});
