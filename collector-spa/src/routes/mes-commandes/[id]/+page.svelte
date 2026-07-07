<script lang="ts">
	import { resolve } from '$app/paths';
	import { canOrderBePaid, getOrderStatusLabel, getOrderStatusTone } from '$lib/orders';
	import { clearCart } from '$lib/stores/cart';
	import { onMount } from 'svelte';
	import type { ActionData, PageData } from './$types';

	let { data, form }: { data: PageData; form: ActionData | null } = $props();

	const dateFormatter = new Intl.DateTimeFormat('fr-FR', {
		dateStyle: 'full',
		timeStyle: 'short'
	});

	const formatOrderDate = (value: string) => {
		const parsed = new Date(value);
		return Number.isNaN(parsed.getTime()) ? 'Date indisponible' : dateFormatter.format(parsed);
	};

	const showPaymentProcessingBanner = $derived(
		data.paymentFlow === 'processing' && canOrderBePaid(data.order.status)
	);
	const showPaymentSuccessBanner = $derived(
		data.paymentFlow === 'processing' && !canOrderBePaid(data.order.status)
	);
	const showPaymentCancelledBanner = $derived(data.paymentFlow === 'cancelled');

	onMount(() => {
		if (data.shouldClearCart) {
			clearCart();
		}
	});
</script>

<section class="space-y-8">
	<div class="theme-section-heading">
		<div>
			<p class="theme-kicker">Commande</p>
			<h1 class="theme-title mt-3 text-4xl font-black">Recapitulatif #{data.order.id}</h1>
			<p class="theme-copy mt-3 max-w-2xl">
				Commande creee le {formatOrderDate(data.order.createdAt)}. Le prix est maintenant fige sur
				cette commande.
			</p>
		</div>
		<a href={resolve('/mes-commandes')} class="theme-button theme-button-secondary">
			Retour a mes commandes
		</a>
	</div>

	<div class="grid gap-8 lg:grid-cols-[1.35fr_0.9fr]">
		<div class="space-y-4">
			<div class="theme-panel p-5 md:p-6">
				<div class="flex flex-wrap items-center justify-between gap-3">
					<div>
						<p class="theme-copy text-sm">Statut actuel</p>
						<h2 class="theme-title mt-2 text-2xl font-black">
							{getOrderStatusLabel(data.order.status)}
						</h2>
					</div>
					<span class={`theme-pill ${getOrderStatusTone(data.order.status)}`}>
						{getOrderStatusLabel(data.order.status)}
					</span>
				</div>
				<div class="status-steps mt-6">
					<div class:step-active={data.order.status === 'awaiting_payment'} class="status-step">
						1. En attente de paiement
					</div>
					<div class:step-active={data.order.status === 'preparation'} class="status-step">
						2. Preparation
					</div>
					<div class:step-active={data.order.status === 'shipping'} class="status-step">
						3. En cours de livraison
					</div>
					<div class:step-active={data.order.status === 'delivered'} class="status-step">
						4. Livree
					</div>
				</div>
			</div>

			<div class="space-y-4">
				{#each data.order.items as item (item.id)}
					<article
						class="theme-card theme-hover-lift flex flex-col gap-4 p-4 sm:flex-row sm:items-center"
					>
						<div class="item-image-shell">
							<img
								src={item.productImageUrl}
								alt={item.productName}
								class="item-image h-24 w-24 rounded-2xl object-cover"
							/>
						</div>

						<div class="min-w-0 flex-1">
							<p class="theme-kicker">{item.categoryName}</p>
							<h2 class="theme-title mt-2 line-clamp-1 text-xl font-black">
								{item.productName}
							</h2>
							<p class="theme-copy mt-2 line-clamp-2 text-sm">{item.productDescription}</p>
							<div class="mt-3 flex flex-wrap items-center gap-3 text-sm">
								<span class="theme-pill theme-pill-contrast">Quantite: {item.quantity}</span>
								{#if item.sellerEmail}
									<span class="theme-pill">Vendeur: {item.sellerEmail}</span>
								{/if}
								{#if item.promotionName}
									<span class="theme-pill">Promo: {item.promotionName}</span>
								{/if}
							</div>
						</div>

						<div class="min-w-32 text-right">
							<p class="theme-copy text-xs">Ligne</p>
							<p class="theme-title mt-1 text-lg font-black">
								{item.lineTotal.toFixed(2)}
								{data.order.currency}
							</p>
							{#if item.lineDiscountTotal > 0}
								<p class="theme-copy mt-1 text-xs">
									Economie: {item.lineDiscountTotal.toFixed(2)}
									{data.order.currency}
								</p>
							{/if}
						</div>
					</article>
				{/each}
			</div>
		</div>

		<aside class="theme-panel p-6 md:p-8">
			<p class="theme-kicker">Paiement</p>
			<h2 class="theme-title mt-3 text-3xl font-black">Validation via Stripe</h2>
			<p class="theme-copy mt-3">
				La commande a deja ete creee pour figer le prix. Le paiement par carte te redirige vers la
				page Stripe hebergee, puis le webhook Stripe confirme la commande ici.
			</p>

			<div class="summary-grid mt-8">
				<div class="summary-row">
					<span class="theme-copy">Sous-total</span>
					<span class="theme-title font-bold">
						{data.order.subtotal.toFixed(2)}
						{data.order.currency}
					</span>
				</div>
				<div class="summary-row">
					<span class="theme-copy">Remises</span>
					<span class="theme-title font-bold">
						-{data.order.discountTotal.toFixed(2)}
						{data.order.currency}
					</span>
				</div>
				<div class="summary-row">
					<span class="theme-copy">Total</span>
					<span class="theme-price text-3xl font-black">
						{data.order.total.toFixed(2)}
						{data.order.currency}
					</span>
				</div>
			</div>

			{#if showPaymentSuccessBanner}
				<div class="theme-alert theme-alert-success mt-6">
					<p class="theme-kicker">Paiement</p>
					<p class="theme-copy mt-2">
						Paiement confirme. La commande est maintenant en preparation.
					</p>
				</div>
			{/if}

			{#if showPaymentProcessingBanner}
				<div class="theme-alert theme-alert-success mt-6">
					<p class="theme-kicker">Paiement</p>
					<p class="theme-copy mt-2">
						Retour de Stripe detecte. La confirmation finale est en cours; recharge la page si le
						statut ne passe pas encore en preparation.
					</p>
				</div>
			{/if}

			{#if showPaymentCancelledBanner}
				<div class="theme-alert theme-alert-error mt-6">
					<p class="theme-kicker">Paiement</p>
					<p class="theme-copy mt-2">
						Le paiement Stripe a ete annule. La commande reste en attente de paiement.
					</p>
				</div>
			{/if}

			{#if form?.error}
				<div class="theme-alert theme-alert-error mt-6">
					<p class="theme-kicker">Paiement</p>
					<p class="theme-copy mt-2">{form.error}</p>
				</div>
			{/if}

			<div class="mt-8 flex flex-col gap-3">
				{#if canOrderBePaid(data.order.status)}
					<form method="POST" action="?/pay" class="contents">
						<button type="submit" class="theme-button theme-button-primary w-full justify-center">
							Payer avec Stripe
						</button>
					</form>
					<p class="theme-copy text-sm">
						Aucune donnee bancaire n'est saisie dans l'application: Stripe gere la page de paiement.
					</p>
				{:else}
					<div class="theme-alert theme-alert-success">
						<p class="theme-kicker">Etat</p>
						<p class="theme-copy mt-2">
							Cette commande n'attend plus de paiement. Son statut actuel est
							{getOrderStatusLabel(data.order.status).toLowerCase()}.
						</p>
					</div>
				{/if}

				<a
					href={resolve('/catalogue')}
					class="theme-button theme-button-contrast w-full justify-center"
				>
					Continuer mes achats
				</a>
			</div>
		</aside>
	</div>
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

	.status-steps {
		display: grid;
		gap: 0.75rem;
	}

	.status-step {
		border-radius: 1rem;
		border: 1px solid var(--color-border);
		background: rgb(var(--color-white-rgb) / 0.84);
		padding: 0.9rem 1rem;
		font-weight: 700;
		color: var(--color-ink-muted);
	}

	.status-step.step-active {
		border-color: rgb(var(--color-secondary-rgb) / 0.3);
		background: rgb(var(--color-secondary-rgb) / 0.18);
		color: var(--color-primary);
	}
</style>
