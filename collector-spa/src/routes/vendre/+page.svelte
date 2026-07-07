<script lang="ts">
	import { PROMOTION_TYPE_FIXED, PROMOTION_TYPE_PERCENTAGE } from '$lib/types';
	import type { ActionData, PageData } from './$types';

	let { data, form }: { data: PageData; form: ActionData | null } = $props();
	let promotionActive = $derived(form?.values?.promotionActive === 'true');
</script>

<section class="space-y-8">
	<div class="theme-section-heading">
		<div>
			<p class="theme-kicker">Marketplace</p>
			<h1 class="theme-title mt-3 text-4xl font-black">Mettre un produit en vente</h1>
			<p class="theme-copy mt-3 max-w-2xl">
				Choisis une categorie existante, fixe ton prix, ton stock et une promotion si besoin.
			</p>
		</div>
	</div>

	<form method="POST" enctype="multipart/form-data" class="theme-panel grid gap-5 p-6 md:p-8">
		{#if form?.error}
			<div class="theme-alert theme-alert-error">
				<p class="theme-copy">{form.error}</p>
			</div>
		{/if}

		<div class="grid gap-5 md:grid-cols-2">
			<label class="grid gap-2">
				<span class="theme-label">Nom</span>
				<input class="theme-input" name="name" value={form?.values?.name ?? ''} required />
			</label>

			<label class="grid gap-2">
				<span class="theme-label">Categorie</span>
				<select class="theme-select" name="category_id" required>
					<option value="">Choisir</option>
					{#each data.categories as category (category.id)}
						<option value={category.id} selected={form?.values?.categoryId === String(category.id)}>
							{category.name}
						</option>
					{/each}
				</select>
			</label>
		</div>

		<label class="grid gap-2">
			<span class="theme-label">Description</span>
			<textarea class="theme-textarea min-h-32" name="description" required
				>{form?.values?.description ?? ''}</textarea
			>
		</label>

		<div class="grid gap-5 md:grid-cols-3">
			<label class="grid gap-2">
				<span class="theme-label">Prix</span>
				<input
					class="theme-input"
					name="price"
					type="number"
					min="0.01"
					step="0.01"
					value={form?.values?.price ?? ''}
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
					value={form?.values?.stock ?? '1'}
					required
				/>
			</label>

			<label class="grid gap-2">
				<span class="theme-label">Image</span>
				<input class="theme-input" name="image" type="file" accept="image/*" />
			</label>
		</div>

		<label class="inline-flex items-center gap-3">
			<input type="checkbox" name="promotion_active" value="true" bind:checked={promotionActive} />
			<span class="theme-title text-sm font-bold">Ajouter une promotion</span>
		</label>

		{#if promotionActive}
			<div class="grid gap-5 md:grid-cols-2">
				<label class="grid gap-2">
					<span class="theme-label">Type</span>
					<select class="theme-select" name="promotion_type">
						<option
							value={PROMOTION_TYPE_PERCENTAGE}
							selected={form?.values?.promotionType !== PROMOTION_TYPE_FIXED}
						>
							Pourcentage
						</option>
						<option
							value={PROMOTION_TYPE_FIXED}
							selected={form?.values?.promotionType === PROMOTION_TYPE_FIXED}
						>
							Montant fixe
						</option>
					</select>
				</label>

				<label class="grid gap-2">
					<span class="theme-label">Valeur</span>
					<input
						class="theme-input"
						name="promotion_value"
						type="number"
						min="0.01"
						step="0.01"
						value={form?.values?.promotionValue ?? ''}
					/>
				</label>
			</div>
		{/if}

		<div>
			<button type="submit" class="theme-button theme-button-primary">Publier le produit</button>
		</div>
	</form>
</section>
