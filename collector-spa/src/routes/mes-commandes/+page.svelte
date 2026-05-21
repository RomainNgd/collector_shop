<script lang="ts">
	import { resolve } from '$app/paths';
	import { getOrderStatusLabel, getOrderStatusTone } from '$lib/orders';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const dateFormatter = new Intl.DateTimeFormat('fr-FR', {
		dateStyle: 'medium',
		timeStyle: 'short'
	});

	const formatOrderDate = (value: string) => {
		const parsed = new Date(value);
		return Number.isNaN(parsed.getTime()) ? 'Date indisponible' : dateFormatter.format(parsed);
	};

	const getOrderPreview = (order: PageData['orders'][number]) => {
		if (order.items.length === 0) {
			return 'Aucun article';
		}

		return order.items
			.slice(0, 2)
			.map((item) => item.productName)
			.join(', ');
	};
</script>

<section class="space-y-8">
	<div class="theme-section-heading">
		<div>
			<p class="theme-kicker">Compte</p>
			<h1 class="theme-title mt-3 text-4xl font-black">Mes commandes</h1>
			<p class="theme-copy mt-3 max-w-2xl">
				Retrouve ici tes commandes, leur recap et leur avancement logistique.
			</p>
		</div>
		<span class="theme-pill theme-pill-contrast">
			{data.orders.length} commande{data.orders.length > 1 ? 's' : ''}
		</span>
	</div>

	{#if data.orders.length === 0}
		<div class="theme-empty theme-panel">
			<p class="theme-title text-lg font-bold">Aucune commande pour le moment.</p>
			<p class="theme-copy mt-3">
				Valide ton panier pour creer une premiere commande et suivre son statut.
			</p>
			<div class="mt-6 flex justify-center">
				<a href={resolve('/catalogue')} class="theme-button theme-button-primary">
					Retour au catalogue
				</a>
			</div>
		</div>
	{:else}
		<div class="grid gap-5">
			{#each data.orders as order (order.id)}
				<a
					href={resolve('/mes-commandes/[id]', { id: String(order.id) })}
					class="theme-card theme-hover-lift order-link p-5 md:p-6"
				>
					<div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
						<div class="space-y-3">
							<div class="flex flex-wrap items-center gap-3">
								<span class={`theme-pill ${getOrderStatusTone(order.status)}`}>
									{getOrderStatusLabel(order.status)}
								</span>
								<span class="theme-copy text-sm">Commande #{order.id}</span>
							</div>
							<h2 class="theme-title text-2xl font-black">
								{order.itemCount} article{order.itemCount > 1 ? 's' : ''}
							</h2>
							<p class="theme-copy max-w-2xl">{getOrderPreview(order)}</p>
						</div>

						<div class="order-meta">
							<p class="theme-copy text-sm">Creee le</p>
							<p class="theme-title font-bold">{formatOrderDate(order.createdAt)}</p>
							<p class="theme-copy mt-4 text-sm">Total</p>
							<p class="theme-price text-2xl font-black">
								{order.total.toFixed(2)}
								{order.currency}
							</p>
						</div>
					</div>
				</a>
			{/each}
		</div>
	{/if}
</section>

<style>
	.order-link {
		display: block;
	}

	.order-meta {
		min-width: 15rem;
	}
</style>
