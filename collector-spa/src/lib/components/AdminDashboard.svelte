<script lang="ts">
	import type { Category, Product, Promotion } from '$lib/types';

	let {
		products,
		categories,
		promotions
	}: {
		products: Product[];
		categories: Category[];
		promotions: Promotion[];
	} = $props();

	const dashboardMetrics = $derived.by(() => {
		const productsCount = products.length;
		const categoriesCount = categories.length;
		const promotionsCount = promotions.length;
		const activePromotionsCount = promotions.filter((promotion) => promotion.isActive).length;
		const totalCatalogValue = products.reduce((total, product) => total + product.price, 0);
		const uncategorizedCount = products.filter(
			(product) => product.category.trim().toLowerCase() === 'non-classe'
		).length;

		return {
			productsCount,
			categoriesCount,
			promotionsCount,
			activePromotionsCount,
			totalCatalogValue,
			uncategorizedCount
		};
	});

	const dashboardCards = $derived.by(() => [
		{
			label: 'Produits',
			value: dashboardMetrics.productsCount,
			description: 'Articles visibles dans le catalogue'
		},
		{
			label: 'Categories',
			value: dashboardMetrics.categoriesCount,
			description: 'Axes de classement disponibles'
		},
		{
			label: 'Promotions',
			value: dashboardMetrics.promotionsCount,
			description: `${dashboardMetrics.activePromotionsCount} actives actuellement`
		},
		{
			label: 'Valeur catalogue',
			value: `${dashboardMetrics.totalCatalogValue.toFixed(2)} EUR`,
			description: 'Somme des prix affiches'
		},
		{
			label: 'A classer',
			value: dashboardMetrics.uncategorizedCount,
			description: 'Produits encore non classes'
		}
	]);

	const categoryInsights = $derived.by(() => {
		const counts: Record<string, number> = {};
		for (const product of products) {
			const key = product.category.trim() || 'non-classe';
			counts[key] = (counts[key] ?? 0) + 1;
		}

		return Object.entries(counts)
			.map(([name, count]) => ({ name, count }))
			.sort((a, b) => b.count - a.count)
			.slice(0, 4);
	});
</script>

<section class="space-y-6">
	<div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
		{#each dashboardCards as card (card.label)}
			<article class="theme-card theme-hover-lift p-6">
				<p class="theme-kicker">{card.label}</p>
				<p class="theme-title mt-3 text-4xl font-black">{card.value}</p>
				<p class="theme-copy mt-2 text-sm">{card.description}</p>
			</article>
		{/each}
	</div>

	<div class="grid gap-6 xl:grid-cols-[1.45fr_1fr]">
		<div class="theme-panel p-6">
			<div class="flex items-start justify-between gap-4">
				<div>
					<p class="theme-kicker">Roadmap ventes</p>
					<h2 class="theme-title mt-3 text-2xl font-black">Vue ventes a venir</h2>
					<p class="theme-copy mt-3 max-w-xl">
						Bloc pret pour accueillir les stats de commandes des que l'API ventes sera disponible.
					</p>
				</div>
				<span class="theme-pill theme-pill-contrast">Bientot</span>
			</div>

			<div class="mt-6 grid gap-4 md:grid-cols-2">
				<article class="theme-card p-4">
					<p class="theme-kicker">Objectif</p>
					<p class="theme-title mt-2 font-bold">KPI commandes, CA et panier moyen</p>
				</article>
				<article class="theme-card p-4">
					<p class="theme-kicker">Suivi</p>
					<p class="theme-title mt-2 font-bold">Bloc reserve aux futurs indicateurs</p>
				</article>
			</div>
		</div>

		<div class="theme-panel p-6">
			<p class="theme-kicker">Categories</p>
			<h2 class="theme-title mt-3 text-2xl font-black">Les plus remplies</h2>
			<div class="mt-5 space-y-3">
				{#each categoryInsights as insight (insight.name)}
					<div class="insight-row">
						<p class="theme-title font-semibold">{insight.name}</p>
						<span class="theme-pill">
							{insight.count} produit{insight.count > 1 ? 's' : ''}
						</span>
					</div>
				{/each}
			</div>
		</div>
	</div>
</section>

<style>
	.insight-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		border-radius: 1.2rem;
		background: rgb(var(--color-white-rgb) / 0.78);
		padding: 0.9rem 1rem;
	}
</style>
