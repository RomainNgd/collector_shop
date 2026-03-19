<script lang="ts">
	import type { Category, Product } from '$lib/types';
	import type { ActionData, PageData } from './$types';

	type AdminSection = 'dashboard' | 'products' | 'categories';

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

	const categoryInsights = $derived.by(() => {
		const counts = new Map<string, number>();
		for (const product of data.products) {
			const key = product.category.trim() || 'non-classe';
			counts.set(key, (counts.get(key) ?? 0) + 1);
		}

		return [...counts.entries()]
			.map(([name, count]) => ({ name, count }))
			.sort((a, b) => b.count - a.count)
			.slice(0, 4);
	});

	const findProductById = (id: string | undefined) => {
		const numericId = Number(id);
		return Number.isFinite(numericId)
			? data.products.find((product) => product.id === numericId) ?? null
			: null;
	};

	const findCategoryById = (id: string | undefined) => {
		const numericId = Number(id);
		return Number.isFinite(numericId)
			? data.categories.find((category) => category.id === numericId) ?? null
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
			selectedCategory = findCategoryById(form.categoryValues?.id ?? form.categoryId) ?? selectedCategory;
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

	const closeCategoryEditModal = () => {
		isEditCategoryModalOpen = false;
		selectedCategory = null;
	};
</script>

<section class="space-y-6">
	{#if form?.success}
		<div class="rounded-3xl border border-emerald-200 bg-linear-to-r from-emerald-50 via-white to-emerald-100 px-6 py-4 shadow-sm">
			<p class="text-sm font-semibold uppercase tracking-[0.2em] text-emerald-700">Confirmation</p>
			<p class="mt-1 text-base font-bold text-emerald-950">{form.success}</p>
		</div>
	{/if}

	{#if form?.error}
		<div class="rounded-3xl border border-red-200 bg-linear-to-r from-red-50 via-white to-red-100 px-6 py-4 shadow-sm">
			<p class="text-sm font-semibold uppercase tracking-[0.2em] text-red-700">Erreur</p>
			<p class="mt-1 text-base font-bold text-red-950">{form.error}</p>
		</div>
	{/if}

	<div class="rounded-3xl border border-white/70 bg-white/90 p-8 shadow-sm">
		<div class="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
			<div>
				<p class="text-sm font-semibold uppercase tracking-[0.22em] text-blue-700">Admin Center</p>
				<h1 class="mt-2 text-4xl font-black text-slate-900">Administration du catalogue</h1>
				<p class="mt-3 max-w-2xl text-slate-600">
					Un espace unique pour piloter les produits, les categories et preparer un futur tableau
					de bord ventes.
				</p>
			</div>

			<div class="grid gap-3 sm:grid-cols-3">
				<button type="button" class="rounded-2xl border px-4 py-3 text-left transition"
					class:border-blue-500={activeSection === 'dashboard'}
					class:bg-blue-50={activeSection === 'dashboard'}
					class:text-blue-900={activeSection === 'dashboard'}
					class:border-slate-200={activeSection !== 'dashboard'}
					class:bg-white={activeSection !== 'dashboard'}
					class:text-slate-700={activeSection !== 'dashboard'}
					onclick={() => (activeSection = 'dashboard')}>
					<p class="text-sm font-black">Dashboard</p>
					<p class="mt-1 text-xs opacity-80">Vue rapide et ventes</p>
				</button>
				<button type="button" class="rounded-2xl border px-4 py-3 text-left transition"
					class:border-blue-500={activeSection === 'products'}
					class:bg-blue-50={activeSection === 'products'}
					class:text-blue-900={activeSection === 'products'}
					class:border-slate-200={activeSection !== 'products'}
					class:bg-white={activeSection !== 'products'}
					class:text-slate-700={activeSection !== 'products'}
					onclick={() => (activeSection = 'products')}>
					<p class="text-sm font-black">Produits</p>
					<p class="mt-1 text-xs opacity-80">Catalogue et images</p>
				</button>
				<button type="button" class="rounded-2xl border px-4 py-3 text-left transition"
					class:border-blue-500={activeSection === 'categories'}
					class:bg-blue-50={activeSection === 'categories'}
					class:text-blue-900={activeSection === 'categories'}
					class:border-slate-200={activeSection !== 'categories'}
					class:bg-white={activeSection !== 'categories'}
					class:text-slate-700={activeSection !== 'categories'}
					onclick={() => (activeSection = 'categories')}>
					<p class="text-sm font-black">Categories</p>
					<p class="mt-1 text-xs opacity-80">Taxonomie du shop</p>
				</button>
			</div>
		</div>
	</div>

	{#if activeSection === 'dashboard'}
		<section class="space-y-6">
			<div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
				<div class="rounded-3xl border border-white/70 bg-white/95 p-6 shadow-sm"><p class="text-sm font-semibold uppercase tracking-[0.18em] text-slate-500">Produits</p><p class="mt-3 text-4xl font-black text-slate-900">{dashboardMetrics.productsCount}</p><p class="mt-2 text-sm text-slate-500">Articles visibles dans le catalogue</p></div>
				<div class="rounded-3xl border border-white/70 bg-white/95 p-6 shadow-sm"><p class="text-sm font-semibold uppercase tracking-[0.18em] text-slate-500">Categories</p><p class="mt-3 text-4xl font-black text-slate-900">{dashboardMetrics.categoriesCount}</p><p class="mt-2 text-sm text-slate-500">Axes de classement disponibles</p></div>
				<div class="rounded-3xl border border-white/70 bg-white/95 p-6 shadow-sm"><p class="text-sm font-semibold uppercase tracking-[0.18em] text-slate-500">Valeur catalogue</p><p class="mt-3 text-4xl font-black text-slate-900">{dashboardMetrics.totalCatalogValue.toFixed(2)} EUR</p><p class="mt-2 text-sm text-slate-500">Somme des prix affiches</p></div>
				<div class="rounded-3xl border border-white/70 bg-white/95 p-6 shadow-sm"><p class="text-sm font-semibold uppercase tracking-[0.18em] text-slate-500">A classer</p><p class="mt-3 text-4xl font-black text-slate-900">{dashboardMetrics.uncategorizedCount}</p><p class="mt-2 text-sm text-slate-500">Produits encore non classes</p></div>
			</div>

			<div class="grid gap-6 xl:grid-cols-[1.5fr_1fr]">
				<div class="rounded-3xl border border-white/70 bg-white/95 p-6 shadow-sm">
					<div class="flex items-start justify-between gap-4">
						<div>
							<h2 class="text-2xl font-black text-slate-900">Vue ventes</h2>
							<p class="mt-2 text-sm text-slate-500">Bloc pret pour accueillir les stats de commandes quand l'API ventes sera disponible.</p>
						</div>
						<span class="rounded-full border border-amber-200 bg-amber-50 px-3 py-1 text-xs font-bold uppercase tracking-[0.16em] text-amber-700">Bientot</span>
					</div>
				</div>

				<div class="rounded-3xl border border-white/70 bg-white/95 p-6 shadow-sm">
					<h2 class="text-2xl font-black text-slate-900">Categories les plus remplies</h2>
					<div class="mt-5 space-y-3">
						{#each categoryInsights as insight (insight.name)}
							<div class="flex items-center justify-between rounded-2xl bg-slate-50 px-4 py-3">
								<p class="font-semibold text-slate-800">{insight.name}</p>
								<span class="rounded-full bg-slate-900 px-3 py-1 text-xs font-bold text-white">{insight.count} produit{insight.count > 1 ? 's' : ''}</span>
							</div>
						{/each}
					</div>
				</div>
			</div>
		</section>
	{/if}

	{#if activeSection === 'products'}
		<section class="space-y-6">
			<div class="rounded-3xl border border-white/70 bg-white/90 p-8 shadow-sm">
				<div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
					<div>
						<h2 class="text-3xl font-black text-slate-900">Administration des produits</h2>
						<p class="mt-2 text-slate-600">Gestion complete du catalogue et des images.</p>
					</div>
					<button type="button" class="rounded-xl bg-blue-600 px-4 py-2.5 font-bold text-white transition hover:bg-blue-700" onclick={openCreateProductModal}>Ajouter un produit</button>
				</div>
			</div>

			<div class="overflow-hidden rounded-3xl border border-white/70 bg-white/95 shadow-sm">
				<div class="flex items-center justify-between border-b border-slate-200 px-6 py-4">
					<div>
						<h3 class="text-2xl font-black text-slate-900">Liste des produits</h3>
						<p class="mt-1 text-sm text-slate-500">{data.products.length} produits</p>
					</div>
				</div>

				<div class="overflow-x-auto">
					<table class="min-w-full divide-y divide-slate-200">
						<thead class="bg-slate-50">
							<tr class="text-left text-sm font-semibold text-slate-600">
								<th class="px-6 py-4">Produit</th>
								<th class="px-6 py-4">Description</th>
								<th class="px-6 py-4">Categorie</th>
								<th class="px-6 py-4">Image</th>
								<th class="px-6 py-4">Prix</th>
								<th class="px-6 py-4">Actions</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-slate-100 bg-white">
							{#each data.products as product (product.id)}
								<tr class="align-top">
									<td class="px-6 py-4"><div class="flex items-center gap-4"><img src={product.imageUrl} alt={product.name} class="h-16 w-16 rounded-2xl object-cover" /><div><p class="font-bold text-slate-900">{product.name}</p><p class="text-sm text-slate-500">ID #{product.id}</p></div></div></td>
									<td class="px-6 py-4 text-sm text-slate-600">{product.description}</td>
									<td class="px-6 py-4 text-sm text-slate-600">{product.category}</td>
									<td class="px-6 py-4 text-sm text-slate-600">{product.imageName ?? 'Aucune image'}</td>
									<td class="px-6 py-4 font-semibold text-slate-900">{product.price.toFixed(2)} EUR</td>
									<td class="px-6 py-4"><div class="flex gap-3"><button type="button" class="rounded-xl border border-slate-300 px-4 py-2 text-sm font-semibold text-slate-700 transition hover:border-slate-400 hover:bg-slate-50" onclick={() => openEditProductModal(product)}>Modifier</button><button type="button" class="rounded-xl border border-red-200 px-4 py-2 text-sm font-semibold text-red-600 transition hover:bg-red-50" onclick={() => openDeleteProductModal(product)}>Supprimer</button></div></td>
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
			<div class="rounded-3xl border border-white/70 bg-white/90 p-8 shadow-sm">
				<div class="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
					<div>
						<h2 class="text-3xl font-black text-slate-900">Administration des categories</h2>
						<p class="mt-2 text-slate-600">Organise la taxonomie pour rendre le catalogue plus clair.</p>
					</div>
					<button type="button" class="rounded-xl bg-blue-600 px-4 py-2.5 font-bold text-white transition hover:bg-blue-700" onclick={openCreateCategoryModal}>Ajouter une categorie</button>
				</div>
			</div>

			<div class="overflow-hidden rounded-3xl border border-white/70 bg-white/95 shadow-sm">
				<div class="flex items-center justify-between border-b border-slate-200 px-6 py-4">
					<div>
						<h3 class="text-2xl font-black text-slate-900">Liste des categories</h3>
						<p class="mt-1 text-sm text-slate-500">{data.categories.length} categories</p>
					</div>
				</div>

				<div class="overflow-x-auto">
					<table class="min-w-full divide-y divide-slate-200">
						<thead class="bg-slate-50">
							<tr class="text-left text-sm font-semibold text-slate-600">
								<th class="px-6 py-4">Nom</th>
								<th class="px-6 py-4">Produits lies</th>
								<th class="px-6 py-4">Actions</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-slate-100 bg-white">
							{#each data.categories as category (category.id)}
								<tr>
									<td class="px-6 py-4"><p class="font-bold text-slate-900">{category.name}</p><p class="text-sm text-slate-500">ID #{category.id}</p><p class="mt-2 text-sm text-slate-500">{category.description || 'Aucune description'}</p></td>
									<td class="px-6 py-4 text-sm text-slate-600">{data.products.filter((product) => product.category.trim().toLowerCase() === category.name.trim().toLowerCase()).length}</td>
									<td class="px-6 py-4"><div class="flex gap-3"><button type="button" class="rounded-xl border border-slate-300 px-4 py-2 text-sm font-semibold text-slate-700 transition hover:border-slate-400 hover:bg-slate-50" onclick={() => openEditCategoryModal(category)}>Modifier</button><button type="button" class="rounded-xl border border-red-200 px-4 py-2 text-sm font-semibold text-red-600 transition hover:bg-red-50" onclick={() => openDeleteCategoryModal(category)}>Supprimer</button></div></td>
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
	<div class="fixed inset-0 z-40 flex items-center justify-center bg-slate-950/45 px-4 py-8"><div class="w-full max-w-xl rounded-3xl border border-white/70 bg-white p-8 shadow-2xl"><div class="flex items-start justify-between gap-4"><div><h2 class="text-2xl font-black text-slate-900">Ajouter un produit</h2><p class="mt-2 text-sm text-slate-500">Le produit sera cree puis l'image sera envoyee si un fichier est selectionne.</p></div><button type="button" class="rounded-full p-2 text-slate-500 transition hover:bg-slate-100 hover:text-slate-700" onclick={() => (isCreateProductModalOpen = false)}>X</button></div><form method="POST" action="?/createProduct" enctype="multipart/form-data" class="mt-6 space-y-4"><div><label for="create-name" class="mb-1 block text-sm font-semibold text-slate-700">Nom</label><input id="create-name" name="name" type="text" required value={form?.action === 'create-product' ? form.values?.name ?? '' : ''} class="w-full rounded-xl border border-slate-300 px-3 py-2 focus:border-blue-500 focus:outline-none" /></div><div><label for="create-price" class="mb-1 block text-sm font-semibold text-slate-700">Prix</label><input id="create-price" name="price" type="number" min="0" step="0.01" required value={form?.action === 'create-product' ? form.values?.price ?? '' : ''} class="w-full rounded-xl border border-slate-300 px-3 py-2 focus:border-blue-500 focus:outline-none" /></div><div><label for="create-category" class="mb-1 block text-sm font-semibold text-slate-700">Categorie</label><select id="create-category" name="category_id" required class="w-full rounded-xl border border-slate-300 px-3 py-2 focus:border-blue-500 focus:outline-none"><option value="">Selectionner une categorie</option>{#each data.categories as category (category.id)}<option value={category.id} selected={form?.action === 'create-product' ? form.values?.categoryId === String(category.id) : false}>{category.name}</option>{/each}</select></div><div><label for="create-description" class="mb-1 block text-sm font-semibold text-slate-700">Description</label><textarea id="create-description" name="description" rows="4" required class="w-full rounded-xl border border-slate-300 px-3 py-2 focus:border-blue-500 focus:outline-none">{form?.action === 'create-product' ? form.values?.description ?? '' : ''}</textarea></div><div class="rounded-2xl border border-slate-200 bg-slate-50/80 p-4"><label for="create-image" class="mb-2 block text-sm font-semibold text-slate-700">Image</label><input id="create-image" name="image" type="file" accept="image/*" class="block w-full text-sm text-slate-600 file:mr-4 file:rounded-xl file:border-0 file:bg-slate-900 file:px-4 file:py-2 file:font-semibold file:text-white hover:file:bg-slate-700" /></div><div class="flex justify-end gap-3 pt-2"><button type="button" class="rounded-xl border border-slate-300 px-4 py-2.5 font-semibold text-slate-700 transition hover:bg-slate-50" onclick={() => (isCreateProductModalOpen = false)}>Annuler</button><button type="submit" class="rounded-xl bg-blue-600 px-4 py-2.5 font-bold text-white transition hover:bg-blue-700">Valider</button></div></form></div></div>
{/if}

{#if isEditProductModalOpen}
	<div class="fixed inset-0 z-40 flex items-center justify-center bg-slate-950/45 px-4 py-8"><div class="w-full max-w-xl rounded-3xl border border-white/70 bg-white p-8 shadow-2xl"><div class="flex items-start justify-between gap-4"><div><h2 class="text-2xl font-black text-slate-900">Modifier un produit</h2><p class="mt-2 text-sm text-slate-500">Ajoute une nouvelle image pour la remplacer, ou coche la suppression si besoin.</p></div><button type="button" class="rounded-full p-2 text-slate-500 transition hover:bg-slate-100 hover:text-slate-700" onclick={closeProductEditModal}>X</button></div><form method="POST" action="?/updateProduct" enctype="multipart/form-data" class="mt-6 space-y-4"><input type="hidden" name="id" value={form?.action === 'edit-product' ? form.values?.id ?? selectedProduct?.id ?? '' : selectedProduct?.id ?? ''} /><input type="hidden" name="currentImageName" value={form?.action === 'edit-product' ? form.values?.currentImageName ?? selectedProduct?.imageName ?? '' : selectedProduct?.imageName ?? ''} /><div><label for="edit-name" class="mb-1 block text-sm font-semibold text-slate-700">Nom</label><input id="edit-name" name="name" type="text" required value={form?.action === 'edit-product' ? form.values?.name ?? selectedProduct?.name ?? '' : selectedProduct?.name ?? ''} class="w-full rounded-xl border border-slate-300 px-3 py-2 focus:border-blue-500 focus:outline-none" /></div><div><label for="edit-price" class="mb-1 block text-sm font-semibold text-slate-700">Prix</label><input id="edit-price" name="price" type="number" min="0" step="0.01" required value={form?.action === 'edit-product' ? form.values?.price ?? String(selectedProduct?.price ?? '') : String(selectedProduct?.price ?? '')} class="w-full rounded-xl border border-slate-300 px-3 py-2 focus:border-blue-500 focus:outline-none" /></div><div><label for="edit-category" class="mb-1 block text-sm font-semibold text-slate-700">Categorie</label><select id="edit-category" name="category_id" required class="w-full rounded-xl border border-slate-300 px-3 py-2 focus:border-blue-500 focus:outline-none"><option value="">Selectionner une categorie</option>{#each data.categories as category (category.id)}<option value={category.id} selected={form?.action === 'edit-product' ? form.values?.categoryId === String(category.id) : getCategoryIdFromProduct(selectedProduct) === String(category.id)}>{category.name}</option>{/each}</select></div><div><label for="edit-description" class="mb-1 block text-sm font-semibold text-slate-700">Description</label><textarea id="edit-description" name="description" rows="4" required class="w-full rounded-xl border border-slate-300 px-3 py-2 focus:border-blue-500 focus:outline-none">{form?.action === 'edit-product' ? form.values?.description ?? selectedProduct?.description ?? '' : selectedProduct?.description ?? ''}</textarea></div><div class="grid gap-4 rounded-2xl border border-slate-200 bg-slate-50/80 p-4 md:grid-cols-[120px_1fr]"><div><p class="mb-2 text-sm font-semibold text-slate-700">Image actuelle</p><img src={selectedProduct?.imageUrl} alt={selectedProduct?.name ?? 'Produit'} class="h-28 w-28 rounded-2xl object-cover" /></div><div class="space-y-4"><div><p class="text-sm font-semibold text-slate-700">{selectedProduct?.imageName ?? 'Aucune image enregistree'}</p><p class="mt-1 text-xs text-slate-500">L'image est servie depuis <code>/upload/&lt;filename&gt;</code>.</p></div><div><label for="edit-image" class="mb-2 block text-sm font-semibold text-slate-700">Remplacer l'image</label><input id="edit-image" name="image" type="file" accept="image/*" class="block w-full text-sm text-slate-600 file:mr-4 file:rounded-xl file:border-0 file:bg-slate-900 file:px-4 file:py-2 file:font-semibold file:text-white hover:file:bg-slate-700" /></div><label class="flex items-center gap-3 rounded-xl border border-slate-200 bg-white px-4 py-3"><input type="checkbox" name="removeImage" value="true" checked={form?.action === 'edit-product' ? form.values?.removeImage === 'true' : false} class="h-4 w-4 rounded border-slate-300 text-red-600 focus:ring-red-500" /><span class="text-sm font-medium text-slate-700">Supprimer l'image actuelle</span></label></div></div><div class="flex justify-end gap-3 pt-2"><button type="button" class="rounded-xl border border-slate-300 px-4 py-2.5 font-semibold text-slate-700 transition hover:bg-slate-50" onclick={closeProductEditModal}>Annuler</button><button type="submit" class="rounded-xl bg-amber-500 px-4 py-2.5 font-bold text-white transition hover:bg-amber-600">Enregistrer</button></div></form></div></div>
{/if}

{#if isDeleteProductModalOpen && selectedProduct}
	<div class="fixed inset-0 z-40 flex items-center justify-center bg-slate-950/45 px-4 py-8"><div class="w-full max-w-md rounded-3xl border border-white/70 bg-white p-8 shadow-2xl"><div class="flex items-start justify-between gap-4"><div><h2 class="text-2xl font-black text-slate-900">Confirmer la suppression</h2><p class="mt-2 text-sm text-slate-500">Cette action supprimera definitivement le produit selectionne.</p></div><button type="button" class="rounded-full p-2 text-slate-500 transition hover:bg-slate-100 hover:text-slate-700" onclick={() => { isDeleteProductModalOpen = false; selectedProduct = null; }}>X</button></div><div class="mt-6 rounded-2xl border border-red-100 bg-red-50/70 p-4"><p class="font-bold text-slate-900">{selectedProduct.name}</p><p class="mt-1 text-sm text-slate-600">{selectedProduct.description}</p><p class="mt-3 text-sm text-slate-600">Image: {selectedProduct.imageName ?? 'Aucune image'}</p><p class="mt-3 text-sm font-semibold text-red-700">ID #{selectedProduct.id} - {selectedProduct.price.toFixed(2)} EUR</p></div><form method="POST" action="?/deleteProduct" class="mt-6 flex justify-end gap-3"><input type="hidden" name="id" value={selectedProduct.id} /><button type="button" class="rounded-xl border border-slate-300 px-4 py-2.5 font-semibold text-slate-700 transition hover:bg-slate-50" onclick={() => { isDeleteProductModalOpen = false; selectedProduct = null; }}>Annuler</button><button type="submit" class="rounded-xl bg-red-600 px-4 py-2.5 font-bold text-white transition hover:bg-red-700">Confirmer la suppression</button></form></div></div>
{/if}

{#if isCreateCategoryModalOpen}
	<div class="fixed inset-0 z-40 flex items-center justify-center bg-slate-950/45 px-4 py-8"><div class="w-full max-w-lg rounded-3xl border border-white/70 bg-white p-8 shadow-2xl"><div class="flex items-start justify-between gap-4"><div><h2 class="text-2xl font-black text-slate-900">Ajouter une categorie</h2><p class="mt-2 text-sm text-slate-500">Une categorie claire aide a mieux organiser le catalogue.</p></div><button type="button" class="rounded-full p-2 text-slate-500 transition hover:bg-slate-100 hover:text-slate-700" onclick={() => (isCreateCategoryModalOpen = false)}>X</button></div><form method="POST" action="?/createCategory" class="mt-6 space-y-4"><div><label for="create-category-name" class="mb-1 block text-sm font-semibold text-slate-700">Nom</label><input id="create-category-name" name="name" type="text" required value={form?.action === 'create-category' ? form.categoryValues?.name ?? '' : ''} class="w-full rounded-xl border border-slate-300 px-3 py-2 focus:border-blue-500 focus:outline-none" /></div><div><label for="create-category-description" class="mb-1 block text-sm font-semibold text-slate-700">Description</label><textarea id="create-category-description" name="description" rows="4" class="w-full rounded-xl border border-slate-300 px-3 py-2 focus:border-blue-500 focus:outline-none">{form?.action === 'create-category' ? form.categoryValues?.description ?? '' : ''}</textarea></div><div class="flex justify-end gap-3 pt-2"><button type="button" class="rounded-xl border border-slate-300 px-4 py-2.5 font-semibold text-slate-700 transition hover:bg-slate-50" onclick={() => (isCreateCategoryModalOpen = false)}>Annuler</button><button type="submit" class="rounded-xl bg-blue-600 px-4 py-2.5 font-bold text-white transition hover:bg-blue-700">Valider</button></div></form></div></div>
{/if}

{#if isEditCategoryModalOpen}
	<div class="fixed inset-0 z-40 flex items-center justify-center bg-slate-950/45 px-4 py-8"><div class="w-full max-w-lg rounded-3xl border border-white/70 bg-white p-8 shadow-2xl"><div class="flex items-start justify-between gap-4"><div><h2 class="text-2xl font-black text-slate-900">Modifier une categorie</h2><p class="mt-2 text-sm text-slate-500">Renomme la categorie pour harmoniser le catalogue.</p></div><button type="button" class="rounded-full p-2 text-slate-500 transition hover:bg-slate-100 hover:text-slate-700" onclick={closeCategoryEditModal}>X</button></div><form method="POST" action="?/updateCategory" class="mt-6 space-y-4"><input type="hidden" name="id" value={form?.action === 'edit-category' ? form.categoryValues?.id ?? selectedCategory?.id ?? '' : selectedCategory?.id ?? ''} /><div><label for="edit-category-name" class="mb-1 block text-sm font-semibold text-slate-700">Nom</label><input id="edit-category-name" name="name" type="text" required value={form?.action === 'edit-category' ? form.categoryValues?.name ?? selectedCategory?.name ?? '' : selectedCategory?.name ?? ''} class="w-full rounded-xl border border-slate-300 px-3 py-2 focus:border-blue-500 focus:outline-none" /></div><div><label for="edit-category-description" class="mb-1 block text-sm font-semibold text-slate-700">Description</label><textarea id="edit-category-description" name="description" rows="4" class="w-full rounded-xl border border-slate-300 px-3 py-2 focus:border-blue-500 focus:outline-none">{form?.action === 'edit-category' ? form.categoryValues?.description ?? selectedCategory?.description ?? '' : selectedCategory?.description ?? ''}</textarea></div><div class="flex justify-end gap-3 pt-2"><button type="button" class="rounded-xl border border-slate-300 px-4 py-2.5 font-semibold text-slate-700 transition hover:bg-slate-50" onclick={closeCategoryEditModal}>Annuler</button><button type="submit" class="rounded-xl bg-amber-500 px-4 py-2.5 font-bold text-white transition hover:bg-amber-600">Enregistrer</button></div></form></div></div>
{/if}

{#if isDeleteCategoryModalOpen && selectedCategory}
	<div class="fixed inset-0 z-40 flex items-center justify-center bg-slate-950/45 px-4 py-8"><div class="w-full max-w-md rounded-3xl border border-white/70 bg-white p-8 shadow-2xl"><div class="flex items-start justify-between gap-4"><div><h2 class="text-2xl font-black text-slate-900">Supprimer la categorie</h2><p class="mt-2 text-sm text-slate-500">Confirme la suppression avant de retirer cette categorie.</p></div><button type="button" class="rounded-full p-2 text-slate-500 transition hover:bg-slate-100 hover:text-slate-700" onclick={() => { isDeleteCategoryModalOpen = false; selectedCategory = null; }}>X</button></div><div class="mt-6 rounded-2xl border border-red-100 bg-red-50/70 p-4"><p class="font-bold text-slate-900">{selectedCategory.name}</p><p class="mt-2 text-sm text-red-700">ID #{selectedCategory.id}</p></div><form method="POST" action="?/deleteCategory" class="mt-6 flex justify-end gap-3"><input type="hidden" name="id" value={selectedCategory.id} /><button type="button" class="rounded-xl border border-slate-300 px-4 py-2.5 font-semibold text-slate-700 transition hover:bg-slate-50" onclick={() => { isDeleteCategoryModalOpen = false; selectedCategory = null; }}>Annuler</button><button type="submit" class="rounded-xl bg-red-600 px-4 py-2.5 font-bold text-white transition hover:bg-red-700">Confirmer la suppression</button></form></div></div>
{/if}
