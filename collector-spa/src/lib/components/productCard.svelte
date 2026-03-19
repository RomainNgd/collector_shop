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

<article
	class="product-card group relative overflow-hidden rounded-2xl border border-white/70 bg-white/95 p-4"
>
	<div class="shine"></div>
	<a href={resolve('/produit/[id]', { id: String(product.id) })} class="block">
		<div class="media-wrap mb-4 aspect-square overflow-hidden rounded-xl">
			<img src={product.imageUrl} alt={product.name} class="media h-full w-full object-cover" />
		</div>

		<div>
			<h3 class="line-clamp-1 text-lg leading-tight font-black text-slate-900">{product.name}</h3>
			<p class="mt-2 line-clamp-2 text-sm text-slate-500">{product.description}</p>
			<p class="text-brand-primary mt-3 text-xl font-black">{product.price} EUR</p>
		</div>
	</a>

	<button
		type="button"
		onclick={handleAdd}
		class="add-btn mt-5 w-full rounded-xl px-4 py-2.5 text-sm font-bold text-white"
		class:added
	>
		{added ? 'Ajoute' : 'Ajouter au panier'}
	</button>
</article>

<style>
	.product-card {
		box-shadow:
			0 20px 32px -28px rgba(30, 41, 59, 0.35),
			inset 0 0 0 1px rgba(255, 255, 255, 0.7);
		transition:
			transform 240ms ease,
			box-shadow 260ms ease;
	}

	.product-card:hover {
		transform: translateY(-4px);
		box-shadow:
			0 24px 42px -24px rgba(30, 41, 59, 0.45),
			inset 0 0 0 1px rgba(255, 255, 255, 0.82);
	}

	.shine {
		position: absolute;
		inset: -150% 35%;
		background: linear-gradient(120deg, transparent, rgba(255, 255, 255, 0.45), transparent);
		transform: translateX(-120%) rotate(12deg);
		transition: transform 650ms ease;
		pointer-events: none;
	}

	.product-card:hover .shine {
		transform: translateX(130%) rotate(12deg);
	}

	.media-wrap {
		background: linear-gradient(145deg, rgba(226, 232, 240, 0.5), rgba(241, 245, 249, 0.8));
	}

	.media {
		transition: transform 400ms ease;
	}

	.product-card:hover .media {
		transform: scale(1.04);
	}

	.add-btn {
		background: linear-gradient(120deg, var(--color-primary), #1d4ed8);
		transition:
			transform 150ms ease,
			filter 220ms ease,
			box-shadow 220ms ease;
		box-shadow: 0 12px 24px -16px rgba(37, 99, 235, 0.95);
	}

	.add-btn:hover {
		filter: brightness(1.03);
	}

	.add-btn:active {
		transform: translateY(1px) scale(0.99);
	}

	.add-btn.added {
		animation: added-pop 620ms cubic-bezier(0.25, 0.9, 0.3, 1);
		background: linear-gradient(120deg, #0891b2, #2563eb);
	}

	@keyframes added-pop {
		0% {
			transform: scale(1);
		}
		45% {
			transform: scale(1.08);
		}
		100% {
			transform: scale(1);
		}
	}
</style>
