<script lang="ts">
	import { resolve } from '$app/paths';
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
</script>

<article class="product-card theme-card theme-hover-lift p-4">
	<a href={resolve('/produit/[id]', { id: String(product.id) })} class="block">
		<div class="media-wrap">
			<span class="theme-pill product-category">{product.category}</span>
			<img
				src={product.imageUrl}
				alt={product.name}
				class="media aspect-square w-full object-cover"
			/>
		</div>

		<div class="mt-5">
			<h3 class="theme-title line-clamp-1 text-xl leading-tight font-black">{product.name}</h3>
			<p class="theme-copy mt-3 line-clamp-2 text-sm">{product.description}</p>

			<div class="mt-5 flex items-end justify-between gap-4">
				<p class="theme-price text-2xl font-black">{product.price} EUR</p>
				<span class="product-link">Voir le detail</span>
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

	.product-category {
		position: absolute;
		top: 0.9rem;
		left: 0.9rem;
		z-index: 2;
	}

	.media {
		border-radius: 1rem;
	}

	.product-link {
		font-size: 0.82rem;
		font-weight: 800;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--color-primary);
	}

	.add-btn.added {
		animation: pulse-soft 620ms cubic-bezier(0.25, 0.9, 0.3, 1);
	}
</style>
