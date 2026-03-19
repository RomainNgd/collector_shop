<script lang="ts">
	import ProductCard from '$lib/components/productCard.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
</script>

<section class="hero relative overflow-hidden rounded-3xl border border-white/60 p-8 md:p-12">
	<div class="hero-orb hero-orb-a"></div>
	<div class="hero-orb hero-orb-b"></div>

	<div class="relative z-10 max-w-3xl">
		<p class="mb-3 text-xs font-semibold tracking-[0.25em] text-slate-500 uppercase">
			COLLECTOR EDITION
		</p>
		<h1 class="mb-4 text-4xl leading-tight font-black text-slate-900 md:text-6xl">
			Articles rares pour collectionneurs exigeants
		</h1>
		<p class="max-w-2xl text-base text-slate-600 md:text-lg">
			Une selection d'objets iconiques, soigneusement choisie pour enrichir ta collection.
		</p>
	</div>
</section>

<section class="mt-10">
	<div class="mb-6 flex items-end justify-between">
		<div>
			<p class="text-xs font-semibold tracking-[0.18em] text-slate-500 uppercase">Catalogue</p>
			<h2 class="text-2xl font-black text-slate-900 md:text-3xl">Liste des produits</h2>
		</div>
		<p class="text-sm text-slate-500">{data.products.length} produits</p>
	</div>

	<div class="grid grid-cols-1 gap-6 sm:grid-cols-2 xl:grid-cols-3">
		{#each data.products as produit, i (produit.id)}
			<div class="stagger-card" style={`animation-delay: ${i * 80}ms`}>
				<ProductCard product={produit} />
			</div>
		{/each}
	</div>
</section>

<style>
	.hero {
		background: linear-gradient(
			125deg,
			rgba(255, 255, 255, 0.96) 0%,
			rgba(248, 250, 252, 0.92) 45%,
			rgba(239, 246, 255, 0.9) 100%
		);
		box-shadow: 0 20px 40px -35px rgba(15, 23, 42, 0.4);
	}

	.hero-orb {
		position: absolute;
		border-radius: 9999px;
		filter: blur(10px);
		pointer-events: none;
	}

	.hero-orb-a {
		top: -7rem;
		right: -3rem;
		height: 16rem;
		width: 16rem;
		background: radial-gradient(circle, rgba(37, 99, 235, 0.35), transparent 66%);
	}

	.hero-orb-b {
		bottom: -8rem;
		left: 20%;
		height: 14rem;
		width: 14rem;
		background: radial-gradient(circle, rgba(245, 158, 11, 0.25), transparent 67%);
	}

	.stagger-card {
		opacity: 0;
		transform: translateY(20px) scale(0.985);
		animation: card-in 520ms cubic-bezier(0.22, 1, 0.36, 1) forwards;
	}

	@keyframes card-in {
		to {
			opacity: 1;
			transform: translateY(0) scale(1);
		}
	}
</style>
