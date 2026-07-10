<script lang="ts">
	import { resolve } from '$app/paths';
	import ProductCard from '$lib/components/productCard.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
</script>

<section class="hero theme-panel p-8 md:p-14">
	<div class="hero-grid">
		<div class="hero-content">
			<span class="theme-pill">Selection</span>
			<p class="theme-kicker mt-5">Collector Shop</p>
			<h1 class="theme-title mt-3 max-w-xl text-4xl leading-tight font-black md:text-6xl">
				Pieces rares
			</h1>
			<p class="theme-copy mt-4 max-w-xl text-sm md:text-base">
				Des objets choisis pour completer ta collection sans detour.
			</p>

			<div class="mt-8 flex flex-wrap items-center gap-3">
				<a href={resolve('/catalogue')} class="theme-button theme-button-primary">
					Voir le catalogue
				</a>
				<span class="theme-pill theme-pill-contrast">{data.products.length} pieces disponibles</span
				>
			</div>
		</div>

		<div class="hero-art" aria-hidden="true">
			<span class="blob blob-a"></span>
			<span class="blob blob-b"></span>
			<span class="blob blob-c"></span>
			<span class="hero-gem"></span>
		</div>
	</div>
</section>

<section id="catalogue" class="mt-12 space-y-6">
	<div class="theme-section-heading">
		<div>
			<p class="theme-kicker">Selection</p>
			<h2 class="theme-title mt-3 text-3xl font-black">Apercu du shop</h2>
		</div>
		<a href={resolve('/catalogue')} class="theme-button theme-button-secondary">
			Ouvrir la liste complete
		</a>
	</div>

	<div class="grid grid-cols-1 gap-6 sm:grid-cols-2 xl:grid-cols-3">
		{#each data.products.slice(0, 6) as produit, i (produit.id)}
			<div class="stagger-card" style={`animation-delay: ${i * 90}ms`}>
				<ProductCard product={produit} />
			</div>
		{/each}
	</div>
</section>

<style>
	.hero {
		box-shadow: var(--shadow-strong);
	}

	.hero-grid {
		display: grid;
		grid-template-columns: 1.1fr 0.9fr;
		align-items: center;
		gap: 2.5rem;
	}

	.hero-art {
		position: relative;
		display: none;
		height: 22rem;
	}

	.blob {
		position: absolute;
		border-radius: 42% 58% 65% 35% / 45% 45% 55% 55%;
		filter: blur(2px);
		animation: float 7s ease-in-out infinite;
	}

	.blob-a {
		top: 0;
		right: 1rem;
		height: 15rem;
		width: 15rem;
		background: var(--gradient-primary);
		opacity: 0.95;
	}

	.blob-b {
		bottom: 0.5rem;
		right: 6rem;
		height: 9rem;
		width: 9rem;
		background: rgb(var(--color-secondary-rgb) / 0.55);
		filter: blur(6px);
		animation-delay: -2.4s;
	}

	.blob-c {
		top: 3.5rem;
		left: 0;
		height: 6rem;
		width: 6rem;
		background: rgb(var(--color-secondary-rgb) / 0.35);
		filter: blur(10px);
		animation-delay: -4.8s;
	}

	.hero-gem {
		position: absolute;
		top: 6.5rem;
		right: 6rem;
		height: 3.75rem;
		width: 3.75rem;
		background: var(--color-secondary);
		clip-path: polygon(50% 0%, 100% 38%, 82% 100%, 18% 100%, 0% 38%);
		box-shadow: var(--shadow-medium);
		animation: float 5.5s ease-in-out infinite;
		animation-delay: -1.2s;
	}

	@keyframes float {
		0%,
		100% {
			transform: translateY(0);
		}
		50% {
			transform: translateY(-14px);
		}
	}

	@media (min-width: 1024px) {
		.hero-art {
			display: block;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.blob,
		.hero-gem {
			animation: none;
		}
	}

	.stagger-card {
		opacity: 0;
		transform: translateY(10px);
		animation: card-in 220ms ease-out forwards;
	}

	@keyframes card-in {
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
</style>
