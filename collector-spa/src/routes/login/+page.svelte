<script lang="ts">
	import { resolve } from '$app/paths';
	import type { ActionData, PageData } from './$types';

	let { form, data }: { form: ActionData; data: PageData } = $props();
</script>

<section class="grid gap-8 lg:grid-cols-[1.05fr_0.9fr] lg:items-center">
	<div class="theme-panel login-showcase p-8 md:p-10">
		<span class="theme-pill theme-pill-contrast">Espace membre</span>
		<p class="theme-kicker showcase-kicker mt-6">Connexion</p>
		<h1 class="showcase-title mt-4 text-4xl font-black md:text-5xl">Retrouve ton compte</h1>
		<p class="showcase-copy mt-4 max-w-xl">
			Connecte-toi pour acceder a ton espace et continuer sur le shop.
		</p>

		<div class="login-benefits mt-8">
			<div class="benefit-row">
				<span class="benefit-dot"></span>
				<p>Ton panier et ton historique de commandes retrouves instantanement.</p>
			</div>
			<div class="benefit-row">
				<span class="benefit-dot"></span>
				<p>Un espace vendeur si tu proposes deja des pieces sur le shop.</p>
			</div>
		</div>
	</div>

	<div class="theme-panel p-8 shadow-sm">
		<p class="theme-kicker">Connexion</p>
		<h2 class="theme-title mt-3 text-3xl font-black">Acces au compte</h2>
		<p class="theme-copy mt-2 text-sm">Connecte-toi pour continuer sur le shop.</p>

		{#if data.registered && !form?.error}
			<div class="theme-alert theme-alert-success mt-6" aria-live="polite">
				<p class="theme-kicker">Compte cree</p>
				<p class="mt-2 text-sm">
					Ton compte est pret. Connecte-toi maintenant pour continuer sur le shop.
				</p>
			</div>
		{/if}

		{#if data.sessionExpired && !form?.error}
			<div class="theme-alert theme-alert-error mt-6" aria-live="polite">
				<p class="theme-kicker">Session expiree</p>
				<p class="mt-2 text-sm">Votre session a expire, veuillez vous reconnecter.</p>
			</div>
		{/if}

		<form method="POST" class="mt-6 space-y-4">
			<div>
				<label for="email" class="theme-label">Email</label>
				<input
					id="email"
					name="email"
					type="email"
					required
					autocomplete="email"
					value={form?.email ?? data.registeredEmail ?? ''}
					class="theme-input"
				/>
			</div>

			<div>
				<label for="password" class="theme-label">Mot de passe</label>
				<input
					id="password"
					name="password"
					type="password"
					required
					autocomplete="current-password"
					class="theme-input"
				/>
			</div>

			{#if form?.error}
				<div class="theme-alert theme-alert-error">
					<p class="theme-kicker">Erreur</p>
					<p class="mt-2 text-sm">{form.error}</p>
				</div>
			{/if}

			<button type="submit" class="theme-button theme-button-primary w-full justify-center">
				Se connecter
			</button>

			<a
				href={resolve('/auth/register')}
				class="theme-button theme-button-secondary w-full justify-center"
			>
				Creer un compte
			</a>
		</form>
	</div>
</section>

<style>
	.login-showcase {
		position: relative;
		overflow: hidden;
		max-width: 32rem;
		background: var(--gradient-primary);
		box-shadow: var(--shadow-strong);
	}

	.login-showcase::after {
		content: '';
		position: absolute;
		top: -4rem;
		right: -4rem;
		height: 14rem;
		width: 14rem;
		border-radius: 42% 58% 65% 35% / 45% 45% 55% 55%;
		background: rgb(var(--color-secondary-rgb) / 0.22);
		filter: blur(2px);
	}

	.showcase-kicker {
		color: rgb(var(--color-secondary-rgb));
	}

	.showcase-title {
		color: var(--color-white);
		letter-spacing: -0.04em;
	}

	.showcase-copy {
		color: rgb(var(--color-white-rgb) / 0.78);
	}

	.login-benefits {
		display: flex;
		flex-direction: column;
		gap: 0.85rem;
	}

	.benefit-row {
		position: relative;
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		font-size: 0.9rem;
		color: rgb(var(--color-white-rgb) / 0.86);
	}

	.benefit-dot {
		margin-top: 0.4rem;
		height: 0.4rem;
		width: 0.4rem;
		flex-shrink: 0;
		border-radius: 999px;
		background: var(--color-secondary);
	}
</style>
