<script lang="ts">
	import { formatPrice, getPromotionBadgeLabel, hasProductPromotion } from '$lib/promotions';
	import type { Product } from '$lib/types';

	type PriceSize = 'sm' | 'md' | 'lg';
	type PriceAlign = 'start' | 'end';

	let {
		product,
		size = 'md',
		align = 'start',
		showPromotionName = true
	}: {
		product: Pick<Product, 'price' | 'basePrice' | 'promotion'>;
		size?: PriceSize;
		align?: PriceAlign;
		showPromotionName?: boolean;
	} = $props();

	const promotionActive = $derived(hasProductPromotion(product));
	const promotionBadgeLabel = $derived(getPromotionBadgeLabel(product.promotion));
</script>

<div
	class="price-block"
	class:align-end={align === 'end'}
	class:compact={size === 'sm'}
	class:hero={size === 'lg'}
>
	{#if promotionActive}
		{#if promotionBadgeLabel}
			<span class="promotion-pill">{promotionBadgeLabel}</span>
		{/if}

		<div class="price-row">
			<span class="base-price">{formatPrice(product.basePrice)}</span>
			<span class="final-price">{formatPrice(product.price)}</span>
		</div>

		{#if showPromotionName && product.promotion}
			<p class="promotion-name">{product.promotion.name}</p>
		{/if}
	{:else}
		<span class="final-price">{formatPrice(product.price)}</span>
	{/if}
</div>

<style>
	.price-block {
		display: inline-flex;
		flex-direction: column;
		gap: 0.35rem;
	}

	.price-block.align-end {
		align-items: flex-end;
		text-align: right;
	}

	.price-row {
		display: inline-flex;
		align-items: baseline;
		flex-wrap: wrap;
		gap: 0.6rem;
	}

	.promotion-pill {
		display: inline-flex;
		align-items: center;
		width: fit-content;
		border-radius: 999px;
		background: linear-gradient(135deg, #f04444 0%, #c91f37 100%);
		padding: 0.3rem 0.7rem;
		font-size: 0.76rem;
		font-weight: 900;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: white;
		box-shadow: 0 10px 22px rgb(201 31 55 / 0.2);
	}

	.base-price {
		color: rgb(78 92 89 / 0.78);
		font-size: 0.95rem;
		font-weight: 700;
		text-decoration: line-through;
	}

	.final-price {
		color: var(--color-primary);
		font-size: 1.55rem;
		font-weight: 900;
		line-height: 1;
	}

	.promotion-name {
		font-size: 0.78rem;
		font-weight: 700;
		color: var(--color-ink-muted);
	}

	.price-block.compact .promotion-pill {
		padding: 0.2rem 0.55rem;
		font-size: 0.68rem;
	}

	.price-block.compact .base-price {
		font-size: 0.8rem;
	}

	.price-block.compact .final-price {
		font-size: 1.05rem;
	}

	.price-block.compact .promotion-name {
		font-size: 0.72rem;
	}

	.price-block.hero .final-price {
		font-size: 2.4rem;
	}

	.price-block.hero .base-price {
		font-size: 1.1rem;
	}
</style>
