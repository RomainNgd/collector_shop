<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { cartAddPulse, cartCount } from '$lib/stores/cart';
	import { ADMIN_ROLE } from '$lib/types';
	import { onDestroy } from 'svelte';
	import type { Snippet } from 'svelte';
	import type { LayoutData } from './$types';
	import '../app.css';

	let { children, data }: { children: Snippet; data: LayoutData } = $props();

	let pulseActive = $state(false);
	let timeoutId: ReturnType<typeof setTimeout> | undefined;

	function isActive(path: string): boolean {
		const current = page.url.pathname;
		if (path === '/') {
			return current === '/';
		}
		return current === path || current.startsWith(`${path}/`);
	}

	$effect(() => {
		if ($cartAddPulse === 0) {
			return;
		}

		pulseActive = true;
		if (timeoutId) {
			clearTimeout(timeoutId);
		}
		timeoutId = setTimeout(() => {
			pulseActive = false;
		}, 520);
	});

	onDestroy(() => {
		if (timeoutId) {
			clearTimeout(timeoutId);
		}
	});
</script>

<div class="app-shell min-h-screen">
	<header class="site-header">
		<div
			class="mx-auto flex max-w-7xl items-center justify-between gap-6 px-4 py-4 sm:px-6 lg:px-8"
		>
			<div class="flex items-center gap-8">
				<a href={resolve('/')} class="brand-link">
					<span class="brand-mark" aria-hidden="true">C</span>
					<span class="brand-lockup">
						<span class="brand-kicker">Collector</span>
						<span class="brand-name">Shop</span>
					</span>
				</a>

				<nav class="hidden items-center gap-3 md:flex">
					<a href={resolve('/')} class="nav-link" class:active={isActive('/')}>Accueil</a>
					<a href={resolve('/catalogue')} class="nav-link" class:active={isActive('/catalogue')}
						>Catalogue</a
					>
					{#if data.user}
						{#if data.user.role !== ADMIN_ROLE}
							<a href={resolve('/vendre')} class="nav-link" class:active={isActive('/vendre')}
								>Vendre</a
							>
							<a
								href={resolve('/mes-produits')}
								class="nav-link"
								class:active={isActive('/mes-produits')}>Dashboard</a
							>
						{/if}
						<a
							href={resolve('/mes-commandes')}
							class="nav-link"
							class:active={isActive('/mes-commandes')}>Mes commandes</a
						>
					{/if}
					{#if data.user?.role === ADMIN_ROLE}
						<a
							href={resolve('/administration')}
							class="nav-link"
							class:active={isActive('/administration')}>Administration</a
						>
					{/if}
				</nav>
			</div>

			<div class="flex items-center gap-2 sm:gap-3">
				{#if data.user}
					<a
						href={resolve('/profil')}
						class="nav-link hidden md:inline-flex"
						class:active={isActive('/profil')}>Profil</a
					>
					<form method="POST" action={resolve('/logout')}>
						<button type="submit" class="header-action header-action-secondary compact-action">
							<span class="hidden sm:inline">Deconnexion</span>
							<span class="sm:hidden" aria-hidden="true">⏻</span>
						</button>
					</form>
				{:else}
					<a href={resolve('/auth/register')} class="header-action header-action-primary compact-action">
						<span class="hidden sm:inline">Inscription</span>
						<span class="sm:hidden">S'inscrire</span>
					</a>
					<a href={resolve('/login')} class="header-action header-action-secondary compact-action">
						<span class="hidden sm:inline">Connexion</span>
						<span class="sm:hidden">Connexion</span>
					</a>
				{/if}

				<a href={resolve('/panier')} class="header-action header-cart compact-action">
					<span class="hidden sm:inline">Panier</span>
					<span class="sm:hidden" aria-hidden="true">🛒</span>
					<span class="sr-only sm:hidden">Panier</span>
					<span class="cart-badge" class:pulse={pulseActive}>{$cartCount}</span>
				</a>
			</div>
		</div>
	</header>

	<main class="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
		<div class="page-reveal">
			{@render children()}
		</div>
	</main>

	<footer class="site-footer">
		<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
			<div class="footer-card">
				<div>
					<p class="theme-kicker">Collector Shop</p>
					<p class="theme-copy mt-3 max-w-2xl text-sm">
						Des pieces choisies pour enrichir chaque collection.
					</p>
				</div>
				<p class="theme-copy text-sm">&copy; 2026 Collector Shop. Fabrique avec passion.</p>
			</div>
		</div>
	</footer>
</div>

<style>
	.app-shell {
		color: var(--color-ink);
	}

	.site-header {
		position: sticky;
		top: 0;
		z-index: 50;
		border-bottom: 1px solid rgb(var(--color-primary-rgb) / 0.08);
		background: linear-gradient(
			180deg,
			rgb(var(--color-white-rgb) / 0.92),
			rgb(var(--color-white-rgb) / 0.72)
		);
		backdrop-filter: blur(20px);
		box-shadow: var(--shadow-soft);
	}

	.brand-link {
		display: inline-flex;
		align-items: center;
		gap: 0.9rem;
	}

	.brand-link:hover .brand-mark {
		transform: rotate(-4deg) scale(1.04);
	}

	.brand-mark {
		display: grid;
		place-items: center;
		height: 2.9rem;
		width: 2.9rem;
		border-radius: 1rem;
		background: var(--gradient-primary);
		color: var(--color-white);
		font-size: 1.25rem;
		font-weight: 900;
		box-shadow:
			var(--shadow-button),
			inset 0 1px 0 rgb(255 255 255 / 0.16);
		transition: transform var(--transition-standard);
	}

	.brand-lockup {
		display: flex;
		flex-direction: column;
		gap: 0.15rem;
	}

	.brand-kicker {
		font-size: 0.72rem;
		font-weight: 800;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		color: var(--color-primary);
	}

	.brand-name {
		font-size: 1.25rem;
		font-weight: 900;
		letter-spacing: -0.05em;
		color: var(--color-black);
	}

	.nav-link {
		position: relative;
		border-radius: 999px;
		padding: 0.55rem 0.9rem;
		font-size: 0.88rem;
		font-weight: 700;
		color: var(--color-primary);
		transition:
			background-color var(--transition-standard),
			color var(--transition-standard);
	}

	.nav-link:hover {
		background: rgb(var(--color-secondary-rgb) / 0.12);
		color: var(--color-black);
	}

	.nav-link::after {
		content: '';
		position: absolute;
		left: 0.9rem;
		right: 0.9rem;
		bottom: 0.15rem;
		height: 2px;
		border-radius: 999px;
		background: var(--color-secondary);
		transform: scaleX(0);
		transform-origin: left;
		transition: transform var(--transition-standard);
	}

	.nav-link.active {
		color: var(--color-black);
	}

	.nav-link.active::after {
		transform: scaleX(1);
	}

	.header-action {
		position: relative;
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		border-radius: 999px;
		padding: 0.65rem 1rem;
		font-size: 0.84rem;
		font-weight: 800;
		transition:
			transform 160ms ease,
			border-color var(--transition-standard),
			background-color var(--transition-standard),
			filter var(--transition-standard);
	}

	@media (max-width: 639px) {
		.compact-action {
			padding: 0.55rem 0.75rem;
			font-size: 0.78rem;
		}

		.header-cart.compact-action {
			padding: 0.6rem 0.85rem;
			margin-right: 0.35rem;
		}
	}

	@media (max-width: 420px) {
		.brand-lockup {
			display: none;
		}
	}

	.header-action-secondary {
		border: 1px solid var(--color-border);
		background: rgb(var(--color-white-rgb) / 0.9);
		color: var(--color-primary);
		box-shadow: inset 0 1px 0 rgb(var(--color-white-rgb) / 0.6);
	}

	.header-action-secondary:hover {
		transform: translateY(-1px);
		border-color: var(--color-border-strong);
		background: rgb(var(--color-secondary-rgb) / 0.1);
	}

	.header-action-primary {
		background: var(--gradient-primary);
		color: var(--color-white);
		box-shadow: var(--shadow-button);
	}

	.header-action-primary:hover {
		transform: translateY(-1px);
		filter: saturate(1.03) brightness(1.02);
	}

	.header-cart {
		padding-right: 1.7rem;
		background: var(--gradient-primary);
		color: var(--color-white);
		box-shadow: var(--shadow-button);
	}

	.header-cart:hover {
		transform: translateY(-1px);
		filter: saturate(1.03) brightness(1.02);
	}

	.account-pill {
		background: rgb(var(--color-secondary-rgb) / 0.18);
		border-color: rgb(var(--color-primary-rgb) / 0.18);
	}

	.cart-badge {
		position: absolute;
		top: -0.35rem;
		right: -0.25rem;
		display: flex;
		height: 1.35rem;
		width: 1.35rem;
		align-items: center;
		justify-content: center;
		border-radius: 999px;
		background: var(--color-secondary);
		color: var(--color-primary);
		font-size: 0.68rem;
		font-weight: 900;
		box-shadow: 0 0 0 4px rgb(var(--color-white-rgb) / 0.86);
	}

	.cart-badge.pulse {
		animation: pulse-soft 520ms cubic-bezier(0.2, 0.8, 0.2, 1);
	}

	.cart-badge.pulse::after {
		content: '';
		position: absolute;
		inset: 0;
		border-radius: 999px;
		background: rgb(var(--color-secondary-rgb) / 0.55);
		animation: ping-ring 620ms cubic-bezier(0.2, 0.8, 0.2, 1);
	}

	@keyframes ping-ring {
		0% {
			box-shadow: 0 0 0 0 rgb(var(--color-secondary-rgb) / 0.45);
			opacity: 1;
		}
		100% {
			box-shadow: 0 0 0 10px rgb(var(--color-secondary-rgb) / 0);
			opacity: 0;
		}
	}

	.site-footer {
		padding: 0 0 3rem;
	}

	.footer-card {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1.5rem;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-panel);
		padding: 1.5rem 1.75rem;
		background: rgb(var(--color-white-rgb) / 0.84);
		box-shadow: var(--shadow-soft);
		backdrop-filter: blur(18px);
	}

	@media (max-width: 768px) {
		.footer-card {
			flex-direction: column;
			align-items: flex-start;
		}
	}
</style>
