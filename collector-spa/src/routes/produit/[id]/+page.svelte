<script lang="ts">
	import { resolve } from '$app/paths';
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
</script>

<section class="product-page rounded-3xl border border-white/70 p-6 md:p-10">
	<div class="grid grid-cols-1 gap-10 lg:grid-cols-[1.1fr_1fr]">
		<div class="image-shell">
			<div class="image-glow"></div>
			<img src={data.product.imageUrl} alt={data.product.name} class="product-image" />
		</div>

		<div class="content-shell">
			<span class="chip">{data.product.category}</span>
			<h1 class="mt-4 text-4xl leading-tight font-black text-slate-900 md:text-5xl">
				{data.product.name}
			</h1>
			<p class="mt-4 max-w-xl text-base text-slate-600 md:text-lg">{data.product.description}</p>
			<p class="text-brand-primary mt-6 text-4xl font-black">{data.product.price} EUR</p>

			<div class="mt-8 flex flex-wrap gap-3">
				<button
					type="button"
					onclick={handleAdd}
					class="buy-btn rounded-2xl px-7 py-3 font-bold text-white"
					class:added
				>
					{added ? 'Ajoute au panier' : 'Ajouter au panier'}
				</button>
				<a href={resolve('/')} class="secondary-btn rounded-2xl px-7 py-3 font-semibold"
					>Retour catalogue</a
				>
			</div>
		</div>
	</div>
</section>

<style>
	.product-page {
		background:
			radial-gradient(circle at 90% 10%, rgba(37, 99, 235, 0.12), transparent 30%),
			radial-gradient(circle at 10% 95%, rgba(245, 158, 11, 0.15), transparent 34%),
			linear-gradient(130deg, rgba(255, 255, 255, 0.94), rgba(248, 250, 252, 0.92));
		box-shadow: 0 25px 42px -36px rgba(15, 23, 42, 0.55);
		animation: detail-in 420ms ease forwards;
	}

	.image-shell {
		position: relative;
		display: grid;
		place-items: center;
		border-radius: 1.5rem;
		background: linear-gradient(155deg, #ffffff, #eff6ff);
		padding: 1.4rem;
		overflow: hidden;
	}

	.image-glow {
		position: absolute;
		inset: -25% auto auto -15%;
		width: 18rem;
		height: 18rem;
		background: radial-gradient(circle, rgba(37, 99, 235, 0.2), transparent 70%);
		filter: blur(12px);
	}

	.product-image {
		position: relative;
		z-index: 1;
		max-height: 510px;
		width: auto;
		object-fit: contain;
		animation: image-pop 500ms cubic-bezier(0.2, 0.9, 0.24, 1);
	}

	.chip {
		display: inline-flex;
		align-items: center;
		border-radius: 9999px;
		padding: 0.35rem 0.85rem;
		font-size: 0.72rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 0.15em;
		background: rgba(245, 158, 11, 0.2);
		color: #92400e;
	}

	.buy-btn {
		background: linear-gradient(120deg, var(--color-primary), #1d4ed8);
		transition:
			transform 160ms ease,
			box-shadow 220ms ease,
			filter 220ms ease;
		box-shadow: 0 18px 28px -20px rgba(37, 99, 235, 1);
	}

	.buy-btn:hover {
		filter: brightness(1.03);
	}

	.buy-btn:active {
		transform: translateY(1px) scale(0.99);
	}

	.buy-btn.added {
		animation: added-pop 620ms cubic-bezier(0.25, 0.9, 0.3, 1);
		background: linear-gradient(120deg, #0891b2, #2563eb);
	}

	.secondary-btn {
		border: 1px solid rgba(30, 41, 59, 0.14);
		background: rgba(255, 255, 255, 0.72);
		color: #0f172a;
		transition: background-color 200ms ease;
	}

	.secondary-btn:hover {
		background: white;
	}

	@keyframes detail-in {
		from {
			opacity: 0;
			transform: translateY(14px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	@keyframes image-pop {
		from {
			opacity: 0;
			transform: scale(0.97) rotate(-0.8deg);
		}
		to {
			opacity: 1;
			transform: scale(1) rotate(0);
		}
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
