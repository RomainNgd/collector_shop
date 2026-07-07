<script lang="ts">
	import { resolve } from '$app/paths';
	import ProductPrice from '$lib/components/ProductPrice.svelte';
	import { PROMOTION_TYPE_FIXED, PROMOTION_TYPE_PERCENTAGE, type Product } from '$lib/types';
	import type { ActionData, PageData } from './$types';

	let { data, form }: { data: PageData; form: ActionData | null } = $props();

	const getCategorySelected = (product: Product, categoryId: number) =>
		(form?.values?.id === String(product.id)
			? form.values.categoryId
			: String(product.categoryId)) === String(categoryId);
</script>

<section class="space-y-8">
	<div class="theme-section-heading">
		<div>
			<p class="theme-kicker">Vendeur</p>
			<h1 class="theme-title mt-3 text-4xl font-black">Mes produits</h1>
			<p class="theme-copy mt-3 max-w-2xl">Pilote tes annonces, ton stock et tes promotions.</p>
		</div>
		<a href={resolve('/vendre')} class="theme-button theme-button-primary">Vendre un produit</a>
	</div>

	{#if form?.error}
		<div class="theme-alert theme-alert-error">
			<p class="theme-copy">{form.error}</p>
		</div>
	{/if}
	{#if form?.success}
		<div class="theme-alert theme-alert-success">
			<p class="theme-copy">{form.success}</p>
		</div>
	{/if}

	{#if data.products.length === 0}
		<div class="theme-empty theme-panel">
			<p class="theme-title text-lg font-bold">Aucun produit en vente.</p>
			<p class="theme-copy mt-3">Publie ta premiere annonce pour apparaitre dans le catalogue.</p>
		</div>
	{:else}
		<div class="space-y-5">
			{#each data.products as product (product.id)}
				<article class="theme-panel grid gap-5 p-5 md:grid-cols-[8rem_1fr] md:p-6">
					<img
						src={product.imageUrl}
						alt={product.name}
						class="h-32 w-32 rounded-xl object-cover"
					/>

					<div class="grid gap-5">
						<div class="flex flex-wrap items-start justify-between gap-4">
							<div>
								<p class="theme-kicker">{product.category}</p>
								<h2 class="theme-title mt-2 text-2xl font-black">{product.name}</h2>
								<p class="theme-copy mt-2 text-sm">Stock: {product.stock}</p>
							</div>
							<ProductPrice {product} />
						</div>

						<form method="POST" action="?/updateProduct" class="grid gap-4">
							<input type="hidden" name="id" value={product.id} />
							<input type="hidden" name="image" value={product.imageName ?? ''} />

							<div class="grid gap-4 md:grid-cols-2">
								<label class="grid gap-2">
									<span class="theme-label">Nom</span>
									<input class="theme-input" name="name" value={product.name} required />
								</label>
								<label class="grid gap-2">
									<span class="theme-label">Categorie</span>
									<select class="theme-select" name="category_id" required>
										{#each data.categories as category (category.id)}
											<option
												value={category.id}
												selected={getCategorySelected(product, category.id)}
											>
												{category.name}
											</option>
										{/each}
									</select>
								</label>
							</div>

							<label class="grid gap-2">
								<span class="theme-label">Description</span>
								<textarea class="theme-textarea" name="description" required
									>{product.description}</textarea
								>
							</label>

							<div class="grid gap-4 md:grid-cols-4">
								<label class="grid gap-2">
									<span class="theme-label">Prix</span>
									<input
										class="theme-input"
										name="price"
										type="number"
										min="0.01"
										step="0.01"
										value={product.basePrice}
										required
									/>
								</label>
								<label class="grid gap-2">
									<span class="theme-label">Stock</span>
									<input
										class="theme-input"
										name="stock"
										type="number"
										min="1"
										step="1"
										value={product.stock}
										required
									/>
								</label>
								<label class="grid gap-2">
									<span class="theme-label">Statut</span>
									<select class="theme-select" name="is_active">
										<option value="true" selected={product.isActive}>Actif</option>
										<option value="false" selected={!product.isActive}>Inactif</option>
									</select>
								</label>
								<label class="grid gap-2">
									<span class="theme-label">Promo active</span>
									<select class="theme-select" name="promotion_active">
										<option value="false" selected={!product.promotion}>Non</option>
										<option value="true" selected={Boolean(product.promotion)}>Oui</option>
									</select>
								</label>
							</div>

							<div class="grid gap-4 md:grid-cols-2">
								<label class="grid gap-2">
									<span class="theme-label">Type promo</span>
									<select class="theme-select" name="promotion_type">
										<option
											value={PROMOTION_TYPE_PERCENTAGE}
											selected={product.promotion?.type !== PROMOTION_TYPE_FIXED}
										>
											Pourcentage
										</option>
										<option
											value={PROMOTION_TYPE_FIXED}
											selected={product.promotion?.type === PROMOTION_TYPE_FIXED}
										>
											Montant fixe
										</option>
									</select>
								</label>
								<label class="grid gap-2">
									<span class="theme-label">Valeur promo</span>
									<input
										class="theme-input"
										name="promotion_value"
										type="number"
										min="0"
										step="0.01"
										value={product.promotion?.value ?? 0}
									/>
								</label>
							</div>

							<button type="submit" class="theme-button theme-button-primary">Enregistrer</button>
						</form>
						<form method="POST" action="?/deleteProduct">
							<input type="hidden" name="id" value={product.id} />
							<button type="submit" class="theme-button theme-button-contrast">Supprimer</button>
						</form>
					</div>
				</article>
			{/each}
		</div>
	{/if}
</section>
