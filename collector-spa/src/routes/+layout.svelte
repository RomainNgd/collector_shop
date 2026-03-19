<script lang="ts">
	import { resolve } from '$app/paths';
	import { cartAddPulse, cartCount } from '$lib/stores/cart';
	import { ADMIN_ROLE } from '$lib/types';
	import { onDestroy } from 'svelte';
	import type { Snippet } from 'svelte';
	import type { LayoutData } from './$types';
	import '../app.css';

	let { children, data }: { children: Snippet; data: LayoutData } = $props();

	let pulseActive = $state(false);
	let timeoutId: ReturnType<typeof setTimeout> | undefined;

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

<div class="app-shell min-h-screen text-slate-900">
	<header class="sticky top-0 z-50 w-full border-b border-white/50 bg-white/70 backdrop-blur-xl">
		<div class="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
			<div class="flex items-center gap-8">
				<a href={resolve('/')} class="text-brand-primary text-2xl font-black tracking-tight">
					COLLECTOR<span class="text-slate-900">SHOP</span>
				</a>
				<nav class="hidden gap-6 text-sm font-semibold md:flex">
					<a href={resolve('/')} class="nav-link">Accueil</a>
					<a href={resolve('/')} class="nav-link">Catalogue</a>
					{#if data.user?.role === ADMIN_ROLE}
						<a href={resolve('/administration')} class="nav-link">Administration</a>
					{/if}
				</nav>
			</div>

			<div class="flex items-center gap-4">
				{#if data.user}
					<span class="hidden text-xs font-semibold text-slate-600 sm:inline">
						Connecte: {data.user.role}
					</span>
					<form method="POST" action={resolve('/logout')}>
						<button type="submit" class="auth-link">Deconnexion</button>
					</form>
				{:else}
					<a href={resolve('/login')} class="auth-link">Connexion</a>
				{/if}

				<a href={resolve('/panier')} class="cart-link">
					<span class="sr-only">Panier</span>
					Panier
					<span class="cart-badge" class:pulse={pulseActive}>{$cartCount}</span>
				</a>
			</div>
		</div>
	</header>

	<main class="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
		{@render children()}
	</main>

	<footer class="border-t border-slate-200/80 bg-white/60 py-12 backdrop-blur">
		<div class="mx-auto max-w-7xl px-4 text-center">
			<p class="text-sm text-slate-500">&copy; 2026 Collector Shop. Fabrique avec passion.</p>
		</div>
	</footer>
</div>

<style>
	.app-shell {
		background:
			radial-gradient(circle at 10% -10%, rgba(37, 99, 235, 0.18), transparent 40%),
			radial-gradient(circle at 90% 5%, rgba(245, 158, 11, 0.14), transparent 35%),
			linear-gradient(180deg, #f8fafc 0%, #eff6ff 100%);
	}

	.nav-link {
		transition: color 220ms ease;
	}

	.nav-link:hover {
		color: var(--color-primary);
	}

	.cart-link {
		position: relative;
		border-radius: 999px;
		padding: 0.45rem 0.9rem;
		font-size: 0.84rem;
		font-weight: 700;
		color: #0f172a;
		background: white;
		border: 1px solid rgba(30, 41, 59, 0.12);
		transition: transform 180ms ease;
	}

	.auth-link {
		border-radius: 999px;
		padding: 0.45rem 0.9rem;
		font-size: 0.84rem;
		font-weight: 700;
		color: #0f172a;
		background: white;
		border: 1px solid rgba(30, 41, 59, 0.12);
		transition: transform 180ms ease;
	}

	.auth-link:hover {
		transform: translateY(-1px);
	}

	.cart-link:hover {
		transform: translateY(-1px);
	}

	.cart-badge {
		position: absolute;
		top: -0.25rem;
		right: -0.2rem;
		display: flex;
		height: 1.1rem;
		width: 1.1rem;
		align-items: center;
		justify-content: center;
		border-radius: 999px;
		background: var(--color-primary);
		color: white;
		font-size: 0.62rem;
		font-weight: 800;
	}

	.cart-badge.pulse {
		animation: cart-badge-pulse 520ms cubic-bezier(0.2, 0.8, 0.2, 1);
	}

	@keyframes cart-badge-pulse {
		0% {
			transform: scale(1);
		}
		45% {
			transform: scale(1.48);
		}
		100% {
			transform: scale(1);
		}
	}
</style>
