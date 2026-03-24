<script lang="ts">
	import { resolve } from '$app/paths';
	import type { Product } from '$lib/types';
	import type { PageData } from './$types';

	type SortField = 'name' | 'category' | 'price';
	type SortDirection = 'asc' | 'desc';

	const collator = new Intl.Collator('fr', { sensitivity: 'base', numeric: true });
	const priceFormatter = new Intl.NumberFormat('fr-FR', {
		style: 'currency',
		currency: 'EUR'
	});

	let { data }: { data: PageData } = $props();

	let sortField = $state<SortField>('name');
	let sortDirection = $state<SortDirection>('asc');
	let categoryFilter = $state('all');

	const categories = $derived.by(() =>
		[...new Set(data.products.map((product) => product.category.trim() || 'non-classe'))].sort(
			(left, right) => collator.compare(left, right)
		)
	);

	const displayedProducts = $derived.by(() => {
		const filteredProducts =
			categoryFilter === 'all'
				? data.products
				: data.products.filter(
						(product) => (product.category.trim() || 'non-classe') === categoryFilter
					);

		return [...filteredProducts].sort((left, right) => {
			let comparison = 0;

			if (sortField === 'price') {
				comparison = left.price - right.price;
			} else if (sortField === 'category') {
				comparison = collator.compare(left.category, right.category);
			} else {
				comparison = collator.compare(left.name, right.name);
			}

			if (comparison === 0) {
				comparison = collator.compare(left.name, right.name);
			}

			return sortDirection === 'asc' ? comparison : -comparison;
		});
	});

	const setSortField = (nextField: SortField) => {
		if (sortField === nextField) {
			sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
			return;
		}

		sortField = nextField;
		sortDirection = nextField === 'price' ? 'desc' : 'asc';
	};

	const getSortIndicator = (field: SortField) => {
		if (sortField !== field) {
			return '--';
		}

		return sortDirection === 'asc' ? '^' : 'v';
	};

	const formatPrice = (price: Product['price']) => priceFormatter.format(price);
</script>

<section class="catalogue-page space-y-6">
	<div class="theme-panel p-5 md:p-6">
		<div class="catalogue-toolbar">
			<div class="toolbar-status">
				<span class="theme-pill">{displayedProducts.length} produits visibles</span>
				{#if categoryFilter !== 'all'}
					<button
						type="button"
						class="theme-button theme-button-ghost clear-filter"
						onclick={() => {
							categoryFilter = 'all';
						}}
					>
						Retirer le filtre
					</button>
				{/if}
			</div>

			<div class="toolbar-controls">
				<div>
					<label for="category-filter" class="theme-label">Categorie</label>
					<select id="category-filter" bind:value={categoryFilter} class="theme-select">
						<option value="all">Toutes</option>
						{#each categories as category (category)}
							<option value={category}>{category}</option>
						{/each}
					</select>
				</div>

				<div>
					<label for="sort-field" class="theme-label">Trier par</label>
					<select id="sort-field" bind:value={sortField} class="theme-select">
						<option value="name">Nom</option>
						<option value="category">Categorie</option>
						<option value="price">Prix</option>
					</select>
				</div>

				<div>
					<label for="sort-direction" class="theme-label">Ordre</label>
					<select id="sort-direction" bind:value={sortDirection} class="theme-select">
						<option value="asc">Croissant</option>
						<option value="desc">Decroissant</option>
					</select>
				</div>
			</div>
		</div>
	</div>

	<div class="theme-panel overflow-hidden">
		{#if displayedProducts.length === 0}
			<div class="theme-empty m-6">
				<p>Aucun produit ne correspond aux filtres selectionnes.</p>
			</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="theme-table catalogue-table">
					<thead>
						<tr>
							<th class="media-col">
								<button type="button" class="sort-button" onclick={() => setSortField('name')}>
									Nom <span>{getSortIndicator('name')}</span>
								</button>
							</th>
							<th>
								<button type="button" class="sort-button" onclick={() => setSortField('category')}>
									Categorie <span>{getSortIndicator('category')}</span>
								</button>
							</th>
							<th class="price-col">
								<button
									type="button"
									class="sort-button sort-button-end"
									onclick={() => setSortField('price')}
								>
									Prix <span>{getSortIndicator('price')}</span>
								</button>
							</th>
							<th class="action-col">Action</th>
						</tr>
					</thead>
					<tbody>
						{#each displayedProducts as product (product.id)}
							<tr>
								<td>
									<div class="product-cell">
										<img src={product.imageUrl} alt={product.name} class="product-thumb" />
										<div>
											<p class="theme-title text-lg font-black">{product.name}</p>
											<p class="theme-copy mt-2 line-clamp-2 text-sm">
												{product.description}
											</p>
										</div>
									</div>
								</td>
								<td>
									<span class="theme-pill">{product.category}</span>
								</td>
								<td class="price-cell">
									<p class="theme-price text-lg font-black">{formatPrice(product.price)}</p>
								</td>
								<td>
									<a
										href={resolve('/produit/[id]', { id: String(product.id) })}
										class="theme-button theme-button-secondary action-link"
									>
										Voir
									</a>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</div>
</section>

<style>
	.catalogue-toolbar {
		display: flex;
		flex-wrap: wrap;
		align-items: end;
		justify-content: space-between;
		gap: 1.25rem;
	}

	.toolbar-status {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.75rem;
	}

	.toolbar-controls {
		display: grid;
		gap: 1rem;
		grid-template-columns: repeat(3, minmax(11rem, 1fr));
	}

	.clear-filter {
		min-height: 2.5rem;
		padding-inline: 1rem;
	}

	.catalogue-table {
		min-width: 58rem;
	}

	.media-col {
		width: 58%;
	}

	.price-col,
	.price-cell {
		text-align: right;
	}

	.action-col {
		width: 1%;
		white-space: nowrap;
	}

	.sort-button {
		display: inline-flex;
		align-items: center;
		gap: 0.45rem;
		border: 0;
		background: transparent;
		padding: 0;
		font: inherit;
		font-weight: 800;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: var(--color-primary);
	}

	.sort-button-end {
		margin-left: auto;
	}

	.product-cell {
		display: grid;
		grid-template-columns: 5rem minmax(0, 1fr);
		align-items: center;
		gap: 1rem;
	}

	.product-thumb {
		height: 5rem;
		width: 5rem;
		border-radius: 1rem;
		border: 1px solid rgb(var(--color-primary-rgb) / 0.08);
		object-fit: cover;
		background: rgb(var(--color-white-rgb) / 0.88);
	}

	.action-link {
		min-height: 2.6rem;
	}

	@media (max-width: 900px) {
		.toolbar-controls {
			grid-template-columns: 1fr;
			width: 100%;
		}
	}
</style>
