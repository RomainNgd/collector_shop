<script lang="ts">
	import type { Category, Product } from '$lib/types';
	import type { ActionData, PageData } from './$types';

	type AdminSection = 'dashboard' | 'products' | 'categories';

	const adminSections = [
		{ id: 'dashboard' as const, title: 'Dashboard', description: 'Vue rapide et indicateurs' },
		{ id: 'products' as const, title: 'Produits', description: 'Catalogue et images' },
		{ id: 'categories' as const, title: 'Categories', description: 'Taxonomie du shop' }
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

	const dashboardMetrics = $derived.by(() => {
		const productsCount = data.products.length;
		const categoriesCount = data.categories.length;
		const totalCatalogValue = data.products.reduce((total, product) => total + product.price, 0);
		const uncategorizedCount = data.products.filter(
			(product) => product.category.trim().toLowerCase() === 'non-classe'
		).length;

		return { productsCount, categoriesCount, totalCatalogValue, uncategorizedCount };
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
		for (const product of data.products) {
			const key = product.category.trim() || 'non-classe';
			counts[key] = (counts[key] ?? 0) + 1;
		}

		return Object.entries(counts)
			.map(([name, count]) => ({ name, count }))
			.sort((a, b) => b.count - a.count)
			.slice(0, 4);
	});

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

	const closeAllModals = () => {
		isCreateProductModalOpen = false;
		isEditProductModalOpen = false;
		isDeleteProductModalOpen = false;
		isCreateCategoryModalOpen = false;
		isEditCategoryModalOpen = false;
		isDeleteCategoryModalOpen = false;
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

		if (form?.success) {
			closeAllModals();
			selectedProduct = null;
			selectedCategory = null;
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
								Bloc pret pour accueillir les stats de commandes des que l'API ventes sera
								disponible.
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
									<td class="theme-title font-semibold">{product.price.toFixed(2)} EUR</td>
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
						min="0"
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
						min="0"
						step="0.01"
						required
						value={form?.action === 'edit-product'
							? (form.values?.price ?? String(selectedProduct?.price ?? ''))
							: String(selectedProduct?.price ?? '')}
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

	.insight-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		border-radius: 1.2rem;
		background: rgb(var(--color-white-rgb) / 0.78);
		padding: 0.9rem 1rem;
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
			grid-template-columns: repeat(3, minmax(0, 1fr));
		}

		.image-preview-card {
			grid-template-columns: 120px minmax(0, 1fr);
			align-items: start;
		}
	}
</style>
