<script lang="ts">
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const stats = $derived([
		{ label: 'Produits achetes', value: data.profile.productsBought },
		{ label: 'Annonces postees', value: data.profile.listingsPosted },
		{ label: 'Produits vendus', value: data.profile.productsSold }
	]);
</script>

<section class="space-y-8">
	<div class="theme-section-heading">
		<div>
			<p class="theme-kicker">Compte</p>
			<h1 class="theme-title mt-3 text-4xl font-black">Mon profil</h1>
			<p class="theme-copy mt-3 max-w-2xl">Un apercu rapide de ton activite sur Collector Shop.</p>
		</div>
		<span class="theme-pill theme-pill-contrast">{data.profile.email}</span>
	</div>

	<div class="grid gap-5 sm:grid-cols-3">
		{#each stats as stat (stat.label)}
			<div class="theme-card p-6 text-center">
				<p class="theme-price text-4xl font-black">{stat.value}</p>
				<p class="theme-copy mt-2 text-sm">{stat.label}</p>
			</div>
		{/each}
	</div>
</section>
