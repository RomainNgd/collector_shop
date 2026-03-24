<script lang="ts">
	import { resolve } from '$app/paths';
	import ProductCard from '$lib/components/productCard.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
</script>

<section class="hero theme-panel p-6 md:p-8">
	<span class="theme-pill">Selection</span>
	<p class="theme-kicker mt-5">Collector Shop</p>
	<h1 class="theme-title mt-3 max-w-xl text-2xl leading-tight font-black md:text-4xl">
		Pieces rares
	</h1>
	<p class="theme-copy mt-3 max-w-xl text-sm md:text-base">
		Des objets choisis pour completer ta collection sans detour.
	</p>

	<div class="mt-6 flex flex-wrap items-center gap-3">
		<a href={resolve('/catalogue')} class="theme-button theme-button-primary">
			Voir le catalogue
		</a>
		<span class="theme-pill theme-pill-contrast">{data.products.length} pieces disponibles</span>
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
		max-width: 44rem;
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
