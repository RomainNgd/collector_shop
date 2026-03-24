<script lang="ts">
	import { browser } from '$app/environment';
	import { enhance } from '$app/forms';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import type { SubmitFunction } from '@sveltejs/kit';
	import type { ActionData } from './$types';

	let { form }: { form: ActionData } = $props();

	let password = $state('');
	let confirmPassword = $state('');
	let clientError = $state<string | null>(null);
	let isSubmitting = $state(false);
	let redirectScheduled = $state(false);

	const passwordTooShort = $derived(password.length > 0 && password.length < 8);
	const passwordsMismatch = $derived(confirmPassword.length > 0 && password !== confirmPassword);
	const activeError = $derived(clientError ?? form?.error ?? null);

	const enhanceRegister: SubmitFunction = ({ cancel }) => {
		clientError = null;

		if (password.length < 8) {
			clientError = 'Le mot de passe doit contenir au moins 8 caracteres.';
			cancel();
			return;
		}

		if (password !== confirmPassword) {
			clientError = 'Les mots de passe ne correspondent pas.';
			cancel();
			return;
		}

		isSubmitting = true;

		return async ({ result, update }) => {
			if (result.type === 'error') {
				clientError = 'Une erreur inattendue est survenue. Reessaie dans un instant.';
				isSubmitting = false;
				return;
			}

			await update({ reset: false });
			isSubmitting = false;
		};
	};

	$effect(() => {
		if (!browser || !form?.success) {
			redirectScheduled = false;
			return;
		}

		if (redirectScheduled) {
			return;
		}

		redirectScheduled = true;

		const timeoutId = window.setTimeout(() => {
			goto(resolve('/login'));
		}, 1600);

		return () => window.clearTimeout(timeoutId);
	});
</script>

<section class="grid gap-8 lg:grid-cols-[1.05fr_0.95fr] lg:items-center">
	<div class="theme-panel register-showcase p-8 md:p-10">
		<span class="theme-pill">Nouveau membre</span>
		<p class="theme-kicker mt-6">Inscription</p>
		<h1 class="theme-title mt-4 text-4xl font-black md:text-5xl">Cree ton espace collection</h1>
		<p class="theme-copy mt-4 max-w-xl">
			Prepare ton compte en quelques secondes et retrouve ensuite ton panier, ton espace et tes
			prochaines trouvailles.
		</p>

		<div class="register-benefits mt-8">
			<div class="benefit-card">
				<p class="theme-kicker">Simple</p>
				<p class="mt-3 text-sm">Une adresse email, un mot de passe solide, et c'est parti.</p>
			</div>
			<div class="benefit-card">
				<p class="theme-kicker">Serein</p>
				<p class="mt-3 text-sm">La confirmation du mot de passe evite les erreurs de saisie.</p>
			</div>
		</div>
	</div>

	<div class="theme-panel p-8 shadow-sm">
		<p class="theme-kicker">Creation de compte</p>
		<h2 class="theme-title mt-3 text-3xl font-black">Inscription pro et rapide</h2>
		<p class="theme-copy mt-2 text-sm">
			Utilise au moins 8 caracteres pour ton mot de passe et valide la double saisie.
		</p>

		{#if form?.success}
			<div class="theme-alert theme-alert-success mt-6" aria-live="polite">
				<p class="theme-kicker">Compte cree</p>
				<p class="mt-2 text-sm">{form.message}</p>
				<a
					href={resolve('/login')}
					class="theme-button theme-button-secondary mt-4 w-full justify-center"
				>
					Aller a la connexion
				</a>
			</div>
		{:else}
			<form method="POST" class="mt-6 space-y-4" use:enhance={enhanceRegister}>
				<div>
					<label for="email" class="theme-label">Email</label>
					<input
						id="email"
						name="email"
						type="email"
						required
						autocomplete="email"
						value={form?.email ?? ''}
						class="theme-input"
					/>
				</div>

				<div class="grid gap-4 sm:grid-cols-2">
					<div>
						<label for="password" class="theme-label">Mot de passe</label>
						<input
							id="password"
							name="password"
							type="password"
							required
							minlength="8"
							autocomplete="new-password"
							bind:value={password}
							aria-invalid={passwordTooShort}
							class="theme-input"
						/>
					</div>

					<div>
						<label for="confirmPassword" class="theme-label">Confirmer le mot de passe</label>
						<input
							id="confirmPassword"
							name="confirmPassword"
							type="password"
							required
							autocomplete="new-password"
							bind:value={confirmPassword}
							aria-invalid={passwordsMismatch}
							class="theme-input"
						/>
					</div>
				</div>

				<div class="register-checks">
					<span class:check-active={!passwordTooShort}>8+ caracteres</span>
					<span class:check-active={confirmPassword.length > 0 && !passwordsMismatch}>
						Confirmation identique
					</span>
				</div>

				{#if activeError}
					<div class="theme-alert theme-alert-error" aria-live="polite">
						<p class="theme-kicker">Erreur</p>
						<p class="mt-2 text-sm">{activeError}</p>
					</div>
				{/if}

				<button
					type="submit"
					disabled={isSubmitting}
					class="theme-button theme-button-primary w-full justify-center"
				>
					{#if isSubmitting}
						Inscription en cours...
					{:else}
						S'inscrire
					{/if}
				</button>
			</form>
		{/if}

		<p class="theme-copy mt-6 text-sm">
			Deja membre ?
			<a href={resolve('/login')} class="inline-link">Retour a la connexion</a>
		</p>
	</div>
</section>

<style>
	.register-showcase {
		max-width: 34rem;
	}

	.register-benefits {
		display: grid;
		gap: 1rem;
		grid-template-columns: repeat(2, minmax(0, 1fr));
	}

	.benefit-card {
		border-radius: 1.35rem;
		border: 1px solid rgb(var(--color-primary-rgb) / 0.08);
		background: rgb(var(--color-white-rgb) / 0.6);
		padding: 1.25rem;
		color: var(--color-ink);
	}

	.register-checks {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
	}

	.register-checks span {
		border-radius: 999px;
		border: 1px solid var(--color-border);
		background: rgb(var(--color-white-rgb) / 0.85);
		padding: 0.55rem 0.85rem;
		font-size: 0.78rem;
		font-weight: 700;
		color: var(--color-ink-muted);
		transition:
			border-color var(--transition-standard),
			background-color var(--transition-standard),
			color var(--transition-standard);
	}

	.register-checks .check-active {
		border-color: rgb(var(--color-primary-rgb) / 0.18);
		background: rgb(var(--color-secondary-rgb) / 0.18);
		color: var(--color-primary);
	}

	.inline-link {
		font-weight: 700;
		color: var(--color-primary);
	}

	.inline-link:hover {
		text-decoration: underline;
		text-underline-offset: 0.18rem;
	}

	@media (max-width: 640px) {
		.register-benefits {
			grid-template-columns: 1fr;
		}
	}
</style>
