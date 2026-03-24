<script lang="ts">
	import { resolve } from '$app/paths';
	import {
		cartItems,
		cartTotal,
		clearCart,
		removeFromCart,
		updateQuantity
	} from '$lib/stores/cart';
</script>

<section class="space-y-8">
	<div class="theme-section-heading">
		<div>
			<p class="theme-kicker">Panier</p>
			<h1 class="theme-title mt-3 text-4xl font-black">Ta selection du moment</h1>
			<p class="theme-copy mt-3 max-w-2xl">Retrouve ici les objets mis de cote avant validation.</p>
		</div>
		<span class="theme-pill theme-pill-contrast">
			{$cartItems.length} ligne{$cartItems.length > 1 ? 's' : ''}
		</span>
	</div>

	{#if $cartItems.length === 0}
		<div class="theme-empty theme-panel">
			<p class="theme-title text-lg font-bold">Votre panier est vide.</p>
			<p class="theme-copy mt-3">Retourne sur le catalogue pour ajouter de nouvelles pieces.</p>
			<div class="mt-6 flex justify-center">
				<a href={resolve('/catalogue')} class="theme-button theme-button-primary">
					Voir le catalogue
				</a>
			</div>
		</div>
	{:else}
		<div class="grid gap-8 lg:grid-cols-[1.5fr_0.85fr]">
			<div class="space-y-4">
				{#each $cartItems as item (item.product.id)}
					<article
						class="theme-card theme-hover-lift flex flex-col gap-4 p-4 sm:flex-row sm:items-center"
					>
						<div class="item-image-shell">
							<img
								src={item.product.imageUrl}
								alt={item.product.name}
								class="item-image h-24 w-24 rounded-2xl object-cover"
							/>
						</div>

						<div class="min-w-0 flex-1">
							<p class="theme-kicker">{item.product.category}</p>
							<h2 class="theme-title mt-2 line-clamp-1 text-xl font-black">
								{item.product.name}
							</h2>
							<p class="theme-copy mt-1 text-sm">{item.product.price} EUR unite</p>
						</div>

						<div class="quantity-control">
							<button
								type="button"
								onclick={() => updateQuantity(item.product.id, item.quantity - 1)}
								class="quantity-btn"
								aria-label="Retirer une unite"
							>
								-
							</button>
							<span class="quantity-value">{item.quantity}</span>
							<button
								type="button"
								onclick={() => updateQuantity(item.product.id, item.quantity + 1)}
								class="quantity-btn"
								aria-label="Ajouter une unite"
							>
								+
							</button>
						</div>

						<div class="min-w-28 text-right">
							<p class="theme-copy text-xs">Sous-total</p>
							<p class="theme-title mt-1 text-lg font-black">
								{(item.product.price * item.quantity).toFixed(2)} EUR
							</p>
						</div>

						<button
							type="button"
							onclick={() => removeFromCart(item.product.id)}
							class="theme-button theme-button-secondary remove-btn"
						>
							Retirer
						</button>
					</article>
				{/each}
			</div>

			<aside class="theme-panel p-6 md:p-8">
				<p class="theme-kicker">Recapitulatif</p>
				<h2 class="theme-title mt-3 text-3xl font-black">Panier pret</h2>
				<p class="theme-copy mt-3">Consulte ton total et finalise ta selection.</p>

				<div class="summary-grid mt-8">
					<div class="summary-row">
						<span class="theme-copy">Articles</span>
						<span class="theme-title font-bold">{$cartItems.length}</span>
					</div>
					<div class="summary-row">
						<span class="theme-copy">Total</span>
						<span class="theme-price text-3xl font-black">{$cartTotal.toFixed(2)} EUR</span>
					</div>
				</div>

				<div class="mt-8 flex flex-col gap-3">
					<button type="button" class="theme-button theme-button-primary w-full justify-center">
						Commander
					</button>
					<button
						type="button"
						onclick={clearCart}
						class="theme-button theme-button-contrast w-full justify-center"
					>
						Vider le panier
					</button>
				</div>
			</aside>
		</div>
	{/if}
</section>

<style>
	.item-image-shell {
		overflow: hidden;
		border-radius: 1.35rem;
		background: linear-gradient(
			145deg,
			rgb(var(--color-white-rgb) / 0.96),
			rgb(var(--color-primary-rgb) / 0.04)
		);
		padding: 0.45rem;
	}

	.item-image {
		border-radius: 1rem;
	}

	.quantity-control {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		border-radius: 999px;
		border: 1px solid var(--color-border);
		background: rgb(var(--color-white-rgb) / 0.72);
		padding: 0.35rem;
	}

	.quantity-btn {
		display: inline-grid;
		place-items: center;
		height: 2.2rem;
		width: 2.2rem;
		border: 0;
		border-radius: 999px;
		background: var(--surface-muted);
		color: var(--color-primary);
		font-size: 1.1rem;
		font-weight: 900;
		transition:
			transform 160ms ease,
			background-color var(--transition-standard);
	}

	.quantity-btn:hover {
		transform: translateY(-1px);
		background: var(--surface-muted-strong);
	}

	.quantity-value {
		min-width: 1.5rem;
		text-align: center;
		font-weight: 800;
		color: var(--color-black);
	}

	.remove-btn {
		min-width: 8rem;
	}

	.summary-grid {
		display: grid;
		gap: 1rem;
	}

	.summary-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		padding: 1rem 0;
		border-top: 1px solid rgb(var(--color-primary-rgb) / 0.08);
	}

	.summary-row:first-child {
		border-top: 0;
		padding-top: 0;
	}
</style>
