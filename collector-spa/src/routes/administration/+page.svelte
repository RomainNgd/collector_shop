<script lang="ts">
	import AdminDashboard from '$lib/components/AdminDashboard.svelte';
	import ProductPrice from '$lib/components/ProductPrice.svelte';
	import { formatPromotionScope, formatPromotionValue } from '$lib/promotions';
	import type { Category, Product, Promotion } from '$lib/types';
	import type { ActionData, PageData } from './$types';

	type AdminSection = 'dashboard' | 'products' | 'categories' | 'promotions';

	const adminSections = [
		{ id: 'dashboard' as const, title: 'Dashboard', description: 'Vue rapide et indicateurs' },
		{ id: 'products' as const, title: 'Produits', description: 'Catalogue et images' },
		{ id: 'categories' as const, title: 'Categories', description: 'Taxonomie du shop' },
		{ id: 'promotions' as const, title: 'Promotions', description: 'Remises globales ou ciblees' }
	];

	let { data, form }: { data: PageData; form: ActionData } = $props();

	let activeSection = $state<AdminSection>('dashboard');
	let isCreateProductModalOpen = $state(false);
	let isEditProductModalOpen = $state(false);
	let isDeleteProductModalOpen = $state(false);
	let selectedProduct = $state<Product | null>(null);
	let isCreateCategoryModalOpen = $state(false);
	let isEditCategoryModalOpen = $state(false);
	let isDeleteCategoryModalOpen = $state(false);
	let selectedCategory = $state<Category | null>(null);
	let isCreatePromotionModalOpen = $state(false);
	let isEditPromotionModalOpen = $state(false);
	let isDeletePromotionModalOpen = $state(false);
	let selectedPromotion = $state<Promotion | null>(null);
	let createPromotionGlobal = $state(false);
	let createPromotionActive = $state(true);
	let editPromotionGlobal = $state(false);
	let editPromotionActive = $state(true);

	const findProductById = (id: string | undefined) => {
		const numericId = Number(id);
		return Number.isFinite(numericId)
			? (data.products.find((product) => product.id === numericId) ?? null)
			: null;
	};

	const findCategoryById = (id: string | undefined) => {
		const numericId = Number(id);
		return Number.isFinite(numericId)
			? (data.categories.find((category) => category.id === numericId) ?? null)
			: null;
	};

	const findPromotionById = (id: string | undefined) => {
		const numericId = Number(id);
		return Number.isFinite(numericId)
			? (data.promotions.find((promotion) => promotion.id === numericId) ?? null)
			: null;
	};

	const getCategoryIdFromProduct = (product: Product | null) => {
		if (!product) {
			return '';
		}

		if (typeof product.categoryId === 'number') {
			return String(product.categoryId);
		}

		const matchedCategory = data.categories.find(
			(category) => category.name.trim().toLowerCase() === product.category.trim().toLowerCase()
		);

		return matchedCategory ? String(matchedCategory.id) : '';
	};

	const countProductsForCategory = (categoryName: string) =>
		data.products.filter(
			(product) => product.category.trim().toLowerCase() === categoryName.trim().toLowerCase()
		).length;

	const countPromotionsForProduct = (productId: number) =>
		data.promotions.filter(
			(promotion) => promotion.appliesToAll || promotion.productIds.includes(productId)
		).length;

	const closeAllModals = () => {
		isCreateProductModalOpen = false;
		isEditProductModalOpen = false;
		isDeleteProductModalOpen = false;
		isCreateCategoryModalOpen = false;
		isEditCategoryModalOpen = false;
		isDeleteCategoryModalOpen = false;
		isCreatePromotionModalOpen = false;
		isEditPromotionModalOpen = false;
		isDeletePromotionModalOpen = false;
	};

	$effect(() => {
		if (form?.action === 'create-product' && (form.error || form.values)) {
			activeSection = 'products';
			isCreateProductModalOpen = true;
		}

		if (form?.action === 'edit-product') {
			activeSection = 'products';
			selectedProduct = findProductById(form.values?.id) ?? selectedProduct;
			if (form.error || form.values) {
				isEditProductModalOpen = true;
			}
		}

		if (form?.action === 'delete-product') {
			activeSection = 'products';
			selectedProduct = findProductById(form.productId) ?? selectedProduct;
			if (form.error || form.productId) {
				isDeleteProductModalOpen = true;
			}
		}

		if (form?.action === 'create-category' && (form.error || form.categoryValues)) {
			activeSection = 'categories';
			isCreateCategoryModalOpen = true;
		}

		if (form?.action === 'edit-category') {
			activeSection = 'categories';
			selectedCategory =
				findCategoryById(form.categoryValues?.id ?? form.categoryId) ?? selectedCategory;
			if (form.error || form.categoryValues) {
				isEditCategoryModalOpen = true;
			}
		}

		if (form?.action === 'delete-category') {
			activeSection = 'categories';
			selectedCategory = findCategoryById(form.categoryId) ?? selectedCategory;
			if (form.error || form.categoryId) {
				isDeleteCategoryModalOpen = true;
			}
		}

		if (form?.action === 'create-promotion' && (form.error || form.promotionValues)) {
			activeSection = 'promotions';
			isCreatePromotionModalOpen = true;
			createPromotionGlobal = form.promotionValues?.appliesToAll === 'true';
			createPromotionActive = form.promotionValues?.isActive !== 'false';
		}

		if (form?.action === 'edit-promotion') {
			activeSection = 'promotions';
			selectedPromotion =
				findPromotionById(form.promotionValues?.id ?? form.promotionId) ?? selectedPromotion;
			if (form.error || form.promotionValues) {
				isEditPromotionModalOpen = true;
			}
			editPromotionGlobal = form.promotionValues?.appliesToAll === 'true';
			editPromotionActive = form.promotionValues?.isActive !== 'false';
		}

		if (form?.action === 'delete-promotion') {
			activeSection = 'promotions';
			selectedPromotion = findPromotionById(form.promotionId) ?? selectedPromotion;
			if (form.error || form.promotionId) {
				isDeletePromotionModalOpen = true;
			}
		}

		if (form?.success) {
			closeAllModals();
			selectedProduct = null;
			selectedCategory = null;
			selectedPromotion = null;
		}
	});

	const openCreateProductModal = () => {
		closeAllModals();
		activeSection = 'products';
		isCreateProductModalOpen = true;
	};

	const openEditProductModal = (product: Product) => {
		closeAllModals();
		activeSection = 'products';
		selectedProduct = product;
		isEditProductModalOpen = true;
	};

	const openDeleteProductModal = (product: Product) => {
		closeAllModals();
		activeSection = 'products';
		selectedProduct = product;
		isDeleteProductModalOpen = true;
	};

	const openCreateCategoryModal = () => {
		closeAllModals();
		activeSection = 'categories';
		isCreateCategoryModalOpen = true;
	};

	const openEditCategoryModal = (category: Category) => {
		closeAllModals();
		activeSection = 'categories';
		selectedCategory = category;
		isEditCategoryModalOpen = true;
	};

	const openDeleteCategoryModal = (category: Category) => {
		closeAllModals();
		activeSection = 'categories';
		selectedCategory = category;
		isDeleteCategoryModalOpen = true;
	};

	const openCreatePromotionModal = () => {
		closeAllModals();
		activeSection = 'promotions';
		createPromotionGlobal = false;
		createPromotionActive = true;
		isCreatePromotionModalOpen = true;
	};

	const openEditPromotionModal = (promotion: Promotion) => {
		closeAllModals();
		activeSection = 'promotions';
		selectedPromotion = promotion;
		editPromotionGlobal = promotion.appliesToAll;
		editPromotionActive = promotion.isActive;
		isEditPromotionModalOpen = true;
	};

	const openDeletePromotionModal = (promotion: Promotion) => {
		closeAllModals();
		activeSection = 'promotions';
		selectedPromotion = promotion;
		isDeletePromotionModalOpen = true;
	};

	const closeProductEditModal = () => {
		isEditProductModalOpen = false;
		selectedProduct = null;
	};

	const closeProductDeleteModal = () => {
		isDeleteProductModalOpen = false;
		selectedProduct = null;
	};

	const closeCategoryEditModal = () => {
		isEditCategoryModalOpen = false;
		selectedCategory = null;
	};

	const closeCategoryDeleteModal = () => {
		isDeleteCategoryModalOpen = false;
		selectedCategory = null;
	};

	const closePromotionEditModal = () => {
		isEditPromotionModalOpen = false;
		selectedPromotion = null;
	};

	const closePromotionDeleteModal = () => {
		isDeletePromotionModalOpen = false;
		selectedPromotion = null;
	};
</script>

<section class="space-y-6">
	{#if form?.success}
		<div class="theme-alert theme-alert-success">
			<p class="theme-kicker">Confirmation</p>
			<p class="theme-title mt-2 text-base font-bold">{form.success}</p>
		</div>
	{/if}

	{#if form?.error}
		<div class="theme-alert theme-alert-error">
			<p class="theme-kicker">Erreur</p>
			<p class="mt-2 text-base font-bold">{form.error}</p>
		</div>
	{/if}

	<div class="theme-panel p-8">
		<div class="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
			<div>
				<span class="theme-pill">Admin center</span>
				<h1 class="theme-title mt-4 text-4xl font-black">Administration du catalogue</h1>
				<p class="theme-copy mt-3 max-w-2xl">
					Un espace unique pour piloter les produits, les categories et preparer le futur suivi des
					ventes.
				</p>
			</div>

			<div class="admin-tabs">
				{#each adminSections as section (section.id)}
					<button
						type="button"
						class="admin-tab"
						class:active={activeSection === section.id}
						onclick={() => (activeSection = section.id)}
					>
						<p class="text-sm font-black">{section.title}</p>
						<p class="mt-1 text-xs opacity-80">{section.description}</p>
					</button>
				{/each}
			</div>
		</div>
	</div>

	{#if activeSection === 'dashboard'}
		<AdminDashboard
			products={data.products}
			categories={data.categories}
			promotions={data.promotions}
		/>
	{/if}

	{#if activeSection === 'products'}
		<section class="space-y-6">
			<div class="theme-panel p-8">
				<div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
					<div>
						<p class="theme-kicker">Produits</p>
						<h2 class="theme-title mt-3 text-3xl font-black">Administration des produits</h2>
						<p class="theme-copy mt-2">Gestion complete du catalogue et des images.</p>
					</div>
					<button
						type="button"
						class="theme-button theme-button-primary"
						onclick={openCreateProductModal}
					>
						Ajouter un produit
					</button>
				</div>
			</div>

			<div class="theme-panel overflow-hidden">
				<div class="table-toolbar">
					<div>
						<p class="theme-kicker">Catalogue</p>
						<h3 class="theme-title mt-2 text-2xl font-black">Liste des produits</h3>
						<p class="theme-copy mt-2 text-sm">{data.products.length} produits</p>
					</div>
				</div>

				<div class="overflow-x-auto">
					<table class="theme-table min-w-full">
						<thead>
							<tr>
								<th>Produit</th>
								<th>Description</th>
								<th>Categorie</th>
								<th>Image</th>
								<th>Promotions</th>
								<th>Prix</th>
								<th>Actions</th>
							</tr>
						</thead>
						<tbody>
							{#each data.products as product (product.id)}
								<tr>
									<td>
										<div class="row-media">
											<img
												src={product.imageUrl}
												alt={product.name}
												class="h-16 w-16 rounded-2xl object-cover"
											/>
											<div>
												<p class="theme-title font-bold">{product.name}</p>
												<p class="theme-copy mt-1 text-sm">ID #{product.id}</p>
											</div>
										</div>
									</td>
									<td class="theme-copy text-sm">{product.description}</td>
									<td class="theme-copy text-sm">{product.category}</td>
									<td class="theme-copy text-sm">{product.imageName ?? 'Aucune image'}</td>
									<td class="theme-copy text-sm">{countPromotionsForProduct(product.id)}</td>
									<td>
										<ProductPrice {product} size="sm" />
									</td>
									<td>
										<div class="action-row">
											<button
												type="button"
												class="theme-button theme-button-secondary action-button"
												onclick={() => openEditProductModal(product)}
											>
												Modifier
											</button>
											<button
												type="button"
												class="theme-button theme-button-contrast action-button"
												onclick={() => openDeleteProductModal(product)}
											>
												Supprimer
											</button>
										</div>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		</section>
	{/if}

	{#if activeSection === 'categories'}
		<section class="space-y-6">
			<div class="theme-panel p-8">
				<div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
					<div>
						<p class="theme-kicker">Categories</p>
						<h2 class="theme-title mt-3 text-3xl font-black">Administration des categories</h2>
						<p class="theme-copy mt-2">
							Organise la taxonomie pour rendre le catalogue plus clair.
						</p>
					</div>
					<button
						type="button"
						class="theme-button theme-button-primary"
						onclick={openCreateCategoryModal}
					>
						Ajouter une categorie
					</button>
				</div>
			</div>

			<div class="theme-panel overflow-hidden">
				<div class="table-toolbar">
					<div>
						<p class="theme-kicker">Taxonomie</p>
						<h3 class="theme-title mt-2 text-2xl font-black">Liste des categories</h3>
						<p class="theme-copy mt-2 text-sm">{data.categories.length} categories</p>
					</div>
				</div>

				<div class="overflow-x-auto">
					<table class="theme-table min-w-full">
						<thead>
							<tr>
								<th>Nom</th>
								<th>Produits lies</th>
								<th>Actions</th>
							</tr>
						</thead>
						<tbody>
							{#each data.categories as category (category.id)}
								<tr>
									<td>
										<p class="theme-title font-bold">{category.name}</p>
										<p class="theme-copy mt-1 text-sm">ID #{category.id}</p>
										<p class="theme-copy mt-3 max-w-md text-sm">
											{category.description || 'Aucune description'}
										</p>
									</td>
									<td class="theme-copy text-sm">{countProductsForCategory(category.name)}</td>
									<td>
										<div class="action-row">
											<button
												type="button"
												class="theme-button theme-button-secondary action-button"
												onclick={() => openEditCategoryModal(category)}
											>
												Modifier
											</button>
											<button
												type="button"
												class="theme-button theme-button-contrast action-button"
												onclick={() => openDeleteCategoryModal(category)}
											>
												Supprimer
											</button>
										</div>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		</section>
	{/if}

	{#if activeSection === 'promotions'}
		<section class="space-y-6">
			<div class="theme-panel p-8">
				<div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
					<div>
						<p class="theme-kicker">Promotions</p>
						<h2 class="theme-title mt-3 text-3xl font-black">Administration des promotions</h2>
						<p class="theme-copy mt-2">
							Configure des remises globales ou ciblees, en pourcentage ou en montant fixe.
						</p>
					</div>
					<button
						type="button"
						class="theme-button theme-button-primary"
						onclick={openCreatePromotionModal}
					>
						Ajouter une promotion
					</button>
				</div>
			</div>

			<div class="theme-panel overflow-hidden">
				<div class="table-toolbar">
					<div>
						<p class="theme-kicker">Remises</p>
						<h3 class="theme-title mt-2 text-2xl font-black">Liste des promotions</h3>
						<p class="theme-copy mt-2 text-sm">{data.promotions.length} promotions</p>
					</div>
				</div>

				<div class="overflow-x-auto">
					<table class="theme-table min-w-full">
						<thead>
							<tr>
								<th>Promotion</th>
								<th>Portee</th>
								<th>Remise</th>
								<th>Statut</th>
								<th>Actions</th>
							</tr>
						</thead>
						<tbody>
							{#each data.promotions as promotion (promotion.id)}
								<tr>
									<td>
										<p class="theme-title font-bold">{promotion.name}</p>
										<p class="theme-copy mt-1 text-sm">ID #{promotion.id}</p>
										<p class="theme-copy mt-3 max-w-md text-sm">
											{promotion.description || 'Aucune description'}
										</p>
									</td>
									<td class="theme-copy text-sm">{formatPromotionScope(promotion)}</td>
									<td class="theme-title font-semibold">{formatPromotionValue(promotion)}</td>
									<td>
										<span class={`theme-pill ${promotion.isActive ? '' : 'theme-pill-contrast'}`}>
											{promotion.isActive ? 'Active' : 'Inactive'}
										</span>
									</td>
									<td>
										<div class="action-row">
											<button
												type="button"
												class="theme-button theme-button-secondary action-button"
												onclick={() => openEditPromotionModal(promotion)}
											>
												Modifier
											</button>
											<button
												type="button"
												class="theme-button theme-button-contrast action-button"
												onclick={() => openDeletePromotionModal(promotion)}
											>
												Supprimer
											</button>
										</div>
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</div>
		</section>
	{/if}
</section>

{#if isCreateProductModalOpen}
	<div class="theme-overlay fixed inset-0 z-40 flex items-center justify-center px-4 py-8">
		<div class="theme-modal w-full max-w-xl p-8">
			<div class="modal-header">
				<div>
					<p class="theme-kicker">Nouveau produit</p>
					<h2 class="theme-title mt-3 text-2xl font-black">Ajouter un produit</h2>
					<p class="theme-copy mt-2 text-sm">
						Le produit sera cree puis l'image sera envoyee si un fichier est selectionne.
					</p>
				</div>
				<button
					type="button"
					class="theme-icon-button theme-close"
					onclick={() => (isCreateProductModalOpen = false)}
				>
					X
				</button>
			</div>

			<form
				method="POST"
				action="?/createProduct"
				enctype="multipart/form-data"
				class="mt-6 space-y-4"
			>
				<div>
					<label for="create-name" class="theme-label">Nom</label>
					<input
						id="create-name"
						name="name"
						type="text"
						required
						value={form?.action === 'create-product' ? (form.values?.name ?? '') : ''}
						class="theme-input"
					/>
				</div>
				<div>
					<label for="create-price" class="theme-label">Prix</label>
					<input
						id="create-price"
						name="price"
						type="number"
						min="0.01"
						step="0.01"
						required
						value={form?.action === 'create-product' ? (form.values?.price ?? '') : ''}
						class="theme-input"
					/>
				</div>
				<div>
					<label for="create-category" class="theme-label">Categorie</label>
					<select id="create-category" name="category_id" required class="theme-select">
						<option value="">Selectionner une categorie</option>
						{#each data.categories as category (category.id)}
							<option
								value={category.id}
								selected={form?.action === 'create-product'
									? form.values?.categoryId === String(category.id)
									: false}
							>
								{category.name}
							</option>
						{/each}
					</select>
				</div>
				<div>
					<label for="create-description" class="theme-label">Description</label>
					<textarea
						id="create-description"
						name="description"
						rows="4"
						required
						class="theme-textarea"
						>{form?.action === 'create-product' ? (form.values?.description ?? '') : ''}</textarea
					>
				</div>
				<div class="theme-card p-4">
					<label for="create-image" class="theme-label">Image</label>
					<input id="create-image" name="image" type="file" accept="image/*" class="theme-file" />
				</div>
				<div class="modal-actions">
					<button
						type="button"
						class="theme-button theme-button-secondary"
						onclick={() => (isCreateProductModalOpen = false)}
					>
						Annuler
					</button>
					<button type="submit" class="theme-button theme-button-primary">Valider</button>
				</div>
			</form>
		</div>
	</div>
{/if}

{#if isEditProductModalOpen}
	<div class="theme-overlay fixed inset-0 z-40 flex items-center justify-center px-4 py-8">
		<div class="theme-modal w-full max-w-xl p-8">
			<div class="modal-header">
				<div>
					<p class="theme-kicker">Edition</p>
					<h2 class="theme-title mt-3 text-2xl font-black">Modifier un produit</h2>
					<p class="theme-copy mt-2 text-sm">
						Ajoute une nouvelle image pour la remplacer, ou coche la suppression si besoin.
					</p>
				</div>
				<button type="button" class="theme-icon-button theme-close" onclick={closeProductEditModal}>
					X
				</button>
			</div>

			<form
				method="POST"
				action="?/updateProduct"
				enctype="multipart/form-data"
				class="mt-6 space-y-4"
			>
				<input
					type="hidden"
					name="id"
					value={form?.action === 'edit-product'
						? (form.values?.id ?? selectedProduct?.id ?? '')
						: (selectedProduct?.id ?? '')}
				/>
				<input
					type="hidden"
					name="currentImageName"
					value={form?.action === 'edit-product'
						? (form.values?.currentImageName ?? selectedProduct?.imageName ?? '')
						: (selectedProduct?.imageName ?? '')}
				/>
				<div>
					<label for="edit-name" class="theme-label">Nom</label>
					<input
						id="edit-name"
						name="name"
						type="text"
						required
						value={form?.action === 'edit-product'
							? (form.values?.name ?? selectedProduct?.name ?? '')
							: (selectedProduct?.name ?? '')}
						class="theme-input"
					/>
				</div>
				<div>
					<label for="edit-price" class="theme-label">Prix</label>
					<input
						id="edit-price"
						name="price"
						type="number"
						min="0.01"
						step="0.01"
						required
						value={form?.action === 'edit-product'
							? (form.values?.price ?? String(selectedProduct?.basePrice ?? ''))
							: String(selectedProduct?.basePrice ?? '')}
						class="theme-input"
					/>
				</div>
				<div>
					<label for="edit-category" class="theme-label">Categorie</label>
					<select id="edit-category" name="category_id" required class="theme-select">
						<option value="">Selectionner une categorie</option>
						{#each data.categories as category (category.id)}
							<option
								value={category.id}
								selected={form?.action === 'edit-product'
									? form.values?.categoryId === String(category.id)
									: getCategoryIdFromProduct(selectedProduct) === String(category.id)}
							>
								{category.name}
							</option>
						{/each}
					</select>
				</div>
				<div>
					<label for="edit-description" class="theme-label">Description</label>
					<textarea
						id="edit-description"
						name="description"
						rows="4"
						required
						class="theme-textarea"
						>{form?.action === 'edit-product'
							? (form.values?.description ?? selectedProduct?.description ?? '')
							: (selectedProduct?.description ?? '')}</textarea
					>
				</div>
				<div class="theme-card image-preview-card p-4">
					<div>
						<p class="theme-label">Image actuelle</p>
						<img
							src={selectedProduct?.imageUrl}
							alt={selectedProduct?.name ?? 'Produit'}
							class="h-28 w-28 rounded-2xl object-cover"
						/>
					</div>
					<div class="space-y-4">
						<div>
							<p class="theme-title font-semibold">
								{selectedProduct?.imageName ?? 'Aucune image enregistree'}
							</p>
							<p class="theme-copy mt-1 text-xs">
								L'image est servie depuis <code>/upload/&lt;filename&gt;</code>.
							</p>
						</div>
						<div>
							<label for="edit-image" class="theme-label">Remplacer l'image</label>
							<input id="edit-image" name="image" type="file" accept="image/*" class="theme-file" />
						</div>
						<label class="checkbox-card">
							<input
								type="checkbox"
								name="removeImage"
								value="true"
								checked={form?.action === 'edit-product'
									? form.values?.removeImage === 'true'
									: false}
							/>
							<span class="theme-title text-sm font-medium">Supprimer l'image actuelle</span>
						</label>
					</div>
				</div>
				<div class="modal-actions">
					<button
						type="button"
						class="theme-button theme-button-secondary"
						onclick={closeProductEditModal}
					>
						Annuler
					</button>
					<button type="submit" class="theme-button theme-button-primary">Enregistrer</button>
				</div>
			</form>
		</div>
	</div>
{/if}

{#if isDeleteProductModalOpen && selectedProduct}
	<div class="theme-overlay fixed inset-0 z-40 flex items-center justify-center px-4 py-8">
		<div class="theme-modal w-full max-w-md p-8">
			<div class="modal-header">
				<div>
					<p class="theme-kicker">Suppression</p>
					<h2 class="theme-title mt-3 text-2xl font-black">Confirmer la suppression</h2>
					<p class="theme-copy mt-2 text-sm">
						Cette action supprimera definitivement le produit selectionne.
					</p>
				</div>
				<button
					type="button"
					class="theme-icon-button theme-close"
					onclick={closeProductDeleteModal}
				>
					X
				</button>
			</div>

			<div class="danger-card mt-6">
				<p class="theme-title font-bold">{selectedProduct.name}</p>
				<p class="theme-copy mt-2 text-sm">{selectedProduct.description}</p>
				<p class="theme-copy mt-3 text-sm">Image: {selectedProduct.imageName ?? 'Aucune image'}</p>
				<p class="theme-title mt-4 font-semibold">
					ID #{selectedProduct.id} - {selectedProduct.price.toFixed(2)} EUR
				</p>
			</div>

			<form method="POST" action="?/deleteProduct" class="modal-actions mt-6">
				<input type="hidden" name="id" value={selectedProduct.id} />
				<button
					type="button"
					class="theme-button theme-button-secondary"
					onclick={closeProductDeleteModal}
				>
					Annuler
				</button>
				<button type="submit" class="theme-button theme-button-contrast">
					Confirmer la suppression
				</button>
			</form>
		</div>
	</div>
{/if}

{#if isCreateCategoryModalOpen}
	<div class="theme-overlay fixed inset-0 z-40 flex items-center justify-center px-4 py-8">
		<div class="theme-modal w-full max-w-lg p-8">
			<div class="modal-header">
				<div>
					<p class="theme-kicker">Nouvelle categorie</p>
					<h2 class="theme-title mt-3 text-2xl font-black">Ajouter une categorie</h2>
					<p class="theme-copy mt-2 text-sm">
						Une categorie claire aide a mieux organiser le catalogue.
					</p>
				</div>
				<button
					type="button"
					class="theme-icon-button theme-close"
					onclick={() => (isCreateCategoryModalOpen = false)}
				>
					X
				</button>
			</div>

			<form method="POST" action="?/createCategory" class="mt-6 space-y-4">
				<div>
					<label for="create-category-name" class="theme-label">Nom</label>
					<input
						id="create-category-name"
						name="name"
						type="text"
						required
						value={form?.action === 'create-category' ? (form.categoryValues?.name ?? '') : ''}
						class="theme-input"
					/>
				</div>
				<div>
					<label for="create-category-description" class="theme-label">Description</label>
					<textarea
						id="create-category-description"
						name="description"
						rows="4"
						class="theme-textarea"
						>{form?.action === 'create-category'
							? (form.categoryValues?.description ?? '')
							: ''}</textarea
					>
				</div>
				<div class="modal-actions">
					<button
						type="button"
						class="theme-button theme-button-secondary"
						onclick={() => (isCreateCategoryModalOpen = false)}
					>
						Annuler
					</button>
					<button type="submit" class="theme-button theme-button-primary">Valider</button>
				</div>
			</form>
		</div>
	</div>
{/if}

{#if isEditCategoryModalOpen}
	<div class="theme-overlay fixed inset-0 z-40 flex items-center justify-center px-4 py-8">
		<div class="theme-modal w-full max-w-lg p-8">
			<div class="modal-header">
				<div>
					<p class="theme-kicker">Edition categorie</p>
					<h2 class="theme-title mt-3 text-2xl font-black">Modifier une categorie</h2>
					<p class="theme-copy mt-2 text-sm">
						Renomme la categorie pour garder le classement clair.
					</p>
				</div>
				<button
					type="button"
					class="theme-icon-button theme-close"
					onclick={closeCategoryEditModal}
				>
					X
				</button>
			</div>

			<form method="POST" action="?/updateCategory" class="mt-6 space-y-4">
				<input
					type="hidden"
					name="id"
					value={form?.action === 'edit-category'
						? (form.categoryValues?.id ?? selectedCategory?.id ?? '')
						: (selectedCategory?.id ?? '')}
				/>
				<div>
					<label for="edit-category-name" class="theme-label">Nom</label>
					<input
						id="edit-category-name"
						name="name"
						type="text"
						required
						value={form?.action === 'edit-category'
							? (form.categoryValues?.name ?? selectedCategory?.name ?? '')
							: (selectedCategory?.name ?? '')}
						class="theme-input"
					/>
				</div>
				<div>
					<label for="edit-category-description" class="theme-label">Description</label>
					<textarea
						id="edit-category-description"
						name="description"
						rows="4"
						class="theme-textarea"
						>{form?.action === 'edit-category'
							? (form.categoryValues?.description ?? selectedCategory?.description ?? '')
							: (selectedCategory?.description ?? '')}</textarea
					>
				</div>
				<div class="modal-actions">
					<button
						type="button"
						class="theme-button theme-button-secondary"
						onclick={closeCategoryEditModal}
					>
						Annuler
					</button>
					<button type="submit" class="theme-button theme-button-primary">Enregistrer</button>
				</div>
			</form>
		</div>
	</div>
{/if}

{#if isDeleteCategoryModalOpen && selectedCategory}
	<div class="theme-overlay fixed inset-0 z-40 flex items-center justify-center px-4 py-8">
		<div class="theme-modal w-full max-w-md p-8">
			<div class="modal-header">
				<div>
					<p class="theme-kicker">Suppression</p>
					<h2 class="theme-title mt-3 text-2xl font-black">Supprimer la categorie</h2>
					<p class="theme-copy mt-2 text-sm">
						Confirme la suppression avant de retirer cette categorie.
					</p>
				</div>
				<button
					type="button"
					class="theme-icon-button theme-close"
					onclick={closeCategoryDeleteModal}
				>
					X
				</button>
			</div>

			<div class="danger-card mt-6">
				<p class="theme-title font-bold">{selectedCategory.name}</p>
				<p class="theme-copy mt-2 text-sm">ID #{selectedCategory.id}</p>
				<p class="theme-copy mt-2 text-sm">
					{selectedCategory.description || 'Aucune description'}
				</p>
			</div>

			<form method="POST" action="?/deleteCategory" class="modal-actions mt-6">
				<input type="hidden" name="id" value={selectedCategory.id} />
				<button
					type="button"
					class="theme-button theme-button-secondary"
					onclick={closeCategoryDeleteModal}
				>
					Annuler
				</button>
				<button type="submit" class="theme-button theme-button-contrast">
					Confirmer la suppression
				</button>
			</form>
		</div>
	</div>
{/if}

{#if isCreatePromotionModalOpen}
	<div class="theme-overlay fixed inset-0 z-40 flex items-center justify-center px-4 py-8">
		<div class="theme-modal w-full max-w-3xl p-8">
			<div class="modal-header">
				<div>
					<p class="theme-kicker">Nouvelle promotion</p>
					<h2 class="theme-title mt-3 text-2xl font-black">Ajouter une promotion</h2>
					<p class="theme-copy mt-2 text-sm">
						Choisis une remise globale ou rattache-la a un ou plusieurs produits.
					</p>
				</div>
				<button
					type="button"
					class="theme-icon-button theme-close"
					onclick={() => (isCreatePromotionModalOpen = false)}
				>
					X
				</button>
			</div>

			<form method="POST" action="?/createPromotion" class="mt-6 space-y-4">
				<div class="promotion-form-grid">
					<div>
						<label for="create-promotion-name" class="theme-label">Nom</label>
						<input
							id="create-promotion-name"
							name="name"
							type="text"
							required
							value={form?.action === 'create-promotion' ? (form.promotionValues?.name ?? '') : ''}
							class="theme-input"
						/>
					</div>
					<div>
						<label for="create-promotion-type" class="theme-label">Type</label>
						<select id="create-promotion-type" name="type" required class="theme-select">
							<option
								value="percentage"
								selected={form?.action === 'create-promotion'
									? (form.promotionValues?.type ?? 'percentage') === 'percentage'
									: true}
							>
								Pourcentage
							</option>
							<option
								value="fixed"
								selected={form?.action === 'create-promotion'
									? form.promotionValues?.type === 'fixed'
									: false}
							>
								Montant fixe
							</option>
						</select>
					</div>
					<div>
						<label for="create-promotion-value" class="theme-label">Valeur</label>
						<input
							id="create-promotion-value"
							name="value"
							type="number"
							min="0.01"
							step="0.01"
							required
							value={form?.action === 'create-promotion' ? (form.promotionValues?.value ?? '') : ''}
							class="theme-input"
						/>
					</div>
				</div>

				<div>
					<label for="create-promotion-description" class="theme-label">Description</label>
					<textarea
						id="create-promotion-description"
						name="description"
						rows="4"
						class="theme-textarea"
						>{form?.action === 'create-promotion'
							? (form.promotionValues?.description ?? '')
							: ''}</textarea
					>
				</div>

				<div class="promotion-toggle-grid">
					<label class="checkbox-card">
						<input
							type="checkbox"
							name="is_active"
							value="true"
							bind:checked={createPromotionActive}
						/>
						<span class="theme-title text-sm font-medium">Promotion active</span>
					</label>
					<label class="checkbox-card">
						<input
							type="checkbox"
							name="applies_to_all"
							value="true"
							bind:checked={createPromotionGlobal}
						/>
						<span class="theme-title text-sm font-medium">Appliquer a tout le catalogue</span>
					</label>
				</div>

				<div class="theme-card p-4">
					<div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
						<div>
							<p class="theme-label">Produits cibles</p>
							<p class="theme-copy mt-1 text-sm">
								Selectionne les produits concernes quand la promotion n'est pas globale.
							</p>
						</div>
						<span class="theme-pill">{data.products.length} produits</span>
					</div>

					<div class="promotion-product-grid mt-4">
						{#each data.products as product (product.id)}
							<label class="promotion-product-card" class:disabled={createPromotionGlobal}>
								<input
									type="checkbox"
									name="product_ids"
									value={product.id}
									checked={form?.action === 'create-promotion'
										? (form.promotionValues?.productIds ?? []).includes(String(product.id))
										: false}
									disabled={createPromotionGlobal}
								/>
								<div>
									<p class="theme-title text-sm font-bold">{product.name}</p>
									<p class="theme-copy mt-1 text-xs">{product.category}</p>
									<div class="mt-2">
										<ProductPrice {product} size="sm" showPromotionName={false} />
									</div>
								</div>
							</label>
						{/each}
					</div>
				</div>

				<div class="modal-actions">
					<button
						type="button"
						class="theme-button theme-button-secondary"
						onclick={() => (isCreatePromotionModalOpen = false)}
					>
						Annuler
					</button>
					<button type="submit" class="theme-button theme-button-primary">Valider</button>
				</div>
			</form>
		</div>
	</div>
{/if}

{#if isEditPromotionModalOpen}
	<div class="theme-overlay fixed inset-0 z-40 flex items-center justify-center px-4 py-8">
		<div class="theme-modal w-full max-w-3xl p-8">
			<div class="modal-header">
				<div>
					<p class="theme-kicker">Edition promotion</p>
					<h2 class="theme-title mt-3 text-2xl font-black">Modifier une promotion</h2>
					<p class="theme-copy mt-2 text-sm">
						Ajuste la portee, le type de remise et les produits cibles si besoin.
					</p>
				</div>
				<button
					type="button"
					class="theme-icon-button theme-close"
					onclick={closePromotionEditModal}
				>
					X
				</button>
			</div>

			<form method="POST" action="?/updatePromotion" class="mt-6 space-y-4">
				<input
					type="hidden"
					name="id"
					value={form?.action === 'edit-promotion'
						? (form.promotionValues?.id ?? selectedPromotion?.id ?? '')
						: (selectedPromotion?.id ?? '')}
				/>
				<div class="promotion-form-grid">
					<div>
						<label for="edit-promotion-name" class="theme-label">Nom</label>
						<input
							id="edit-promotion-name"
							name="name"
							type="text"
							required
							value={form?.action === 'edit-promotion'
								? (form.promotionValues?.name ?? selectedPromotion?.name ?? '')
								: (selectedPromotion?.name ?? '')}
							class="theme-input"
						/>
					</div>
					<div>
						<label for="edit-promotion-type" class="theme-label">Type</label>
						<select id="edit-promotion-type" name="type" required class="theme-select">
							<option
								value="percentage"
								selected={form?.action === 'edit-promotion'
									? (form.promotionValues?.type ?? selectedPromotion?.type ?? 'percentage') ===
										'percentage'
									: selectedPromotion?.type === 'percentage'}
							>
								Pourcentage
							</option>
							<option
								value="fixed"
								selected={form?.action === 'edit-promotion'
									? (form.promotionValues?.type ?? selectedPromotion?.type ?? 'percentage') ===
										'fixed'
									: selectedPromotion?.type === 'fixed'}
							>
								Montant fixe
							</option>
						</select>
					</div>
					<div>
						<label for="edit-promotion-value" class="theme-label">Valeur</label>
						<input
							id="edit-promotion-value"
							name="value"
							type="number"
							min="0.01"
							step="0.01"
							required
							value={form?.action === 'edit-promotion'
								? (form.promotionValues?.value ?? String(selectedPromotion?.value ?? ''))
								: String(selectedPromotion?.value ?? '')}
							class="theme-input"
						/>
					</div>
				</div>

				<div>
					<label for="edit-promotion-description" class="theme-label">Description</label>
					<textarea
						id="edit-promotion-description"
						name="description"
						rows="4"
						class="theme-textarea"
						>{form?.action === 'edit-promotion'
							? (form.promotionValues?.description ?? selectedPromotion?.description ?? '')
							: (selectedPromotion?.description ?? '')}</textarea
					>
				</div>

				<div class="promotion-toggle-grid">
					<label class="checkbox-card">
						<input
							type="checkbox"
							name="is_active"
							value="true"
							bind:checked={editPromotionActive}
						/>
						<span class="theme-title text-sm font-medium">Promotion active</span>
					</label>
					<label class="checkbox-card">
						<input
							type="checkbox"
							name="applies_to_all"
							value="true"
							bind:checked={editPromotionGlobal}
						/>
						<span class="theme-title text-sm font-medium">Appliquer a tout le catalogue</span>
					</label>
				</div>

				<div class="theme-card p-4">
					<div class="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
						<div>
							<p class="theme-label">Produits cibles</p>
							<p class="theme-copy mt-1 text-sm">
								La selection est ignoree quand la promotion est globale.
							</p>
						</div>
						<span class="theme-pill">{selectedPromotion?.productCount ?? 0} cible(s)</span>
					</div>

					<div class="promotion-product-grid mt-4">
						{#each data.products as product (product.id)}
							<label class="promotion-product-card" class:disabled={editPromotionGlobal}>
								<input
									type="checkbox"
									name="product_ids"
									value={product.id}
									checked={form?.action === 'edit-promotion'
										? (form.promotionValues?.productIds ?? []).includes(String(product.id))
										: (selectedPromotion?.productIds ?? []).includes(product.id)}
									disabled={editPromotionGlobal}
								/>
								<div>
									<p class="theme-title text-sm font-bold">{product.name}</p>
									<p class="theme-copy mt-1 text-xs">{product.category}</p>
									<div class="mt-2">
										<ProductPrice {product} size="sm" showPromotionName={false} />
									</div>
								</div>
							</label>
						{/each}
					</div>
				</div>

				<div class="modal-actions">
					<button
						type="button"
						class="theme-button theme-button-secondary"
						onclick={closePromotionEditModal}
					>
						Annuler
					</button>
					<button type="submit" class="theme-button theme-button-primary">Enregistrer</button>
				</div>
			</form>
		</div>
	</div>
{/if}

{#if isDeletePromotionModalOpen && selectedPromotion}
	<div class="theme-overlay fixed inset-0 z-40 flex items-center justify-center px-4 py-8">
		<div class="theme-modal w-full max-w-md p-8">
			<div class="modal-header">
				<div>
					<p class="theme-kicker">Suppression</p>
					<h2 class="theme-title mt-3 text-2xl font-black">Supprimer la promotion</h2>
					<p class="theme-copy mt-2 text-sm">
						La remise ne sera plus appliquee aux produits concernes.
					</p>
				</div>
				<button
					type="button"
					class="theme-icon-button theme-close"
					onclick={closePromotionDeleteModal}
				>
					X
				</button>
			</div>

			<div class="danger-card mt-6">
				<p class="theme-title font-bold">{selectedPromotion.name}</p>
				<p class="theme-copy mt-2 text-sm">
					{selectedPromotion.description || 'Aucune description'}
				</p>
				<p class="theme-copy mt-3 text-sm">{formatPromotionScope(selectedPromotion)}</p>
				<p class="theme-title mt-4 font-semibold">{formatPromotionValue(selectedPromotion)}</p>
			</div>

			<form method="POST" action="?/deletePromotion" class="modal-actions mt-6">
				<input type="hidden" name="id" value={selectedPromotion.id} />
				<button
					type="button"
					class="theme-button theme-button-secondary"
					onclick={closePromotionDeleteModal}
				>
					Annuler
				</button>
				<button type="submit" class="theme-button theme-button-contrast">
					Confirmer la suppression
				</button>
			</form>
		</div>
	</div>
{/if}

<style>
	.admin-tabs {
		display: grid;
		gap: 0.85rem;
		width: 100%;
		max-width: 36rem;
	}

	.admin-tab {
		border-radius: 1.35rem;
		border: 1px solid var(--color-border);
		background: rgb(var(--color-white-rgb) / 0.76);
		padding: 1rem 1.1rem;
		text-align: left;
		color: var(--color-ink-muted);
		transition:
			transform 160ms ease,
			border-color var(--transition-standard),
			background-color var(--transition-standard),
			color var(--transition-standard);
	}

	.admin-tab:hover {
		transform: translateY(-1px);
		border-color: var(--color-border-strong);
		background: var(--color-white);
	}

	.admin-tab.active {
		border-color: rgb(var(--color-primary-rgb) / 0.32);
		background: var(--surface-muted);
		color: var(--color-primary);
		box-shadow: inset 0 1px 0 rgb(var(--color-white-rgb) / 0.55);
	}

	.table-toolbar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		padding: 1.6rem 1.75rem 0.75rem;
	}

	.row-media {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.action-row {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
	}

	.action-button {
		min-height: 2.65rem;
		padding: 0.65rem 1rem;
		font-size: 0.88rem;
	}

	.modal-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
	}

	.modal-actions {
		display: flex;
		justify-content: flex-end;
		flex-wrap: wrap;
		gap: 0.75rem;
		padding-top: 0.25rem;
	}

	.image-preview-card {
		display: grid;
		gap: 1rem;
	}

	.promotion-form-grid {
		display: grid;
		gap: 1rem;
	}

	.promotion-toggle-grid {
		display: grid;
		gap: 1rem;
	}

	.promotion-product-grid {
		display: grid;
		gap: 0.85rem;
		max-height: 20rem;
		overflow: auto;
	}

	.promotion-product-card {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr);
		align-items: start;
		gap: 0.85rem;
		border-radius: 1rem;
		border: 1px solid var(--color-border);
		background: rgb(var(--color-white-rgb) / 0.78);
		padding: 0.9rem 1rem;
	}

	.promotion-product-card.disabled {
		opacity: 0.55;
	}

	.promotion-product-card input {
		margin-top: 0.2rem;
		accent-color: var(--color-primary);
	}

	.checkbox-card {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		border-radius: 1rem;
		border: 1px solid var(--color-border);
		background: rgb(var(--color-white-rgb) / 0.86);
		padding: 0.9rem 1rem;
	}

	.checkbox-card input {
		accent-color: var(--color-primary);
	}

	.danger-card {
		border-radius: 1.25rem;
		border: 1px solid rgb(var(--color-primary-rgb) / 0.12);
		background: linear-gradient(
			135deg,
			rgb(var(--color-black-rgb) / 0.05) 0%,
			rgb(var(--color-white-rgb) / 0.9) 100%
		);
		padding: 1rem 1.1rem;
	}

	@media (min-width: 640px) {
		.admin-tabs {
			grid-template-columns: repeat(4, minmax(0, 1fr));
		}

		.image-preview-card {
			grid-template-columns: 120px minmax(0, 1fr);
			align-items: start;
		}

		.promotion-form-grid {
			grid-template-columns: repeat(3, minmax(0, 1fr));
		}

		.promotion-toggle-grid {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}

		.promotion-product-grid {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}
</style>
