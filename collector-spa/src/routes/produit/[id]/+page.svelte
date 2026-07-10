<script lang="ts">
	import { resolve } from '$app/paths';
	import ProductPrice from '$lib/components/ProductPrice.svelte';
	import { getPromotionBadgeLabel } from '$lib/promotions';
	import { addToCart } from '$lib/stores/cart';
	import { onDestroy } from 'svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	let added = $state(false);
	let timeoutId: ReturnType<typeof setTimeout> | undefined;

	const handleAdd = () => {
		addToCart(data.product);
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

	const promotionBadgeLabel = $derived(getPromotionBadgeLabel(data.product.promotion));
</script>

<section class="product-page theme-panel p-6 md:p-10">
	<div class="grid grid-cols-1 gap-10 lg:grid-cols-[1.05fr_1fr]">
		<div class="image-shell">
			<span class="theme-pill image-pill">{data.product.category}</span>
			{#if promotionBadgeLabel}
				<span class="promotion-badge">{promotionBadgeLabel}</span>
			{/if}
			{#if data.product.imageName}
				<img src={data.product.imageUrl} alt={data.product.name} class="product-image" />
			{:else}
				<div class="theme-media-fallback product-image-fallback" aria-hidden="true">
					<span>{data.product.name.charAt(0).toUpperCase()}</span>
				</div>
			{/if}
		</div>

		<div class="content-shell">
			<p class="theme-kicker">Edition detaillee</p>
			<h1 class="theme-title mt-4 text-4xl leading-tight font-black md:text-5xl">
				{data.product.name}
			</h1>
			<p class="theme-copy mt-5 max-w-xl text-base md:text-lg">{data.product.description}</p>

			<div class="meta-list mt-8">
				<div class="meta-item">
					<span class="detail-label">Categorie</span>
					<span class="detail-value">{data.product.category}</span>
				</div>
				<div class="meta-item">
					<span class="detail-label">Reference</span>
					<span class="detail-value">#{data.product.id}</span>
				</div>
				<div class="meta-item">
					<span class="detail-label">Vendeur</span>
					<span class="detail-value">{data.product.sellerEmail ?? 'Vendeur marketplace'}</span>
				</div>
				<div class="meta-item">
					<span class="detail-label">Stock</span>
					<span class="detail-value">{data.product.stock}</span>
				</div>
			</div>

			<div class="mt-8">
				<ProductPrice product={data.product} size="lg" />
			</div>

			<div class="mt-8 flex flex-wrap gap-3">
				{#if data.isOwnProduct}
					<a href={resolve('/mes-produits')} class="theme-button theme-button-primary">
						Gerer mon produit
					</a>
				{:else}
					<button
						type="button"
						onclick={handleAdd}
						class="theme-button theme-button-primary buy-btn"
						class:added
					>
						{added ? 'Ajoute au panier' : 'Ajouter au panier'}
					</button>
				{/if}
				<a href={resolve('/catalogue')} class="theme-button theme-button-secondary">
					Retour catalogue
				</a>
			</div>
		</div>
	</div>
</section>

<style>
	.product-page {
		box-shadow: var(--shadow-strong);
	}

	.image-shell {
		position: relative;
		display: grid;
		place-items: center;
		overflow: hidden;
		border-radius: 1.5rem;
		border: 1px solid var(--color-border);
		background: linear-gradient(
			155deg,
			rgb(var(--color-white-rgb) / 0.98),
			rgb(var(--color-primary-rgb) / 0.04)
		);
		padding: 1.5rem;
		min-height: 24rem;
	}

	.image-pill {
		position: absolute;
		top: 1.25rem;
		left: 1.25rem;
		z-index: 2;
	}

	.promotion-badge {
		position: absolute;
		top: 1.25rem;
		right: 1.25rem;
		z-index: 2;
		border-radius: 999px;
		background: linear-gradient(135deg, #f04444 0%, #c91f37 100%);
		padding: 0.45rem 0.9rem;
		font-size: 0.8rem;
		font-weight: 900;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: white;
		box-shadow: 0 12px 24px rgb(201 31 55 / 0.18);
	}

	.product-image {
		position: relative;
		z-index: 1;
		max-height: 32rem;
		width: auto;
		object-fit: contain;
		animation: image-pop 520ms cubic-bezier(0.2, 0.9, 0.24, 1);
	}

	.product-image-fallback {
		position: relative;
		z-index: 1;
		max-width: 24rem;
		aspect-ratio: 1 / 1;
		animation: image-pop 520ms cubic-bezier(0.2, 0.9, 0.24, 1);
	}

	.product-image-fallback span {
		font-size: 4.5rem;
	}

	.meta-list {
		display: flex;
		flex-wrap: wrap;
		gap: 1rem 1.5rem;
	}

	.meta-item {
		display: flex;
		flex-direction: column;
		gap: 0.45rem;
	}

	.detail-label {
		font-size: 0.78rem;
		font-weight: 800;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: var(--color-ink-muted);
	}

	.detail-value {
		font-size: 1.1rem;
		font-weight: 800;
		color: var(--color-black);
	}

	.buy-btn.added {
		animation: pulse-soft 620ms cubic-bezier(0.25, 0.9, 0.3, 1);
	}

	@keyframes image-pop {
		from {
			opacity: 0;
			transform: scale(0.98);
		}
		to {
			opacity: 1;
			transform: scale(1);
		}
	}
</style>
