import type { Product, Promotion, PromotionSummary } from '$lib/types';

const currencyFormatter = new Intl.NumberFormat('fr-FR', {
	style: 'currency',
	currency: 'EUR'
});

const numberFormatter = new Intl.NumberFormat('fr-FR', {
	maximumFractionDigits: 2
});

export const formatPrice = (value: number) => currencyFormatter.format(value);

export const hasProductPromotion = (product: Pick<Product, 'price' | 'basePrice' | 'promotion'>) =>
	Boolean(product.promotion) && product.price < product.basePrice;

export const formatPromotionValue = (
	promotion: Pick<PromotionSummary | Promotion, 'type' | 'value'>
) =>
	promotion.type === 'percentage'
		? `-${numberFormatter.format(promotion.value)}%`
		: `-${formatPrice(promotion.value)}`;

export const getPromotionBadgeLabel = (promotion: PromotionSummary | null) =>
	promotion ? formatPromotionValue(promotion) : null;

export const formatPromotionScope = (
	promotion: Pick<Promotion, 'appliesToAll' | 'productCount'>
) => {
	if (promotion.appliesToAll) {
		return 'Globale';
	}

	const suffix = promotion.productCount > 1 ? 's' : '';
	return `${promotion.productCount} produit${suffix}`;
};
