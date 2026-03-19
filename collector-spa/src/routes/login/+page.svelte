<script lang="ts">
	import type { ActionData } from './$types';

	let { form }: { form: ActionData } = $props();
</script>

<section class="mx-auto max-w-md rounded-3xl border border-white/70 bg-white/90 p-8 shadow-sm">
	<h1 class="text-3xl font-black text-slate-900">Connexion</h1>
	<p class="mt-2 text-sm text-slate-500">
		Connecte-toi pour acceder aux fonctions d'administration.
	</p>

	<form method="POST" class="mt-6 space-y-4">
		<div>
			<label for="email" class="mb-1 block text-sm font-semibold text-slate-700">Email</label>
			<input
				id="email"
				name="email"
				type="email"
				required
				value={form?.email ?? ''}
				class="w-full rounded-xl border border-slate-300 px-3 py-2 focus:border-blue-500 focus:outline-none"
			/>
		</div>

		<div>
			<label for="password" class="mb-1 block text-sm font-semibold text-slate-700"
				>Mot de passe</label
			>
			<input
				id="password"
				name="password"
				type="password"
				required
				class="w-full rounded-xl border border-slate-300 px-3 py-2 focus:border-blue-500 focus:outline-none"
			/>
		</div>

		{#if form?.error}
			<p class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
				{form.error}
			</p>
		{/if}

		<button
			type="submit"
			class="w-full rounded-xl bg-blue-600 px-4 py-2.5 font-bold text-white transition hover:bg-blue-700"
		>
			Se connecter
		</button>
	</form>
</section>
