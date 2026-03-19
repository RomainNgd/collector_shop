<script lang="ts">
	import {
		cartItems,
		cartTotal,
		clearCart,
		removeFromCart,
		updateQuantity
	} from '$lib/stores/cart';
</script>

<section class="mx-auto max-w-4xl py-8">
	<h1 class="mb-6 text-3xl font-extrabold text-gray-900">Panier</h1>

	{#if $cartItems.length === 0}
		<div
			class="rounded-xl border border-dashed border-gray-300 bg-white p-8 text-center text-gray-500"
		>
			Votre panier est vide.
		</div>
	{:else}
		<div class="space-y-4">
			{#each $cartItems as item (item.product.id)}
				<article
					class="flex flex-col gap-4 rounded-xl border bg-white p-4 sm:flex-row sm:items-center"
				>
					<img
						src={item.product.imageUrl}
						alt={item.product.name}
						class="h-20 w-20 rounded-md object-cover"
					/>

					<div class="flex-1">
						<h2 class="font-bold text-gray-900">{item.product.name}</h2>
						<p class="text-sm text-gray-500">{item.product.price} EUR unite</p>
					</div>

					<div class="flex items-center gap-2">
						<button
							type="button"
							onclick={() => updateQuantity(item.product.id, item.quantity - 1)}
							class="h-8 w-8 rounded border hover:bg-gray-50"
							aria-label="Retirer une unite"
						>
							-
						</button>
						<span class="w-8 text-center">{item.quantity}</span>
						<button
							type="button"
							onclick={() => updateQuantity(item.product.id, item.quantity + 1)}
							class="h-8 w-8 rounded border hover:bg-gray-50"
							aria-label="Ajouter une unite"
						>
							+
						</button>
					</div>

					<div class="w-24 text-right font-semibold text-gray-900">
						{(item.product.price * item.quantity).toFixed(2)} EUR
					</div>

					<button
						type="button"
						onclick={() => removeFromCart(item.product.id)}
						class="rounded border border-red-200 px-3 py-1 text-sm text-red-600 hover:bg-red-50"
					>
						Retirer
					</button>
				</article>
			{/each}
		</div>

		<div class="mt-8 rounded-xl border bg-white p-6">
			<div class="mb-4 flex items-center justify-between text-lg">
				<span class="font-medium text-gray-600">Total</span>
				<span class="text-2xl font-extrabold text-gray-900">{$cartTotal.toFixed(2)} EUR</span>
			</div>
			<div class="flex flex-col gap-3 sm:flex-row">
				<button
					type="button"
					class="bg-brand-primary rounded-lg px-6 py-3 font-semibold text-white hover:bg-blue-700"
				>
					Commander
				</button>
				<button
					type="button"
					onclick={clearCart}
					class="rounded-lg border px-6 py-3 font-semibold text-gray-700 hover:bg-gray-50"
				>
					Vider le panier
				</button>
			</div>
		</div>
	{/if}
</section>
