<script lang="ts">
	import { resolve } from '$app/paths';
	import ProductPrice from '$lib/components/ProductPrice.svelte';
	import { getPromotionBadgeLabel } from '$lib/promotions';
	import { addToCart } from '$lib/stores/cart';
	import type { Product } from '$lib/types';
	import { onDestroy } from 'svelte';

	let { product }: { product: Product } = $props();

	let added = $state(false);
	let timeoutId: ReturnType<typeof setTimeout> | undefined;

	const handleAdd = () => {
		addToCart(product);
		added = true;

		if (timeoutId) {
			clearTimeout(timeoutId);
		}
		timeoutId = setTimeout(() => {
			added = false;
		}, 640);
	};

	onDestroy(() => {
		if (timeoutId) {
			clearTimeout(timeoutId);
		}
	});

	const promotionBadgeLabel = $derived(getPromotionBadgeLabel(product.promotion));
</script>

<article class="product-card theme-card theme-hover-lift p-4">
	<a href={resolve('/produit/[id]', { id: String(product.id) })} class="block">
		<div class="media-wrap">
			<span class="theme-pill product-category">{product.category}</span>
			{#if promotionBadgeLabel}
				<span class="promotion-badge">{promotionBadgeLabel}</span>
			{/if}
			{#if product.imageName}
				<img
					src={product.imageUrl}
					alt={product.name}
					class="media aspect-square w-full object-cover"
				/>
			{:else}
				<div class="theme-media-fallback" aria-hidden="true">
					<span>{product.name.charAt(0).toUpperCase()}</span>
				</div>
			{/if}
		</div>

		<div class="mt-5">
			<h3 class="theme-title line-clamp-1 text-xl leading-tight font-black">{product.name}</h3>
			<p class="theme-copy mt-3 line-clamp-2 text-sm">{product.description}</p>
			<p class="theme-copy mt-2 text-xs">
				{product.sellerEmail ?? 'Marketplace'} · Stock: {product.stock}
			</p>

			<div class="mt-5 flex items-end justify-between gap-4">
				<ProductPrice {product} />
				<span class="product-link">Voir le detail <span class="product-link-arrow">→</span></span>
			</div>
		</div>
	</a>

	<button
		type="button"
		onclick={handleAdd}
		class="theme-button theme-button-primary add-btn mt-5 w-full justify-center"
		class:added
	>
		{added ? 'Ajoute' : 'Ajouter au panier'}
	</button>
</article>

<style>
	.product-card {
		height: 100%;
	}

	.media-wrap {
		position: relative;
		overflow: hidden;
		border-radius: 1.2rem;
		background: linear-gradient(
			145deg,
			rgb(var(--color-white-rgb) / 0.96),
			rgb(var(--color-primary-rgb) / 0.04)
		);
		padding: 0.9rem;
	}

	.media-wrap :global(.media) {
		transition: transform var(--transition-slow);
	}

	.product-card:hover .media-wrap :global(.media) {
		transform: scale(1.045);
	}

	.product-category {
		position: absolute;
		top: 0.9rem;
		left: 0.9rem;
		z-index: 2;
		backdrop-filter: blur(8px);
		background: rgb(var(--color-white-rgb) / 0.72);
	}

	.promotion-badge {
		position: absolute;
		top: 0.9rem;
		right: 0.9rem;
		z-index: 2;
		border-radius: 999px;
		background: linear-gradient(135deg, #f04444 0%, #c91f37 100%);
		padding: 0.32rem 0.75rem;
		font-size: 0.72rem;
		font-weight: 900;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: white;
		box-shadow: 0 10px 22px rgb(201 31 55 / 0.18);
	}

	.media {
		border-radius: 1rem;
	}

	.product-link {
		display: inline-flex;
		align-items: center;
		gap: 0.3rem;
		font-size: 0.82rem;
		font-weight: 800;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--color-primary);
	}

	.product-link-arrow {
		display: inline-block;
		transition: transform var(--transition-standard);
	}

	.product-card:hover .product-link-arrow {
		transform: translateX(3px);
	}

	.add-btn.added {
		animation: pulse-soft 620ms cubic-bezier(0.25, 0.9, 0.3, 1);
	}
</style>
