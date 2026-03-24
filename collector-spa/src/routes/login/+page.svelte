<script lang="ts">
	import { resolve } from '$app/paths';
	import type { ActionData, PageData } from './$types';

	let { form, data }: { form: ActionData; data: PageData } = $props();
</script>

<section class="grid gap-8 lg:grid-cols-[1.05fr_0.9fr] lg:items-center">
	<div class="theme-panel login-showcase p-8 md:p-10">
		<span class="theme-pill">Espace membre</span>
		<p class="theme-kicker mt-6">Connexion</p>
		<h1 class="theme-title mt-4 text-4xl font-black md:text-5xl">Retrouve ton compte</h1>
		<p class="theme-copy mt-4 max-w-xl">
			Connecte-toi pour acceder a ton espace et continuer sur le shop.
		</p>
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
		max-width: 32rem;
	}
</style>
